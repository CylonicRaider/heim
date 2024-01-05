package proto_test

import (
	"testing"

	"euphoria.leet.nu/heim/backend"
	"euphoria.leet.nu/heim/backend/mock"
	"euphoria.leet.nu/heim/proto"
)

func TestIntegration(t *testing.T) {
	backend.IntegrationTest(
		t, func(*proto.Heim) (proto.Backend, error) { return &mock.TestBackend{}, nil })
}
