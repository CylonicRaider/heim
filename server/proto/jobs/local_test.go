package jobs

import (
	"fmt"
	"io"
	"testing"
	"time"

	"euphoria.leet.nu/heim/proto/logging"

	"euphoria.leet.nu/lib/scope"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	debugLogs = false
)

func receiveNowOrNil(ch <-chan bool) *bool {
	select {
	case v := <-ch:
		return &v
	default:
		return nil
	}
}

func TestLocalJobQueue(t *testing.T) {
	savedBackoff, savedMaxLifetime := LocalMinBackoff, LocalMaxLifetime
	defer func() {
		LocalMinBackoff, LocalMaxLifetime = savedBackoff, savedMaxLifetime
	}()
	LocalMinBackoff = 10 * time.Millisecond
	LocalMaxLifetime = 50 * time.Millisecond

	Convey("Local job queue", t, func() {
		jq := NewLocalJobQueue()
		closed := false
		doClose := func() {
			if !closed {
				jq.Close()
				closed = true
			}
		}
		defer doClose()

		ctx := scope.New()
		if !debugLogs {
			logging.SetDefaultWriter(ctx, io.Discard)
		}
		jq.StartScheduler(ctx)
		jq.StartWorker(ctx)

		Convey("Shuts down", func() {
			doClose()
			// This should not hang:
			ctx.WaitGroup().Wait()
			So(nil, ShouldBeNil)
		})

		Convey("Executes jobs", func() {
			ch := make(chan bool)
			jq.PushNew("test job 1", func() error { ch <- true; return nil })
			So(<-ch, ShouldBeTrue)
		})

		Convey("Executes multiple jobs", func() {
			chA, chB := make(chan bool, 1), make(chan bool, 1)
			jq.PushNew("test job 2a", func() error { chA <- true; return nil })
			jq.PushNew("test job 2b", func() error { chB <- true; return nil })
			So(<-chA, ShouldBeTrue)
			So(<-chB, ShouldBeTrue)
		})

		Convey("Executes jobs concurrently", func() {
			jq.StartWorker(ctx)

			chA, chB := make(chan bool), make(chan bool)
			chAB, chBA := make(chan bool), make(chan bool)
			jq.PushNew("test job 3a", func() error {
				chAB <- true
				<-chBA
				chA <- true
				return nil
			})
			jq.PushNew("test job 3b", func() error {
				<-chAB
				chBA <- true
				chB <- true
				return nil
			})
			So(<-chA, ShouldBeTrue)
			So(<-chB, ShouldBeTrue)
		})

		Convey("Finishes pending jobs when closed", func() {
			sync1, sync2 := make(chan bool), make(chan bool)
			jq.PushNew("test job 4a", func() error {
				sync1 <- true
				<-sync2
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			// Wait for the worker to pick up the first job.
			<-sync1

			jobBDone := make(chan bool)
			jq.PushNew("test job 4b", func() error {
				jobBDone <- true
				return nil
			})

			// Close before the first job is complete.
			doClose()
			// Unblock the first job; the second one must run after the queue has been closed.
			sync2 <- true

			syncWG := make(chan bool)
			go func() {
				ctx.WaitGroup().Wait()
				syncWG <- true
			}()

			// The wait group should not be done yet.
			So(receiveNowOrNil(syncWG), ShouldBeNil)
			// The next line should not hang.
			So(<-jobBDone, ShouldBeTrue)
			// Allow the background goroutines to clean themselves up.
			time.Sleep(10 * time.Millisecond)
			// Now, the wait group *should* be done.
			So(receiveNowOrNil(syncWG), ShouldNotBeNil)
		})

		Convey("Retries failed jobs", func() {
			attempts, done := 0, make(chan bool)
			jq.PushNew("test job 5", func() error {
				attempts++
				if attempts == 1 {
					return fmt.Errorf("not yet")
				} else {
					done <- true
					return nil
				}
			})
			So(<-done, ShouldBeTrue)
			So(attempts, ShouldEqual, 2)
		})

		Convey("Abandons persistently failing jobs", func() {
			attempts := 0
			jq.PushNew("test job 6", func() error {
				attempts++
				// The API does not provide feedback on job status (maybe this is a sign that it
				// should?), so we have to rely on timing.
				time.Sleep(2 * LocalMaxLifetime)
				return fmt.Errorf("no")
			})
			time.Sleep(3 * LocalMaxLifetime)
			So(attempts, ShouldEqual, 1)
		})
	})
}
