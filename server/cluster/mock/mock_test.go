package mock

import (
	"testing"

	"euphoria.leet.nu/heim/cluster"
)

func TestMockCluster(t *testing.T) {
	cluster.BehavioralTest(t, MockCluster)
}
