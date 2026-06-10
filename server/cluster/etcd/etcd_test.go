package etcd_test

import (
	"testing"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/cluster/etcd/clustertest"
)

func TestEtcdCluster(t *testing.T) {
	etcd, err := clustertest.StartEtcd()
	if err != nil {
		t.Fatal(err)
	}
	if etcd == nil {
		t.Skipf("etcd not in PATH, skipping etcd cluster tests")
	}
	defer etcd.Shutdown()

	cluster.BehavioralTest(t, func(desc *cluster.PeerDesc) cluster.Cluster {
		return etcd.Join("/tests", desc)
	}, true)
}
