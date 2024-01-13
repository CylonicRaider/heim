package worker

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/jobs"
	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/snowflake"
	"github.com/euphoria-io/scope"
)

const (
	PollTime      = 2 * time.Second
	StatsInterval = 10 * time.Second
	StealChance   = 0.25
)

var workers = map[string]Worker{}

func register(w Worker) { workers[w.QueueName()] = w }

func NewController(ctx scope.Context, heim *proto.Heim, workerName, queueName string) (*Controller, error) {
	jq, err := heim.Backend.Jobs().GetQueue(ctx, queueName)
	if err != nil {
		return nil, err
	}

	worker, ok := workers[queueName]
	if !ok {
		return nil, fmt.Errorf("no worker registered for queue %s", queueName)
	}

	sf, err := snowflake.New()
	if err != nil {
		return nil, err
	}

	if err := worker.Init(heim); err != nil {
		return nil, err
	}

	ctrl := &Controller{
		id: fmt.Sprintf("%s-%s", workerName, sf),
		jq: jq,
		w:  worker,
	}
	return ctrl, nil
}

type Controller struct {
	id string
	jq jobs.JobQueue
	w  Worker
}

func (c *Controller) background(ctx scope.Context) {
	defer ctx.WaitGroup().Done()

	var lastStatCheck time.Time
	for {
		logging.Logger(ctx).Printf("[%s] background loop", c.w.QueueName())
		if time.Now().Sub(lastStatCheck) > StatsInterval {
			logging.Logger(ctx).Printf("[%s] collecting stats", c.w.QueueName())
			stats, err := c.jq.Stats(ctx)
			if err != nil {
				logging.Logger(ctx).Printf("error: %s stats: %s", c.w.QueueName(), err)
				return
			}
			lastStatCheck = time.Now()
			labels := map[string]string{"queue": c.w.QueueName()}
			claimedGauge.With(labels).Set(float64(stats.Claimed))
			dueGauge.With(labels).Set(float64(stats.Due))
			waitingGauge.With(labels).Set(float64(stats.Waiting))
		}
		if err := c.processOne(ctx); err != nil {
			// TODO: retry a couple times before giving up
			logging.Logger(ctx).Printf("error: %s: %s", c.w.QueueName(), err)
			errorCounter.Inc()
			return
		}
	}
}

func (c *Controller) processOne(ctx scope.Context) error {
	job, err := c.claimOrSteal(ctx.ForkWithTimeout(StatsInterval))
	if err != nil {
		if err == scope.TimedOut {
			return nil
		}
		return err
	}

	if job.Type != jobs.EmailJobType {
		return jobs.ErrInvalidJobType
	}

	payload, err := job.Payload()
	if err != nil {
		return err
	}

	return job.Exec(ctx, func(ctx scope.Context) error {
		labels := prometheus.Labels{"queue": c.jq.Name()}
		defer processedCounter.With(labels).Inc()
		if err := c.w.Work(ctx, job, payload); err != nil {
			failedCounter.With(labels).Inc()
			return err
		}
		completedCounter.With(labels).Inc()
		return nil
	})
}

func (c *Controller) claimOrSteal(ctx scope.Context) (*jobs.Job, error) {
	logging.Logger(ctx).Printf("[%s] attempting to claim", c.w.QueueName())
	return jobs.Claim(ctx, c.jq, c.id, PollTime, StealChance)
}
