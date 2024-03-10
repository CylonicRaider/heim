package worker

import (
	"github.com/euphoria-io/scope"

	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
)

func Loop(ctx scope.Context, heim *proto.Heim, workerName, queueName string) error {
	logging.Logger(ctx).Printf("Loop\n")
	ctrl, err := NewController(ctx, heim, workerName, queueName)
	if err != nil {
		logging.Logger(ctx).Printf("error: %s\n", err)
		return err
	}

	ctx.WaitGroup().Add(1)
	go ctrl.background(ctx)
	ctx.WaitGroup().Wait()
	return ctx.Err()
}
