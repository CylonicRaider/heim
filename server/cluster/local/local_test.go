package local

import (
	"os"
	"strings"
	"testing"

	"euphoria.leet.nu/heim/cluster"
)

func NewTestLocalCluster(desc *cluster.PeerDesc) cluster.Cluster {
	rootDir, err := os.MkdirTemp("", "heim-local-cluster-*")
	if err != nil {
		panic(err)
	}
	return cluster.AutoJoinOrPanic(&testLocalCluster{
		localCluster{
			rootDir: strings.TrimRight(rootDir, "/") + "/",
			c:       make(chan cluster.PeerEvent, 16),
		},
	}, desc)
}

type testLocalCluster struct {
	localCluster
}

func (lc *testLocalCluster) Part() {
	os.RemoveAll(lc.rootDir)
	lc.localCluster.Part()
}

func TestLocalCluster(t *testing.T) {
	cluster.BehavioralTest(t, NewTestLocalCluster)
}
