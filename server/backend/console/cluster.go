package console

import "euphoria.leet.nu/heim/console"

func init() {
	register("peers", "List the current cluster peers.", &peers{})
}

type peers struct {
	handlerBase
}

func (p *peers) Run(env console.CLIEnv) error {
	for i, peer := range p.cli.backend.Peers() {
		env.Printf("%d. %s: version=%s, era=%s\n", i+1, peer.ID, peer.Version, peer.Era)
	}
	return nil
}
