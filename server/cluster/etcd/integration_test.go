package etcd_test

import (
	"testing"

	"euphoria.leet.nu/heim/backend"
	"euphoria.leet.nu/heim/backend/mock"
	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/cluster/etcd/clustertest"
	"euphoria.leet.nu/heim/proto"
)

func TestIntegration(t *testing.T) {
	etcd, err := clustertest.StartEtcd()
	if err != nil {
		t.Fatal(err)
	}
	if etcd == nil {
		t.Skipf("etcd not in PATH, skipping etcd integration tests")
	}
	defer etcd.Shutdown()

	backend.IntegrationTest(
		t, func(heim *proto.Heim) (proto.Backend, error) {
			heim.Cluster = etcd.Join("/test", &cluster.PeerDesc{ID: "testcase", Era: "era"})
			return &mock.TestBackend{}, nil
		})
}
