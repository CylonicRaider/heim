package proto

import (
	"time"

	"euphoria.leet.nu/heim/proto/snowflake"
)

type Presence struct {
	SessionView
	LastInteracted time.Time           `json:"last_interacted"`
	MessageID      snowflake.Snowflake `json:"message_id"`
	Typing         bool                `json:"typing"`
}
