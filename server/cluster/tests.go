package cluster

import (
	"fmt"
	"testing"
	"time"

	"euphoria.leet.nu/heim/proto/security"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	debugStackTraces = false
)

func init() {
	if debugStackTraces {
		SetDefaultStackMode(StackFail)
	}
}

func expectNoPeerEvents(watch <-chan PeerEvent) error {
	select {
	case ev, ok := <-watch:
		if !ok {
			return fmt.Errorf("Expected no peer events, got channel close")
		}
		return fmt.Errorf("Expected no peer events, got %#v", ev)
	case <-time.After(10 * time.Millisecond):
		return nil
	}
}

func expectOnePeerEvent(watch <-chan PeerEvent) (PeerEvent, error) {
	var result PeerEvent
	select {
	case ev, ok := <-watch:
		if !ok {
			return nil, fmt.Errorf("Expected peer event, got channel close")
		}
		result = ev
	case <-time.After(10 * time.Millisecond):
		return nil, fmt.Errorf("Expected peer event, got none")
	}
	select {
	case ev, ok := <-watch:
		if ok {
			return nil, fmt.Errorf("Expected exactly one peer event, got %#v and %#v", result, ev)
		}
	case <-time.After(10 * time.Millisecond):
	}
	return result, nil
}

func expectClose(watch <-chan PeerEvent) error {
	select {
	case ev, ok := <-watch:
		if ok {
			return fmt.Errorf("Expected channel close, got %#v", ev)
		}
		return nil
	case <-time.After(10 * time.Millisecond):
		return fmt.Errorf("Expected channel close, got none")
	}
}

