package backend_test

import (
	"testing"

	"euphoria.leet.nu/heim/backend"
	"euphoria.leet.nu/heim/backend/mock"
	"euphoria.leet.nu/heim/proto"
)

func TestBackend(t *testing.T) {
	backend.IntegrationTest(
		t, func(*proto.Heim) (proto.Backend, error) { return &mock.TestBackend{}, nil })
}
