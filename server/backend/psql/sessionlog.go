package psql

import "time"

type SessionLog struct {
	SessionID string    `db:"session_id"`
	IP        string    `db:"ip"`
	Room      string    `db:"room"`
	UserAgent string    `db:"user_agent"`
	Connected time.Time `db:"connected"`
}
