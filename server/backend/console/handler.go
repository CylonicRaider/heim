package console

import (
	"fmt"
	"strings"

	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"
	"github.com/euphoria-io/scope"
)

type cli struct {
	*console.CLI

	ctrl    *Controller
	backend proto.Backend
	kms     security.KMS
}

func newCLI(ctrl *Controller) *cli {
	result := &cli{
		CLI:     console.NewCLI("> ", true),
		ctrl:    ctrl,
		backend: ctrl.backend,
		kms:     ctrl.kms,
	}
	for name, reg := range registry {
		boundHnd := console.CloneParams(reg.template).(handler)
		boundHnd.Bind(result)
		result.AddNewCommand(name, reg.description, boundHnd)
	}
	result.CLI.Callback = result.cliCallback
	return result
}

func (c *cli) cliCallback(cli *console.CLI, done bool, argv []string, err error) error {
	if !done {
		ctx := cli.Console.(console.ConsoleWithContext).Context()
		logging.Logger(ctx).Printf("> %v\n", argv)
	}
	return err
}

func (c *cli) Session() proto.Session { return (*consoleSession)(c) }

func (c *cli) resolveAccount(ctx scope.Context, ref string) (proto.Account, error) {
	idx := strings.IndexRune(ref, ':')
	if idx < 0 {
		var accountID snowflake.Snowflake
		if err := accountID.FromString(ref); err != nil {
			return nil, err
		}
		return c.backend.AccountManager().Get(ctx, accountID)
	}
	return c.backend.AccountManager().Resolve(ctx, ref[:idx], ref[idx+1:])
}

type consoleSession cli

// TODO: log details about the client
func (c *consoleSession) Identity() proto.Identity { return (*consoleIdentity)(c) }
func (c *consoleSession) ID() string               { return "console" }
func (c *consoleSession) AgentID() string          { return "!console!" }
func (c *consoleSession) ServerID() string         { return "" }
func (c *consoleSession) SetName(name string)      {}
func (c *consoleSession) Close()                   {}
func (c *consoleSession) CheckAbandoned() error    { return nil }

func (c *consoleSession) View(level proto.PrivilegeLevel) proto.SessionView {
	return proto.SessionView{
		IdentityView: c.Identity().View(),
		SessionID:    "console",
	}
}

func (c *consoleSession) Send(scope.Context, proto.PacketType, interface{}) error {
	return fmt.Errorf("not implemented")
}

type consoleIdentity cli

func (c *consoleIdentity) ID() proto.UserID { return proto.UserID("console") }
func (c *consoleIdentity) Name() string     { return "console" }
func (c *consoleIdentity) ServerID() string { return "" }

func (c *consoleIdentity) View() proto.IdentityView {
	return proto.IdentityView{ID: "console", Name: "console"}
}

type handler interface {
	console.CommandParams
	Bind(parent *cli)
}

type handlerBase struct {
	cli *cli
}

func (h *handlerBase) Bind(parent *cli) {
	h.cli = parent
}

type registeredHandler struct {
	description string
	template    handler
}

var registry = map[string]registeredHandler{}

func register(name, description string, h handler) {
	registry[name] = registeredHandler{description, h}
}
