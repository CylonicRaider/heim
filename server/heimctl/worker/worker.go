package worker

import (
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/jobs"
	"github.com/euphoria-io/scope"
)

type Worker interface {
	Init(heim *proto.Heim) error
	QueueName() string
	JobType() jobs.JobType
	Work(ctx scope.Context, job *jobs.Job, payload interface{}) error
}
