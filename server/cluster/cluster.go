package cluster

import (
	"fmt"
	"time"

	"euphoria.leet.nu/heim/proto/security"
)

var (
	TTL = 30 * time.Second

	ErrNotFound = fmt.Errorf("not found")
)

type Cluster interface {
	GetDir(key string) (map[string]string, error)
	GetValue(key string) (string, error)
	SetValue(key, value string) error
	GetValueWithDefault(key string, setter func() (string, error)) (string, error)

	GetSecret(kms security.KMS, name string, bytes int) ([]byte, error)

	Update(desc *PeerDesc) error
	Part()
	Peers() []PeerDesc
	Watch() <-chan PeerEvent
}

type PeerEvent interface {
	Peer() *PeerDesc
}

type PeerDesc struct {
	ID      string `json:"id"`
	Era     string `json:"era"`
	Version string `json:"version"`
}

func (p *PeerDesc) Peer() *PeerDesc { return p }

type PeerJoinedEvent struct {
	PeerDesc
}

type PeerAliveEvent struct {
	PeerDesc
}

type PeerLostEvent struct {
	PeerDesc
}

type PeerList []PeerDesc

func (ps PeerList) Len() int           { return len(ps) }
func (ps PeerList) Less(i, j int) bool { return ps[i].ID < ps[j].ID }
func (ps PeerList) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
