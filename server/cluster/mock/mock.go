package mock

import (
	"fmt"
	"strings"
	"sync"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/proto/security"
)

func MockCluster() cluster.Cluster {
	// The channel must be buffered as the backend's background goroutine both reads to and writes from it.
	return &mockCluster{c: make(chan cluster.PeerEvent, 16)}
}

type mockCluster struct {
	sync.Mutex
	data    map[string]string
	peers   map[string]cluster.PeerDesc
	secrets map[string][]byte
	c       chan cluster.PeerEvent
	myID    string
}

func (tc *mockCluster) GetDir(key string) (map[string]string, error) {
	tc.Lock()
	defer tc.Unlock()
	key = strings.TrimRight(key, "/") + "/"
	result := map[string]string{}
	for k, v := range tc.data {
		if strings.HasPrefix(k, key) {
			rk := k[len(key):]
			if !strings.Contains(rk, "/") {
				result[rk] = v
			}
		}
	}
	return result, nil
}

func (tc *mockCluster) GetValue(key string) (string, error) {
	tc.Lock()
	defer tc.Unlock()
	data, ok := tc.data[key]
	if !ok {
		return "", cluster.ErrNotFound
	}
	return data, nil
}

func (tc *mockCluster) SetValue(key, value string) error {
	tc.Lock()
	if tc.data == nil {
		tc.data = map[string]string{key: value}
	} else {
		tc.data[key] = value
	}
	tc.Unlock()
	return nil
}

func (tc *mockCluster) GetValueWithDefault(key string, setter func() (string, error)) (string, error) {
	tc.Lock()
	defer tc.Unlock()
	if val, ok := tc.data[key]; ok {
		return val, nil
	}
	val, err := setter()
	if err != nil {
		return "", err
	}
	if tc.data == nil {
		tc.data = map[string]string{}
	}
	tc.data[key] = val
	return val, nil
}

func (tc *mockCluster) GetSecret(kms security.KMS, name string, bytes int) ([]byte, error) {
	tc.Lock()
	defer tc.Unlock()

	if secret, ok := tc.secrets[name]; ok {
		if len(secret) != bytes {
			return nil, fmt.Errorf("secret inconsistent: expected %d bytes, got %d", bytes, len(secret))
		}
		return secret, nil
	}

	secret, err := kms.GenerateNonce(bytes)
	if err != nil {
		return nil, err
	}

	if tc.secrets == nil {
		tc.secrets = map[string][]byte{name: secret}
	} else {
		tc.secrets[name] = secret
	}
	return secret, nil
}

func (tc *mockCluster) update(desc *cluster.PeerDesc) cluster.PeerEvent {
	tc.Lock()
	defer tc.Unlock()

	if tc.myID == "" {
		tc.myID = desc.ID
	}

	if tc.peers == nil {
		tc.peers = map[string]cluster.PeerDesc{}
	}

	if tc.c == nil {
		tc.peers[desc.ID] = *desc
		return nil
	}

	_, ok := tc.peers[desc.ID]
	tc.peers[desc.ID] = *desc
	if ok {
		return &cluster.PeerAliveEvent{*desc}
	} else {
		return &cluster.PeerJoinedEvent{*desc}
	}
}

func (tc *mockCluster) Update(desc *cluster.PeerDesc) error {
	if event := tc.update(desc); event != nil {
		tc.c <- event
	}
	return nil
}

func (tc *mockCluster) part() cluster.PeerEvent {
	tc.Lock()
	defer tc.Unlock()
	desc, ok := tc.peers[tc.myID]
	delete(tc.peers, tc.myID)
	if ok {
		return &cluster.PeerLostEvent{desc}
	}
	return nil
}

func (tc *mockCluster) Part() {
	if event := tc.part(); event != nil {
		tc.c <- event
	}
}

func (tc *mockCluster) Peers() []cluster.PeerDesc {
	tc.Lock()
	defer tc.Unlock()
	peers := []cluster.PeerDesc{}
	for _, peer := range tc.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (tc *mockCluster) Watch() <-chan cluster.PeerEvent {
	tc.Lock()
	defer tc.Unlock()
	if tc.c == nil {
		tc.c = make(chan cluster.PeerEvent)
	}
	return tc.c
}
