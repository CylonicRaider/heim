package jobs

import (
	"container/heap"
	"sync/atomic"
	"time"

	"euphoria.leet.nu/lib/scope"

	"euphoria.leet.nu/heim/proto/logging"
)

var (
	LocalMinBackoff  = 1 * time.Second
	LocalMaxLifetime = 1 * time.Hour
)

type LocalJobCallback func() error

type LocalJobService interface {
	// Start launches the background goroutines powering the job queue.
	// The goroutines participate in the context's WaitGroup.
	Start(ctx scope.Context, numWorkers int)

	// Close initiates a clean shutdown of the job queue.
	// Acceptance of new jobs (including retries of failed jobs) stops before Close returns.
	// Asynchronously, the background coroutines drain all currently pending jobs at their respective due dates.
	Close()

	// Schedule the job's callback to run not before the job's due date.
	// Returns whether the job was accepted.
	Push(job *LocalJob) bool

	// Convenience function for scheduling the given callback to run ASAP.
	PushNew(name string, callback LocalJobCallback) bool
}

type LocalJob struct {
	Name    string
	Created time.Time
	Due     time.Time
	Func    LocalJobCallback
}

func NewLocalJob(name string, callback LocalJobCallback) *LocalJob {
	now := time.Now()
	return &LocalJob{
		Name:    name,
		Created: now,
		Due:     now,
		Func:    callback,
	}
}

func (j *LocalJob) IsDue() bool                { return j.IsDueAt(time.Now()) }
func (j *LocalJob) IsDueAt(now time.Time) bool { return j.Due.Before(now) }

func (j *LocalJob) Reschedule() bool {
	newDue := j.Created.Add(max(j.Due.Sub(j.Created)*2, LocalMinBackoff))
	now := time.Now()
	if newDue.Before(now) {
		newDue = now
	}
	if newDue.Sub(j.Created) > LocalMaxLifetime {
		return false
	}
	j.Due = newDue
	return true
}

func (j *LocalJob) Run() error {
	return j.Func()
}

type localJobHeap []*LocalJob

func (h localJobHeap) Len() int            { return len(h) }
func (h localJobHeap) Less(i, j int) bool  { return h[i].Due.Before(h[j].Due) }
func (h localJobHeap) Swap(i, j int)       { h[j], h[i] = h[i], h[j] }
func (h *localJobHeap) Push(x interface{}) { *h = append(*h, x.(*LocalJob)) }

func (h *localJobHeap) Pop() interface{} {
	item := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return item
}

type LocalJobQueue struct {
	data   localJobHeap
	inbox  chan<- *LocalJob
	outbox <-chan *LocalJob
	closed atomic.Bool
}

func NewLocalJobQueue() *LocalJobQueue {
	return &LocalJobQueue{}
}

func (q *LocalJobQueue) Push(job *LocalJob) bool {
	if q.closed.Load() {
		return false
	}
	q.inbox <- job
	return true
}

func (q *LocalJobQueue) PushNew(name string, callback LocalJobCallback) bool {
	return q.Push(NewLocalJob(name, callback))
}

func (q *LocalJobQueue) scheduler(ctx scope.Context, inbox <-chan *LocalJob, outbox chan<- *LocalJob) {
	defer ctx.WaitGroup().Done()
	defer close(outbox)

	// Is there a way to create a timer without starting it? In any case, one initial turn of the main loop will
	// do no harm.
	timer := time.NewTimer(0)
	defer timer.Stop()

	hasDue := false

	for {
		if hasDue {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-inbox:
				if ok {
					heap.Push(&q.data, job)
				} else {
					inbox = nil
				}
			case outbox <- q.data[0]:
				heap.Pop(&q.data)
				hasDue = false
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-inbox:
				if ok {
					heap.Push(&q.data, job)
				} else {
					inbox = nil
				}
			case <-timer.C:
			}
		}

		if len(q.data) == 0 {
			if inbox == nil {
				return
			}
			continue
		}
		now := time.Now()
		if !q.data[0].IsDueAt(now) {
			timer.Reset(q.data[0].Due.Sub(now))
			continue
		}
		hasDue = true
	}
}

func (q *LocalJobQueue) StartScheduler(ctx scope.Context) {
	inbox := make(chan *LocalJob, 16)
	outbox := make(chan *LocalJob)
	q.inbox = inbox
	q.outbox = outbox
	ctx.WaitGroup().Add(1)
	go q.scheduler(ctx, inbox, outbox)
}

func (q *LocalJobQueue) worker(ctx scope.Context, outbox <-chan *LocalJob) {
	defer ctx.WaitGroup().Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-outbox:
			if !ok {
				return
			} else if err := job.Run(); err == nil {
				// OK!
			} else if job.Reschedule() {
				if !q.Push(job) {
					logging.Logger(ctx).Printf(
						"giving up on job %s from %s due to closure, last error: %s",
						job.Name, job.Created, err)
				}
			} else {
				logging.Logger(ctx).Printf("giving up on job %s from %s, last error: %s",
					job.Name, job.Created, err)
			}
		}
	}
}

func (q *LocalJobQueue) StartWorker(ctx scope.Context) {
	ctx.WaitGroup().Add(1)
	go q.worker(ctx, q.outbox)
}

func (q *LocalJobQueue) Start(ctx scope.Context, numWorkers int) {
	q.StartScheduler(ctx)
	for i := 0; i < numWorkers; i++ {
		q.StartWorker(ctx)
	}
}

func (q *LocalJobQueue) Close() {
	if q.closed.Swap(true) {
		return
	}
	close(q.inbox)
}
