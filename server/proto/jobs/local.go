package jobs

import (
	"container/heap"
	"time"

	"euphoria.leet.nu/lib/scope"

	"euphoria.leet.nu/heim/proto/logging"
)

var (
	LocalMinBackoff  = 1 * time.Second
	LocalMaxLifetime = 1 * time.Hour
)

type LocalJob struct {
	Name    string
	Created time.Time
	Due     time.Time
	Func    func() error
}

func NewLocalJob(name string, callback func() error) *LocalJob {
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
	delay := max(j.Due.Sub(j.Created)*2, LocalMinBackoff)
	if delay > LocalMaxLifetime {
		return false
	}
	j.Due = j.Created.Add(delay)
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
}

func NewLocalJobQueue() *LocalJobQueue {
	return &LocalJobQueue{}
}

func (q *LocalJobQueue) Push(job *LocalJob) {
	q.inbox <- job
}

func (q *LocalJobQueue) PushNew(name string, callback func() error) {
	q.Push(NewLocalJob(name, callback))
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
				q.Push(job)
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

func (q *LocalJobQueue) Close() {
	close(q.inbox)
}
