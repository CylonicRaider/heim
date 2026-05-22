package psql

import (
	"database/sql"
	"time"

	"gopkg.in/gorp.v1"
)

type BannedAgent struct {
	AgentID       string         `db:"agent_id"`
	Room          sql.NullString `db:"room"`
	Created       time.Time      `db:"created"`
	Expires       gorp.NullTime  `db:"expires"`
	RoomReason    string         `db:"room_reason"`
	AgentReason   string         `db:"agent_reason"`
	PrivateReason string         `db:"private_reason"`
}

type BannedIP struct {
	IP      string         `db:"ip"`
	Room    sql.NullString `db:"room"`
	Created time.Time      `db:"created"`
	Expires gorp.NullTime  `db:"expires"`
	Reason  string         `db:"reason"`
}
