package psql

type MessageKey struct {
	ID           string
	EncryptedKey ByteANonNull `db:"encrypted_key"`
	IV           ByteAOrNull
	Nonce        ByteANonNull
}

type Capability struct {
	ID                   string
	AccountID            string       `db:"account_id"`
	NonceBytes           ByteAOrNull  `db:"nonce"`
	EncryptedPrivateData ByteANonNull `db:"encrypted_private_data"`
	PublicData           ByteANonNull `db:"public_data"`
}

func (c *Capability) CapabilityID() string     { return c.ID }
func (c *Capability) Nonce() []byte            { return c.NonceBytes.v }
func (c *Capability) PublicPayload() []byte    { return c.PublicData.v }
func (c *Capability) EncryptedPayload() []byte { return c.EncryptedPrivateData.v }
