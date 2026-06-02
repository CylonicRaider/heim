package local

import (
	"os"
	"testing"

	"euphoria.leet.nu/lib/scope"

	"euphoria.leet.nu/heim/cluster"
)

func NewTestLocalCluster(desc *cluster.PeerDesc) cluster.Cluster {
	rootDir, err := os.MkdirTemp("", "heim-local-cluster-*")
	if err != nil {
		panic(err)
	}
	return testLocalCluster{LocalCluster(scope.New(), rootDir, desc).(*localCluster)}
}

type testLocalCluster struct {
	*localCluster
}

func (lc testLocalCluster) Part() {
	os.RemoveAll(lc.rootDir)
	lc.localCluster.Part()
}

func TestLocalCluster(t *testing.T) {
	cluster.BehavioralTest(t, NewTestLocalCluster)
}
