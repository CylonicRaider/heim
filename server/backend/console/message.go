package console

import (
	"fmt"
	"strings"

	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/snowflake"
	"github.com/euphoria-io/scope"
)

func init() {
	register("delete-message", "Delete messages.", &deleteMessage{})
	register("undelete-message", "Reverse message deletion.", &undeleteMessage{})
}

type deleteMessage struct {
	handlerBase
	Quiet    bool     `usage:"Suppress edit-message-event broadcast."`
	Messages []string "usage:\"`room:message-id` pairs to delete.\" cli:\"message,arg,required\""
}

func (d *deleteMessage) Run(env console.CLIEnv) error {
	return setDeleted(env.Context(), env, d.cli, d.Quiet, d.Messages, true)
}

type undeleteMessage struct {
	handlerBase
	Quiet    bool     `usage:"Suppress edit-message-event broadcast."`
	Messages []string "usage:\"`room:message-id` pairs to undelete.\" cli:\"message,arg,required\""
}

func (u *undeleteMessage) Run(env console.CLIEnv) error {
	return setDeleted(env.Context(), env, u.cli, u.Quiet, u.Messages, false)
}

func parseDeleteMessageArg(arg string) (string, snowflake.Snowflake, error) {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("format should be <room>:<message-id>")
	}

	var msgID snowflake.Snowflake
	if err := msgID.FromString(parts[1]); err != nil {
		return "", 0, err
	}

	return parts[0], msgID, nil
}

func setDeleted(ctx scope.Context, env console.CLIEnv, c *cli, quiet bool, messages []string, deleted bool) error {
	for _, arg := range messages {
		roomName, msgID, err := parseDeleteMessageArg(arg)
		if err != nil {
			return err
		}

		room, err := c.backend.GetRoom(ctx, roomName)
		if err != nil {
			return fmt.Errorf("%s: %s", arg, err)
		}

		msg, err := room.GetMessage(ctx, msgID)
		if err != nil {
			return fmt.Errorf("%s: %s", arg, err)
		}

		var action string
		if deleted {
			action = "Deleting"
		} else {
			action = "Undeleting"
		}
		env.Printf("%s message %s in room %s... ", action, msgID.String(), roomName)
		edit := proto.EditMessageCommand{
			ID:             msgID,
			PreviousEditID: msg.PreviousEditID,
			Delete:         deleted,
			Announce:       !quiet,
		}
		if _, err := room.EditMessage(ctx, c.Session(), edit); err != nil {
			return fmt.Errorf("%s: %s", arg, err)
		}
		env.Printf("OK!\n")
	}
	return nil
}
