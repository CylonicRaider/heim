package console

import (
	"fmt"
	"time"

	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto"
)

func init() {
	register("ban", "Ban an agent or IP.", &ban{})
	register("unban", "Unban an agent or IP.", &unban{})
}

type ban struct {
	handlerBase
	Room     string                `usage:"Ban only in the given room."`
	Duration console.DurationValue `usage:"Duration of ban. (default: forever)"`
	Agent    string                `usage:"Agent ID to ban."`
	IP       string                `usage:"IP to ban." cli:"ip"`
}

func (b *ban) Run(env console.CLIEnv) error {
	ctx := env.Context()

	var until time.Time
	var untilStr string
	switch b.Duration.Duration {
	case 0:
		until = time.Time{}
		untilStr = "forever"
	default:
		until = time.Now().Add(b.Duration.Duration)
		untilStr = fmt.Sprintf("until %s", until)
	}

	ban := proto.Ban{}
	switch {
	case b.Agent != "":
		ban.ID = proto.UserID(b.Agent)
	case b.IP != "":
		ban.IP = b.IP
	default:
		return fmt.Errorf("-agent <agent-id> or -ip <ip> is required")
	}

	if b.Room == "" {
		if err := b.cli.backend.Ban(ctx, ban, until); err != nil {
			return err
		}
		env.Printf("banned globally for %s: %#v\n", untilStr, ban)
	} else {
		room, err := b.cli.backend.GetRoom(ctx, b.Room)
		if err != nil {
			return err
		}
		if err := room.Ban(ctx, ban, until); err != nil {
			return err
		}
		env.Printf("banned in room %s for %s: %#v\n", b.Room, untilStr, ban)
	}

	return nil
}

type unban struct {
	handlerBase
	Room  string `usage:"Unban only in the given room."`
	Agent string `usage:"Agent ID to unban."`
	IP    string `usage:"IP to unban." cli:"ip"`
}

func (u *unban) Run(env console.CLIEnv) error {
	ctx := env.Context()

	ban := proto.Ban{}
	switch {
	case u.Agent != "":
		ban.ID = proto.UserID(u.Agent)
	case u.IP != "":
		ban.IP = u.IP
	default:
		return fmt.Errorf("-agent <agent-id> or -ip <ip> is required")
	}

	if u.Room == "" {
		if err := u.cli.backend.Unban(ctx, ban); err != nil {
			return err
		}
		env.Printf("global unban: %#v\n", ban)
	} else {
		room, err := u.cli.backend.GetRoom(ctx, u.Room)
		if err != nil {
			return err
		}
		if err := room.Unban(ctx, ban); err != nil {
			return err
		}
		env.Printf("unban in room %s: %#v\n", u.Room, ban)
	}

	return nil
}
