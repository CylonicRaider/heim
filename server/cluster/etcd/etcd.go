package etcd

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"euphoria.leet.nu/lib/scope"
	"github.com/prometheus/client_golang/prometheus"
	etcdpb "go.etcd.io/etcd/api/v3/etcdserverpb"
	etcd "go.etcd.io/etcd/client/v3"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/security"
)

var ErrParted = errors.New("parted from cluster")

var (
	selfAnnouncements = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "self_announcements",
		Subsystem: "peer",
		Help:      "Count of self-announcements to the cluster by this backend.",
	})

	peerEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "events",
		Subsystem: "peer",
		Help:      "Count of cluster peer events observed by this backend.",
	})

	peerLiveCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "live_count",
		Subsystem: "peer",
		Help:      "Count of peers currently live (including self).",
	})

	peerWatchErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "watch_errors",
		Subsystem: "peer",
		Help:      "Count of errors encountered while watching for peer events.",
	})
)

const (
	maxConsecutiveWatchFailures = 5
	initialWatchBackoff         = 2 * time.Second
)

func init() {
	prometheus.MustRegister(selfAnnouncements)
	prometheus.MustRegister(peerEvents)
	prometheus.MustRegister(peerLiveCount)
	prometheus.MustRegister(peerWatchErrors)
}

type etcdEventType int

const (
	eetError etcdEventType = iota
	eetPut
	eetDelete
)

type etcdNode struct {
	Key              string
	Value            string
	CreatedRevision  int64
	ModifiedRevision int64
}

type etcdEvent struct {
	Type     etcdEventType
	etcdNode       // Zero if type is eetError.
	Error    error // Only present if type is eetError.
}

// *insert invective targeting whoever wrote the etcd client library*
func extractValueGet(key string, resp *etcd.GetResponse) (string, error) {
	return extractValueRange(key, (*etcdpb.RangeResponse)(resp), true)
}

// *insert more invective targeting whoever wrote the etcd client library*
func extractValueTxn(key string, resp *etcd.TxnResponse) (string, error) {
	// This assumes a transaction that contains a single get operation (on the path that was taken).
	if len(resp.Responses) != 1 {
		return "", fmt.Errorf("etcd broken: expected one txn response, got %d", len(resp.Responses))
	}
	return extractValueRange(key, resp.Responses[0].GetResponseRange(), false)
}

// *insert even more invective targeting whoever wrote the etcd client library*
func extractValueRange(key string, resp *etcdpb.RangeResponse, allowEmpty bool) (string, error) {
	if resp.Count == 0 && allowEmpty {
		if len(resp.Kvs) != 0 {
			return "", fmt.Errorf("etcd broken: expected no get kvs, got %d", len(resp.Kvs))
		}
		return "", cluster.ErrNotFound
	}

	if resp.Count != 1 {
		return "", fmt.Errorf("etcd broken: expected get count 1, got %d", resp.Count)
	}
	if len(resp.Kvs) != 1 {
		return "", fmt.Errorf("etcd broken: expected one get kv, got %d", len(resp.Kvs))
	}
	kv := resp.Kvs[0]
	if string(kv.Key) != key {
		return "", fmt.Errorf("etcd broken: expected get key %#v, got %#v", key, string(kv.Key))
	}
	return string(kv.Value), nil
}

// *no invective here because this is actually almost ok*
func extractValueList(prefix string, resp *etcd.GetResponse, recursive bool) (map[string]etcdNode, error) {
	if int64(len(resp.Kvs)) != resp.Count {
		return nil, fmt.Errorf("etcd broken: response count is %d but %d entries returned",
			resp.Count, len(resp.Kvs))
	}
	result := map[string]etcdNode{}
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		if !strings.HasPrefix(key, prefix) {
			return nil, fmt.Errorf("etcd broken: query for prefix %#v returned out-of-bailiwick key %#v",
				prefix, key)
		}
		trimmedKey := key[len(prefix):]
		if !recursive && strings.ContainsRune(trimmedKey, '/') {
			continue
		}
		result[trimmedKey] = etcdNode{
			Key:              key,
			Value:            string(kv.Value),
			CreatedRevision:  kv.CreateRevision,
			ModifiedRevision: kv.ModRevision,
		}
	}
	return result, nil
}

