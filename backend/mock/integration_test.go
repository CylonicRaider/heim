package mock

import (
	"testing"

	"euphoria.leet.nu/heim/backend"
	"euphoria.leet.nu/heim/proto"
)

func TestTestBackend(t *testing.T) {
	backend.IntegrationTest(t, func(*proto.Heim) (proto.Backend, error) { return &TestBackend{}, nil })
}
