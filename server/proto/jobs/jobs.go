package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/snowflake"
	"github.com/euphoria-io/scope"
)

const (
	DefaultMaxWorkDuration = time.Minute

	EmailQueue = "emails"
)

type JobType string

var (
	BackoffDuration = 5 * time.Second

	EmailJobType    = JobType("email")
	EmailJobOptions = []JobOption{
		JobOptions.MaxAttempts(3),
		JobOptions.MaxWorkDuration(30 * time.Second),
	}

	jobPayloadMap = map[JobType]reflect.Type{
		EmailJobType: reflect.TypeOf(EmailJob{}),
	}
)

type EmailJob struct {
	AccountID snowflake.Snowflake
	EmailID   string
}

type JobService interface {
	GetQueue(ctx scope.Context, name string) (JobQueue, error)
}

type JobQueue interface {
	// Name returns the name of the queue.
	Name() string

	// Add enqueues a new job, as defined by the given type/payload.
	// If any callers waiting in WaitForJob (not just in the local
	// process), at least one should be woken.
	Add(ctx scope.Context, jobType JobType, payload interface{}, options ...JobOption) (
		snowflake.Snowflake, error)

	// AddAndClaim enqueues a new job, atomically marking it as claimed by the
	// caller. Returns the added and claimed job.
	AddAndClaim(
		ctx scope.Context, jobType JobType, payload interface{}, handlerID string, options ...JobOption) (*Job, error)

	// WaitForJob blocks until notification of a new claimable job
	// in the queue. This does not guarantee that a job will be
	// immediately claimable.
	WaitForJob(ctx scope.Context) error

	// TryClaim tries to acquire a currently unclaimed job. If none is
	// available, returns ErrJobNotFound.
	TryClaim(ctx scope.Context, handlerID string) (*Job, error)

	// TrySteal attempts to preempt another handler's claim. Only jobs
	// that have been claimed longer than their MaxWorkDuration setting
	// can be stolen. Only jobs claimed by a different handlerID can
	// be stolen.
	//
	// If no job can be immediately stolen, returns ErrJobNotFound.
	//
	// Stolen jobs are at risk of being completed twice. It's important
	// for handlers to set a completion timeout and self-cancel well
	// within the job's MaxWorkDuration.
	TrySteal(ctx scope.Context, handlerID string) (*Job, error)

	// Cancel removes a job from the queue. If it is currently claimed, then it
	// may still be completed by the handler that claimed it, but no future call
	// to Claim or Steal will return this job.
	Cancel(ctx scope.Context, jobID snowflake.Snowflake) error

	// Complete marks a job as completed. If the job has been stolen
	// by another handler, the queueing service should attempt to
	// cancel the other handler's work in progress, but this cannot
	// be guaranteed.
	Complete(ctx scope.Context, jobID snowflake.Snowflake, handlerID string, attemptNumber int32, log []byte) error

	// Fail marks a job as failed and releases the claim on it.
	// If the job has not been stolen and still has attempts
	// remaining, it will return to the queue and be immediately
	// up for claim again.
	Fail(ctx scope.Context, jobID snowflake.Snowflake, handlerID string, attemptNumber int32, reason string, log []byte) error

	// Stats returns information about the number of jobs in the queue.
	Stats(ctx scope.Context) (JobQueueStats, error)

	// Log returns the output of a given job attempt.
	Log(ctx scope.Context, jobID snowflake.Snowflake, attemptNumber int32) (*JobLog, error)
}

type JobOption interface {
	Apply(*Job) error
}

type JobMaxAttempts int32

func (a JobMaxAttempts) Apply(job *Job) error {
	job.AttemptsRemaining = int32(a)
	return nil
}

type JobMaxWorkDuration time.Duration

func (d JobMaxWorkDuration) Apply(job *Job) error {
	job.MaxWorkDuration = time.Duration(d)
	return nil
}

type JobDue time.Time

func (t JobDue) Apply(job *Job) error {
	job.Due = time.Time(t)
	return nil
}

type JobOptionConstructor struct{}

func (JobOptionConstructor) MaxAttempts(n int32) JobMaxAttempts { return JobMaxAttempts(n) }

func (JobOptionConstructor) MaxWorkDuration(d time.Duration) JobMaxWorkDuration {
	return JobMaxWorkDuration(d)
}

func (JobOptionConstructor) Due(t time.Time) JobDue { return JobDue(t) }

var JobOptions JobOptionConstructor

type JobQueueStats struct {
	Waiting int64 // number of jobs waiting to be claimed
	Due     int64 // number of jobs that are due (whether claimed or waiting)
	Claimed int64 // number of jobs currently claimed
}

type Job struct {
	ID                snowflake.Snowflake
	Type              JobType
	Data              json.RawMessage
	Created           time.Time
	Due               time.Time
	MaxWorkDuration   time.Duration
	AttemptsMade      int32
	AttemptsRemaining int32

	*JobClaim
}

func (j *Job) Payload() (interface{}, error) {
	payloadType, ok := jobPayloadMap[j.Type]
	if !ok {
		return nil, fmt.Errorf("invalid job type: %s", j.Type)
	}
	payload := reflect.New(payloadType).Interface()
	if payload != nil && payloadType.NumField() > 0 {
		if err := json.Unmarshal(j.Data, payload); err != nil {
			return nil, err
		}
	}
	return payload, nil
}

func (j *Job) Encode() ([]byte, error) { return json.Marshal(j) }

func (j *Job) Exec(ctx scope.Context, f func(scope.Context) error) error {
	if j.JobClaim == nil {
		return ErrJobNotClaimed
	}

	w := io.MultiWriter(logging.GetDefaultWriterOrFallback(ctx, os.Stdout), j)
	prefix := fmt.Sprintf("[%s-%s] ", j.Queue.Name(), j.HandlerID)
	deadline := time.Now().Add(j.MaxWorkDuration)
	child := logging.LoggingContextOverride(ctx.ForkWithTimeout(j.MaxWorkDuration), w, prefix)
	if err := f(child); err != nil {
		logging.Logger(child).Printf("error: %s", err)
		if err != scope.TimedOut {
			delay := time.Duration(j.AttemptsMade+1) * BackoffDuration
			if time.Now().Add(delay).After(deadline) {
				delay = deadline.Sub(time.Now())
			}
			time.Sleep(delay)
		}
		return j.Fail(ctx, err.Error())
	}
	return j.Complete(ctx)
}

type JobLog struct {
	AttemptNumber int32
	HandlerID     string
	Success       bool
	FailureReason string
	Log           []byte
}