// *insert yet more invective targeting whoever wrote the etcd client library*
func extractWatchEvents(rawEvents etcd.WatchChan) <-chan etcdEvent {
	output := make(chan etcdEvent)

	go func() {
		defer close(output)

		for wev := range rawEvents {
			if wev.Err() != nil {
				output <- etcdEvent{
					Type:  eetError,
					Error: fmt.Errorf("watch error: %s", wev.Err()),
				}
			}

			for _, ev := range wev.Events {
				oev := etcdEvent{
					etcdNode: etcdNode{
						Key:              string(ev.Kv.Key),
						Value:            string(ev.Kv.Value),
						CreatedRevision:  ev.Kv.CreateRevision,
						ModifiedRevision: ev.Kv.ModRevision,
					},
				}

				switch {
				case ev.Type == etcd.EventTypePut:
					oev.Type = eetPut
				case ev.Type == etcd.EventTypeDelete:
					oev.Type = eetDelete
				default:
					output <- etcdEvent{
						Type:  eetError,
						Error: fmt.Errorf("got unexpected watch event type %d", ev.Type),
					}
					continue
				}

				output <- oev
			}
		}
	}()

	return output
}

func EtcdCluster(ctx scope.Context, root, addr string, desc *cluster.PeerDesc) (cluster.Cluster, error) {
	logging.Logger(ctx).Printf("connecting to %#v", addr)
	client, err := etcd.NewFromURL(addr)
	if err != nil {
		return nil, fmt.Errorf("cluster init: create client: %s", err)
	}
	e := &etcdCluster{
		root:  strings.TrimRight(root, "/") + "/",
		c:     client,
		ch:    make(chan cluster.PeerEvent),
		peers: map[string]cluster.PeerDesc{},
		ctx:   ctx,
		wctx:  ctx.Fork(),
	}
	rev, err := e.init(desc)
	if err != nil {
		return nil, err
	}
	go e.background(rev)
	return e, nil
}

type etcdCluster struct {
	m     sync.RWMutex
	c     *etcd.Client
	root  string
	me    string
	lease etcd.LeaseID
	ch    chan cluster.PeerEvent
	peers map[string]cluster.PeerDesc
	ctx   scope.Context
	wctx  scope.Context
}

func (e *etcdCluster) key(format string, args ...interface{}) string {
	return e.root + strings.Trim(fmt.Sprintf(format, args...), "/")
}

func (e *etcdCluster) init(desc *cluster.PeerDesc) (int64, error) {
	if err := e.c.Sync(e.ctx); err != nil {
		return 0, fmt.Errorf("cluster error: failed to sync with %s: %s", e.c.Endpoints(), err)
	}

	resp, err := e.c.Grant(e.ctx, int64(cluster.TTL/time.Second))
	if err != nil {
		return 0, fmt.Errorf("cluster error: acquire lease: %s", err)
	}
	if resp.Error != "" {
		return 0, fmt.Errorf("cluster error: acquire lease: why is there another error field: %s", resp.Error)
	}
	if resp.ID == etcd.NoLease {
		return 0, fmt.Errorf("cluster error: acquire lease: returned no lease")
	}
	e.lease = resp.ID

	if desc != nil {
		err = e.Update(desc)
		if err != nil {
			return 0, err
		}
	}

	peerNodes, err := e.getDirEx("/peers")
	if err != nil {
		return 0, err
	}

	latestRevision := int64(0)
	for _, node := range peerNodes {
		var desc cluster.PeerDesc
		if err := json.Unmarshal([]byte(node.Value), &desc); err != nil {
			return 0, fmt.Errorf("cluster error: init: bad node %s: %s", node.Key, err)
		}
		e.peers[desc.ID] = desc
		latestRevision = max(latestRevision, node.ModifiedRevision)
	}
	if latestRevision > 0 {
		latestRevision++
	}

	return latestRevision, nil
}

func (e *etcdCluster) GetDir(key string) (map[string]string, error) {
	nodes, err := e.getDirEx(key)
	if err != nil {
		return nil, fmt.Errorf("cluster get of dir %s: %s", e.key("%s", key), err)
	}
	result := map[string]string{}
	for key, node := range nodes {
		result[key] = node.Value
	}
	return result, nil
}

