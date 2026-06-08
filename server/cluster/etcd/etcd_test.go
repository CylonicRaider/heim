package etcd_test

import (
	"testing"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/cluster/etcd/clustertest"
	"euphoria.leet.nu/heim/proto/security"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEtcdCluster(t *testing.T) {
	s, err := clustertest.StartEtcd()
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Skipf("etcd not in PATH, skipping tests")
	}
	defer s.Shutdown()

	peerDesc := func(id string) *cluster.PeerDesc {
		return &cluster.PeerDesc{
			ID:  id,
			Era: "0",
		}
	}
	peerDescNE := func(id string) *cluster.PeerDesc {
		return &cluster.PeerDesc{
			ID: id,
		}
	}

	cluster.BehavioralTest(t, func(desc *cluster.PeerDesc) cluster.Cluster {
		return s.Join("/general", desc)
	}, true)

	Convey("Observe peer departure", t, func() {
		a := s.Join("/departure", peerDesc("a"))
		// no defer a.Part() because we'll do that explicitly
		b := s.Join("/departure", peerDesc("b"))
		defer b.Part()

		So(<-a.Watch(), ShouldResemble, &cluster.PeerJoinedEvent{*peerDesc("b")})
		a.Part()
		So(<-b.Watch(), ShouldResemble, &cluster.PeerLostEvent{*peerDescNE("a")})
	})

	Convey("Observe initial peers upon joining", t, func() {
		a := s.Join("/initial", peerDesc("a"))
		defer a.Part()
		So(a.Peers(), ShouldResemble, []cluster.PeerDesc{*peerDesc("a")})

		b := s.Join("/initial", peerDesc("b"))
		defer b.Part()
		So(b.Peers(), ShouldResemble, []cluster.PeerDesc{*peerDesc("a"), *peerDesc("b")})
	})

	Convey("Updates are seen", t, func() {
		a := s.Join("/updates", peerDesc("a"))
		defer a.Part()
		b := s.Join("/updates", peerDesc("b"))
		defer b.Part()

		b.Update(&cluster.PeerDesc{ID: "b", Era: "1"})
		b.Update(&cluster.PeerDesc{ID: "b", Era: "2"})
		So(<-a.Watch(), ShouldResemble, &cluster.PeerJoinedEvent{cluster.PeerDesc{ID: "b", Era: "0"}})
		So(<-a.Watch(), ShouldResemble, &cluster.PeerAliveEvent{cluster.PeerDesc{ID: "b", Era: "1"}})
		So(<-a.Watch(), ShouldResemble, &cluster.PeerAliveEvent{cluster.PeerDesc{ID: "b", Era: "2"}})
	})

	Convey("Secrets are created if necessary", t, func() {
		kms := security.LocalKMS()
		a := s.Join("/secrets1", peerDesc("a"))
		defer a.Part()

		secret, err := a.GetSecret(kms, "test1", 16)
		So(err, ShouldBeNil)
		So(len(secret), ShouldEqual, 16)

		secretCopy, err := a.GetSecret(kms, "test1", 16)
		So(err, ShouldBeNil)
		So(string(secretCopy), ShouldEqual, string(secret))
	})

	Convey("Race to create secret is conceded gracefully", t, func() {
		kms := &syncKMS{
			KMS: security.LocalKMS(),
			c:   make(chan struct{}),
		}

		a := s.Join("/secrets1", peerDesc("a"))
		defer a.Part()

		sc := make(chan []byte)
		errc := make(chan error)
		go func() {
			s, err := a.GetSecret(kms, "test2", 16)
			errc <- err
			sc <- s
		}()

		// Synchronize with secret generation.
		<-kms.c

		// Set the secret before releasing the goroutine.
		secret, err := a.GetSecret(kms.KMS, "test2", 16)
		So(err, ShouldBeNil)

		// Release the goroutine and verify it gets the secret that was set.
		kms.c <- struct{}{}
		So(<-errc, ShouldBeNil)
		So(string(<-sc), ShouldEqual, string(secret))
	})
}

type syncKMS struct {
	security.KMS
	c chan struct{}
}

func (s *syncKMS) GenerateNonce(bytes int) ([]byte, error) {
	s.c <- struct{}{}
	<-s.c
	return s.KMS.GenerateNonce(bytes)
}
