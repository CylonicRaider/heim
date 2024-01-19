package local

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/proto/security"
)

type localCluster struct {
	sync.Mutex
	rootDir string
	me      *cluster.PeerDesc
	c       chan cluster.PeerEvent
}

func LocalCluster(rootDir string) cluster.Cluster {
	return &localCluster{
		rootDir: strings.TrimRight(rootDir, "/") + "/",
		c:       make(chan cluster.PeerEvent, 16),
	}
}

func (lc *localCluster) path(key string) string {
	return lc.rootDir + strings.TrimLeft(key, "/")
}

func (lc *localCluster) GetDir(key string) (map[string]string, error) {
	path := lc.path(key)

	lc.Lock()
	defer lc.Unlock()

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cluster.ErrNotFound
		} else {
			return nil, err
		}
	}

	result := map[string]string{}
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}

		name := ent.Name()
		bytes, err := os.ReadFile(filepath.Join(path, name))
		if err != nil {
			return nil, err
		}
		result[name] = string(bytes)
	}

	return result, nil
}

func (lc *localCluster) transaction(key string, setter func() (string, error), override bool) (string, error) {
	path := lc.path(key)

	lc.Lock()
	defer lc.Unlock()

	if !override {
		bytes, err := os.ReadFile(path)
		if err == nil {
			return string(bytes), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	if setter == nil {
		return "", cluster.ErrNotFound
	}

	newText, err := setter()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}

	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err == nil {
		_, err := fp.Write([]byte(newText))
		if err != nil {
			return "", err
		} else {
			return newText, err
		}
	} else if !os.IsExist(err) {
		return "", err
	}

	// yes, this is racy; however, this cluster implementation is not intended for concurrent use anyway
	time.Sleep(10 * time.Millisecond)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (lc *localCluster) GetValue(key string) (string, error) {
	return lc.transaction(key, nil, false)
}

func (lc *localCluster) SetValue(key, value string) error {
	_, err := lc.transaction(key, func() (string, error) { return value, nil }, true)
	return err
}

func (lc *localCluster) GetValueWithDefault(key string, setter func() (string, error)) (string, error) {
	return lc.transaction(key, setter, false)
}

func (lc *localCluster) GetSecret(kms security.KMS, name string, bytes int) ([]byte, error) {
	generateNew := func() (string, error) {
		newSecret, err := kms.GenerateNonce(bytes)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(newSecret), nil
	}

	resultText, err := lc.transaction("/secrets/"+name, generateNew, false)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(resultText)
}

func (lc *localCluster) update(desc *cluster.PeerDesc) cluster.PeerEvent {
	lc.Lock()
	defer lc.Unlock()

	hadMe := lc.me != nil
	copyOfDesc := *desc
	lc.me = &copyOfDesc
	if hadMe {
		return &cluster.PeerAliveEvent{copyOfDesc}
	} else {
		return &cluster.PeerJoinedEvent{copyOfDesc}
	}
}

func (lc *localCluster) Update(desc *cluster.PeerDesc) error {
	if event := lc.update(desc); event != nil {
		lc.c <- event
	}
	return nil
}

func (lc *localCluster) part() cluster.PeerEvent {
	lc.Lock()
	defer lc.Unlock()

	oldDesc := lc.me
	lc.me = nil
	if oldDesc != nil {
		return &cluster.PeerLostEvent{*oldDesc}
	} else {
		return nil
	}
}

func (lc *localCluster) Part() {
	if event := lc.part(); event != nil {
		lc.c <- event
	}
}

func (lc *localCluster) Peers() []cluster.PeerDesc {
	lc.Lock()
	defer lc.Unlock()
	result := []cluster.PeerDesc{}
	if lc.me != nil {
		result = append(result, *lc.me)
	}
	return result
}

func (lc *localCluster) Watch() <-chan cluster.PeerEvent {
	return lc.c
}
