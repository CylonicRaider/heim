package console

import (
	"fmt"

	"euphoria.leet.nu/heim/console"
)

func init() {
	register("shutdown", "Shut down this backend.", &shutdown{})
}

type shutdown struct {
	handlerBase
}

func (s *shutdown) Run(env console.CLIEnv) error {
	s.cli.ctrl.ctx.Terminate(fmt.Errorf("shutdown initiated from console"))
	return nil
}
