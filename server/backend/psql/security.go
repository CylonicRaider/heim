package psql

type MessageKey struct {
	ID           string       `db:"id"`
	EncryptedKey ByteANonNull `db:"encrypted_key"`
	IV           ByteAOrNull  `db:"iv"`
	Nonce        ByteANonNull `db:"nonce"`
}

type Capability struct {
	ID                   string       `db:"id"`
	AccountID            string       `db:"account_id"`
	NonceBytes           ByteAOrNull  `db:"nonce"`
	EncryptedPrivateData ByteANonNull `db:"encrypted_private_data"`
	PublicData           ByteANonNull `db:"public_data"`
}

func (c *Capability) CapabilityID() string     { return c.ID }
func (c *Capability) Nonce() []byte            { return c.NonceBytes.v }
func (c *Capability) PublicPayload() []byte    { return c.PublicData.v }
func (c *Capability) EncryptedPayload() []byte { return c.EncryptedPrivateData.v }
