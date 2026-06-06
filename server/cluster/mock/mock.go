package mock

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/proto/security"
)

func NewMockClusterGroup() *mockClusterGroup {
	return &mockClusterGroup{
		data:  map[string]string{},
		peers: map[string]*mockCluster{},
	}
}

type mockClusterGroup struct {
	sync.Mutex
	data  map[string]string
	peers map[string]*mockCluster
}

func (g *mockClusterGroup) getDir(key string) (map[string]string, error) {
	g.Lock()
	defer g.Unlock()

	key = strings.TrimRight(key, "/") + "/"

	result := map[string]string{}
	for k, v := range g.data {
		if !strings.HasPrefix(k, key) {
			continue
		}
		rk := k[len(key):]
		if strings.ContainsRune(rk, '/') {
			continue
		}
		result[rk] = v
	}
	return result, nil
}

func (g *mockClusterGroup) getSet(key string, setter func() (string, error), override bool) (string, error) {
	g.Lock()
	defer g.Unlock()

	if !override {
		if value, ok := g.data[key]; ok {
			return value, nil
		}
	}

	value, err := setter()
	if err != nil {
		return "", err
	}

	if g.data == nil {
		g.data = map[string]string{key: value}
	} else {
		g.data[key] = value
	}

	return value, nil
}

func (g *mockClusterGroup) describePeers() []cluster.PeerDesc {
	g.Lock()
	defer g.Unlock()

	result := make(cluster.PeerList, 0, len(g.peers))
	for _, p := range g.peers {
		result = append(result, p.me)
	}
	sort.Sort(result)
	return result
}

func (g *mockClusterGroup) updatePeer(peer *mockCluster, desc *cluster.PeerDesc, eventToSelf bool) error {
	g.Lock()
	defer g.Unlock()

	if peer.me.ID != "" && desc.ID != peer.me.ID {
		return fmt.Errorf("changing peer ID is not allowed")
	}

	found, ok := g.peers[desc.ID]
	if ok && found != peer {
		return fmt.Errorf("cluster has another member with ID %#v", desc.ID)
	}

	peer.me = *desc
	g.peers[desc.ID] = peer

	var ev cluster.PeerEvent
	if ok {
		ev = &cluster.PeerAliveEvent{*desc}
	} else {
		ev = &cluster.PeerJoinedEvent{*desc}
	}

	for _, p := range g.peers {
		if p == peer && !eventToSelf {
			continue
		}
		p.c <- ev
	}
	return nil
}

func (g *mockClusterGroup) removePeer(peer *mockCluster) {
	g.Lock()
	defer g.Unlock()

	peerID := peer.me.ID
	found, ok := g.peers[peerID]
	if !ok {
		return
	}
	if found != peer {
		panic("multiple peers with same ID in cluster")
	}

	delete(g.peers, peerID)
	peer.me = cluster.PeerDesc{}
	close(peer.c)

	ev := &cluster.PeerLostEvent{cluster.PeerDesc{ID: peerID}}
	for _, p := range g.peers {
		p.c <- ev
	}
}

func (g *mockClusterGroup) NewCluster(desc *cluster.PeerDesc) cluster.Cluster {
	// The channel must be buffered as the backend's background goroutine both reads to and writes from it.
	result := &mockCluster{
		c: make(chan cluster.PeerEvent, 16),
		g: g,
	}
	if desc != nil {
		if err := g.updatePeer(result, desc, false); err != nil {
			panic(err)
		}
	}
	return result
}

func MockCluster(desc *cluster.PeerDesc) cluster.Cluster {
	return NewMockClusterGroup().NewCluster(desc)
}

type mockCluster struct {
	g  *mockClusterGroup
	me cluster.PeerDesc
	c  chan cluster.PeerEvent
}

func (tc *mockCluster) GetDir(key string) (map[string]string, error) {
	return tc.g.getDir(key)
}

func (tc *mockCluster) GetValue(key string) (string, error) {
	return tc.g.getSet(key, func() (string, error) {
		return "", cluster.ErrNotFound
	}, false)
}

func (tc *mockCluster) SetValue(key, value string) error {
	_, err := tc.g.getSet(key, func() (string, error) {
		return value, nil
	}, true)
	return err
}

func (tc *mockCluster) GetValueWithDefault(key string, setter func() (string, error)) (string, error) {
	return tc.g.getSet(key, setter, false)
}

func (tc *mockCluster) GetSecret(kms security.KMS, name string, bytes int) ([]byte, error) {
	secret, err := tc.g.getSet("/secrets/"+strings.TrimLeft(name, "/"), func() (string, error) {
		secret, err := kms.GenerateNonce(bytes)
		if err != nil {
			return "", err
		}
		return string(secret), nil
	}, false)
	if err != nil {
		return nil, err
	}
	if len(secret) != bytes {
		return nil, fmt.Errorf("secret inconsistent: expected %d bytes, got %d", bytes, len(secret))
	}
	return []byte(secret), nil
}

func (tc *mockCluster) Update(desc *cluster.PeerDesc) error {
	return tc.g.updatePeer(tc, desc, true)
}

func (tc *mockCluster) Part() {
	tc.g.removePeer(tc)
}

func (tc *mockCluster) Peers() []cluster.PeerDesc {
	return tc.g.describePeers()
}

func (tc *mockCluster) Watch() <-chan cluster.PeerEvent {
	return tc.c
}