func (e *etcdCluster) getDirEx(key string) (map[string]etcdNode, error) {
	fullPrefix := e.key("%s", key) + "/"
	resp, err := e.c.Get(e.ctx, fullPrefix, etcd.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("cluster get of dir %s: %s", fullPrefix[:len(fullPrefix)-1], err)
	}
	return extractValueList(fullPrefix, resp, false)
}

func (e *etcdCluster) GetValue(key string) (string, error) {
	fullKey := e.key("%s", key)
	resp, err := e.c.Get(e.ctx, fullKey)
	if err != nil {
		return "", fmt.Errorf("cluster get of %s: %s", fullKey, err)
	}
	result, err := extractValueGet(fullKey, resp)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (e *etcdCluster) SetValue(key, value string) error {
	fullKey := e.key("%s", key)
	_, err := e.c.Put(e.ctx, fullKey, value)
	if err != nil {
		return fmt.Errorf("cluster put to %s: %s", fullKey, err)
	}
	return nil
}

func (e *etcdCluster) Peers() []cluster.PeerDesc {
	e.m.RLock()
	defer e.m.RUnlock()
	peers := make(cluster.PeerList, 0, len(e.peers))
	for _, desc := range e.peers {
		peers = append(peers, desc)
	}
	sort.Sort(peers)
	return peers
}

func (e *etcdCluster) Update(desc *cluster.PeerDesc) error {
	valueBytes, err := json.Marshal(desc)
	if err != nil {
		return fmt.Errorf("cluster: serialize peer desc: %s", err)
	}

	e.m.Lock()
	e.me = desc.ID
	e.peers[desc.ID] = *desc
	e.m.Unlock()

	meKey := e.key("/peers/%s", e.me)
	logging.Logger(e.ctx).Printf("writing %s to %s", string(valueBytes), meKey)
	_, err = e.c.Put(e.ctx, meKey, string(valueBytes), etcd.WithLease(e.lease))
	if err != nil {
		return fmt.Errorf("cluster put to %s: %s", meKey, err)
	}

	selfAnnouncements.Inc()

	return nil
}

func (e *etcdCluster) Part() {
	e.wctx.Terminate(ErrParted)
	e.m.Lock()
	lease := e.lease
	e.lease = 0
	e.m.Unlock()
	if lease != etcd.NoLease {
		// etcd should do this on its own, but doing it here ensures that it is done by the time the next
		// test starts
		_, err := e.c.Revoke(e.ctx, lease)
		if err != nil {
			logging.Logger(e.ctx).Printf("cluster part: revoke lease error: %s", err)
		}
	}
	err := e.c.Close()
	if err != nil {
		logging.Logger(e.ctx).Printf("cluster part failed: %s", err)
	}
}

func (e *etcdCluster) Watch() <-chan cluster.PeerEvent { return e.ch }

func (e *etcdCluster) background(watchRev int64) {
	defer close(e.ch)

	wctx := etcd.WithRequireLeader(e.wctx)

	keepAlives, err := e.c.KeepAlive(wctx, e.lease)
	if err != nil {
		logging.Logger(e.ctx).Printf("peer watch: launch lease keepalive: %s", err)
		return
	}

	rawWatch := e.c.Watch(wctx, e.key("/peers")+"/", etcd.WithPrefix(), etcd.WithRev(watchRev))
	watch := extractWatchEvents(rawWatch)

	seenMe := false

	for {
		var ev etcdEvent
		var ok bool
		select {
		case <-e.wctx.Done():
			logging.Logger(e.ctx).Printf("peer watch: context done, exiting")
			return
		case _, ok = <-keepAlives:
			if !ok {
				err = fmt.Errorf("keep-alive channel closed unexpectedly")
				e.wctx.Terminate(err)
				logging.Logger(e.ctx).Printf("peer watch: %s", err)
				return
			}
			continue
		case ev, ok = <-watch:
			if !ok {
				err = fmt.Errorf("watch channel closed unexpectedly")
				e.wctx.Terminate(err)
				logging.Logger(e.ctx).Printf("peer watch: %s", err)
				return
			}
		}

		if ev.Type == eetError {
			logging.Logger(e.ctx).Printf("peer watch: error: %s", ev.Error)
			peerWatchErrors.Inc()
			continue
		}

		peerID := strings.TrimPrefix(ev.Key, e.key("/peers")+"/")

		switch ev.Type {
		case eetPut:
			var desc cluster.PeerDesc
			if err := json.Unmarshal([]byte(ev.Value), &desc); err != nil {
				logging.Logger(e.ctx).Printf("peer watch: decode announcement: %s", err)
				peerWatchErrors.Inc()
				continue
			}
			e.m.Lock()
			prev, updated := e.peers[desc.ID]
			e.peers[desc.ID] = desc
			if desc.ID == e.me && !seenMe {
				updated = false
				seenMe = true
			}
			e.m.Unlock()
			if updated {
				if prev.Era != desc.Era {
					logging.Logger(e.ctx).Printf("peer watch: update %s", desc.ID)
				}
				e.ch <- &cluster.PeerAliveEvent{desc}
			} else {
				logging.Logger(e.ctx).Printf("peer watch: create %s", desc.ID)
				e.ch <- &cluster.PeerJoinedEvent{desc}
			}
			peerEvents.Inc()

		case eetDelete:
			logging.Logger(e.ctx).Printf("peer watch: delete %s", peerID)
			e.m.Lock()
			delete(e.peers, peerID)
			e.m.Unlock()
			e.ch <- &cluster.PeerLostEvent{cluster.PeerDesc{ID: peerID}}
			peerEvents.Inc()
		}

		peerLiveCount.Set(float64(len(e.peers)))
	}
}

func (e *etcdCluster) getWithDefault(fullKey string, setter func() (string, error)) (string, error) {
	value, err := setter()
	if err != nil {
		return "", err
	}

	resp, err := e.c.Txn(e.ctx).If(
		etcd.Compare(etcd.Version(fullKey), ">", 0),
	).Then(
		etcd.OpGet(fullKey),
	).Else(
		etcd.OpPut(fullKey, value),
	).Commit()
	if err != nil {
		return "", err
	}

	if resp.Succeeded {
		return extractValueTxn(fullKey, resp)
	} else {
		// Of course, the value could have been replaced by the time we get here, but an attempt was made.
		return value, nil
	}
}

func (e *etcdCluster) GetSecret(kms security.KMS, name string, bytes int) ([]byte, error) {
	fullKey := e.key("/secrets/%s", name)
	resp, err := e.c.Get(e.ctx, fullKey)
	if err != nil {
		return nil, err
	}
	result, err := extractValueGet(fullKey, resp)
	if err == cluster.ErrNotFound {
		return e.setSecret(kms, name, bytes)
	} else if err != nil {
		return nil, err
	}
	return e.decodeSecret(result, bytes)
}

func (e *etcdCluster) setSecret(kms security.KMS, name string, bytes int) ([]byte, error) {
	secret, err := kms.GenerateNonce(bytes)
	if err != nil {
		return nil, err
	}
	if len(secret) != bytes {
		return nil, fmt.Errorf("kms broken: requested %d bytes, got %d", bytes, len(secret))
	}
	hexSecret := hex.EncodeToString(secret)

	result, err := e.getWithDefault(e.key("/secrets/%s", name), func() (string, error) {
		return hexSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if result == hexSecret {
		return secret, nil
	}
	return e.decodeSecret(result, bytes)
}

func (e *etcdCluster) decodeSecret(value string, bytes int) ([]byte, error) {
	secret, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(secret) != bytes {
		return nil, fmt.Errorf("secret inconsistent: expected %d bytes, got %d", bytes, len(secret))
	}
	return secret, nil
}

func (e *etcdCluster) GetValueWithDefault(key string, setter func() (string, error)) (string, error) {
	fullKey := e.key("%s", key)
	resp, err := e.c.Get(e.ctx, fullKey)
	if err != nil {
		return "", fmt.Errorf("cluster get of %s: %s", fullKey, err)
	}
	result, err := extractValueGet(fullKey, resp)
	if err == cluster.ErrNotFound {
		return e.getWithDefault(fullKey, setter)
	} else if err != nil {
		return "", fmt.Errorf("cluster get of %s: %s", fullKey, err)
	}
	return result, nil
}