func BehavioralTest(t *testing.T, clusterFactory func(desc *PeerDesc) Cluster, supportsMultiple bool) {
	newNonce := func() string {
		return time.Now().String()
	}

	defaultPeerDesc := func() *PeerDesc {
		return &PeerDesc{
			ID:  "test",
			Era: newNonce(),
		}
	}

	namedPeerDesc := func(name string) *PeerDesc {
		return &PeerDesc{
			ID:  name,
			Era: newNonce(),
		}
	}

	Convey("Initially without data", t, func() {
		cluster := clusterFactory(defaultPeerDesc())
		defer cluster.Part()

		_, err := cluster.GetValue("nonexistent")
		So(err, ShouldEqual, ErrNotFound)
	})

	Convey("Get and set", t, func() {
		cluster := clusterFactory(defaultPeerDesc())
		defer cluster.Part()

		_, err := cluster.GetValue("getset")
		So(err, ShouldEqual, ErrNotFound)

		nonce := newNonce()
		err = cluster.SetValue("getset", nonce)
		So(err, ShouldBeNil)

		v, err := cluster.GetValue("getset")
		So(err, ShouldBeNil)
		So(v, ShouldEqual, nonce)
	})

	Convey("Get with default", t, func() {
		cluster := clusterFactory(defaultPeerDesc())
		defer cluster.Part()

		Convey("Handles setter errors", func() {
			testError := fmt.Errorf("%s", newNonce())

			setterCalled := false
			_, err := cluster.GetValueWithDefault("getdefault", func() (string, error) {
				setterCalled = true
				return "", testError
			})
			So(setterCalled, ShouldBeTrue)
			So(err, ShouldEqual, testError)

			_, err = cluster.GetValue("getdefault")
			So(err, ShouldEqual, ErrNotFound)
		})

		Convey("Creates and retains value", func() {
			nonce := newNonce()
			setterCalled := false
			got, err := cluster.GetValueWithDefault("getdefault", func() (string, error) {
				setterCalled = true
				return nonce, nil
			})
			So(err, ShouldBeNil)
			So(setterCalled, ShouldBeTrue)
			So(got, ShouldEqual, nonce)

			nonce2 := newNonce()
			So(nonce2, ShouldNotEqual, nonce)
			setterCalled = false
			got, err = cluster.GetValueWithDefault("getdefault", func() (string, error) {
				setterCalled = true
				return nonce2, nil
			})
			So(err, ShouldBeNil)
			So(setterCalled, ShouldBeFalse)
			So(got, ShouldEqual, nonce)
		})

		Convey("Handles race condition correctly", func() {
			type result struct {
				Data  string
				Error error
			}

			bgsc := make(chan struct{})
			bgrc := make(chan result)
			go func() {
				value, err := cluster.GetValueWithDefault("getdefaultrace", func() (string, error) {
					bgsc <- struct{}{}
					<-bgsc
					return "this should not be the result", nil
				})
				bgrc <- result{Data: value, Error: err}
			}()

			// Wait for background goroutine to enter the setter function
			<-bgsc

			// Let another call succeed first
			text, err := cluster.GetValueWithDefault("getdefaultrace", func() (string, error) {
				return "this should be the result", nil
			})
			So(err, ShouldBeNil)
			So(text, ShouldEqual, "this should be the result")

			// Release the goroutine
			bgsc <- struct{}{}

			// Retrieve and validate the goroutine's result
			bgResult := <-bgrc
			So(bgResult.Error, ShouldBeNil)
			So(bgResult.Data, ShouldEqual, text)
		})
	})

	Convey("Secrets", t, func() {
		cluster := clusterFactory(defaultPeerDesc())
		defer cluster.Part()

		Convey("Basic", func() {
			kms := &testKMS{security.LocalKMS(), false, nil}

			result, err := cluster.GetSecret(kms, "testsecret", 17)
			So(err, ShouldBeNil)
			So(kms.PopGenerateNonceCalled(), ShouldBeTrue)
			So(result, ShouldHaveLength, 17)

			result2, err := cluster.GetSecret(kms, "testsecret", 17)
			So(err, ShouldBeNil)
			So(kms.PopGenerateNonceCalled(), ShouldBeFalse)
			So(result2, ShouldEqual, result)

			_, err = cluster.GetSecret(kms, "testsecret", 13)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "secret inconsistent:")
			So(kms.PopGenerateNonceCalled(), ShouldBeFalse)
		})

		Convey("Handles race condition correctly", func() {
			kms := &testKMS{security.LocalKMS(), false, make(chan struct{})}

			type result struct {
				Data  []byte
				Error error
			}

			bgrc := make(chan result)
			go func() {
				secret, err := cluster.GetSecret(kms, "secretrace", 23)
				bgrc <- result{Data: secret, Error: err}
			}()

			// Wait until the goroutine is in the middle of its GetSecret call
			<-kms.gnSync

			// Now, let another call succeed first
			secret, err := cluster.GetSecret(kms.KMS, "secretrace", 23)
			So(err, ShouldBeNil)

			// Unblock the goroutine
			kms.gnSync <- struct{}{}

			// Get and validate the goroutine's result
			bgResult := <-bgrc
			So(bgResult.Error, ShouldBeNil)
			So(bgResult.Data, ShouldEqual, secret)
		})
	})

	Convey("Presence", t, func() {
		Convey("Without initial announce", func() {
			cluster := clusterFactory(nil)
			parted := false
			defer func() {
				if !parted {
					cluster.Part()
				}
			}()

			// Initially, the cluster should be empty.
			peers := cluster.Peers()
			So(peers, ShouldBeEmpty)

			events := cluster.Watch()
			err := expectNoPeerEvents(events)
			So(err, ShouldBeNil)

			// The first update should be a join.
			refPeerDesc := PeerDesc{ID: "test1", Era: newNonce()}
			peerDesc := refPeerDesc
			err = cluster.Update(&peerDesc)
			So(err, ShouldBeNil)
			So(peerDesc, ShouldEqual, refPeerDesc)

			ev, err := expectOnePeerEvent(events)
			So(err, ShouldBeNil)
			So(ev, ShouldResemble, &PeerJoinedEvent{refPeerDesc})

			So(cluster.Peers(), ShouldEqual, []PeerDesc{refPeerDesc})

			// Another update should lead to an alive event.
			refPeerDesc.Era = newNonce()
			peerDesc = refPeerDesc
			err = cluster.Update(&peerDesc)
			So(err, ShouldBeNil)
			So(peerDesc, ShouldEqual, refPeerDesc)

			ev, err = expectOnePeerEvent(events)
			So(err, ShouldBeNil)
			So(ev, ShouldResemble, &PeerAliveEvent{refPeerDesc})

			So(cluster.Peers(), ShouldEqual, []PeerDesc{refPeerDesc})

			// A peer should not observe its own part event.
			parted = true
			cluster.Part()

			err = expectClose(events)
			So(err, ShouldBeNil)
		})

		Convey("With initial announce", func() {
			peerDesc := *defaultPeerDesc()
			cluster := clusterFactory(&peerDesc)
			defer cluster.Part()

			// Initially, there should be one peer; we should not see ourselves joining.
			peers := cluster.Peers()
			So(peers, ShouldEqual, []PeerDesc{peerDesc})

			events := cluster.Watch()
			err := expectNoPeerEvents(events)
			So(err, ShouldBeNil)

			// This update should not be a join, either.
			newPeerDesc := PeerDesc{ID: peerDesc.ID, Era: newNonce()}
			err = cluster.Update(&newPeerDesc)
			So(err, ShouldBeNil)

			ev, err := expectOnePeerEvent(events)
			So(err, ShouldBeNil)
			So(ev, ShouldResemble, &PeerAliveEvent{newPeerDesc})

			So(cluster.Peers(), ShouldEqual, []PeerDesc{newPeerDesc})
		})
	})

	if !supportsMultiple {
		return
	}

	Convey("Concurrent presence", t, func() {
		// Launch first cluster member.
		descA := *namedPeerDesc("test-a")
		clusterA := clusterFactory(&descA)
		defer clusterA.Part()

		eventsA := clusterA.Watch()
		So(expectNoPeerEvents(eventsA), ShouldBeNil)

		So(clusterA.Peers(), ShouldEqual, []PeerDesc{descA})

		// Launch second cluster member.
		descB := *namedPeerDesc("test-b")
		clusterB := clusterFactory(&descB)
		partedB := false
		defer func() {
			if !partedB {
				clusterB.Part()
			}
		}()

		eventsB := clusterB.Watch()
		So(expectNoPeerEvents(eventsB), ShouldBeNil)

		ev, err := expectOnePeerEvent(eventsA)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerJoinedEvent{descB})

		So(clusterA.Peers(), ShouldEqual, []PeerDesc{descA, descB})
		So(clusterB.Peers(), ShouldEqual, []PeerDesc{descA, descB})

		// Update a member's data.
		descB.Era = newNonce()
		clusterB.Update(&descB)

		ev, err = expectOnePeerEvent(eventsA)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerAliveEvent{descB})

		ev, err = expectOnePeerEvent(eventsB)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerAliveEvent{descB})

		So(clusterA.Peers(), ShouldEqual, []PeerDesc{descA, descB})
		So(clusterB.Peers(), ShouldEqual, []PeerDesc{descA, descB})

		// Another time, just to be sure.
		descB.Era = newNonce()
		clusterB.Update(&descB)

		ev, err = expectOnePeerEvent(eventsA)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerAliveEvent{descB})

		ev, err = expectOnePeerEvent(eventsB)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerAliveEvent{descB})

		So(clusterA.Peers(), ShouldEqual, []PeerDesc{descA, descB})
		So(clusterB.Peers(), ShouldEqual, []PeerDesc{descA, descB})

		// Remove one peer.
		partedB = true
		clusterB.Part()

		So(expectClose(eventsB), ShouldBeNil)

		ev, err = expectOnePeerEvent(eventsA)
		So(err, ShouldBeNil)
		So(ev, ShouldResemble, &PeerLostEvent{PeerDesc{ID: descB.ID}})

		So(clusterA.Peers(), ShouldEqual, []PeerDesc{descA})
	})
}

type testKMS struct {
	security.KMS
	gnCalled bool
	gnSync   chan struct{}
}

func (tkms *testKMS) GenerateNonce(bytes int) ([]byte, error) {
	tkms.gnCalled = true
	if tkms.gnSync != nil {
		tkms.gnSync <- struct{}{}
		<-tkms.gnSync
	}
	return tkms.KMS.GenerateNonce(bytes)
}

func (tkms *testKMS) PopGenerateNonceCalled() bool {
	result := tkms.gnCalled
	tkms.gnCalled = false
	return result
}
