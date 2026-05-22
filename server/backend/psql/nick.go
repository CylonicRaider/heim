package psql

type Nick struct {
	Room   string `db:"room"`
	UserID string `db:"user_id"`
	Nick   string `db:"nick"`
}
