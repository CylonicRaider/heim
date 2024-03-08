package proto

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/euphoria-io/scope"
	"golang.org/x/crypto/poly1305"

	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"
)

const (
	RoomManagerKeyType = security.AES128
	RoomMessageKeyType = security.AES128
)

type PrivilegeLevel byte

const (
	General PrivilegeLevel = iota
	Host
	Staff
)

// A Listing is a sortable list of Identitys present in a Room.
// TODO: these should be Sessions
type Listing []SessionView

func (l Listing) Len() int      { return len(l) }
func (l Listing) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func (l Listing) Less(i, j int) bool {
	if l[i].Name == l[j].Name {
		if l[i].ID == l[j].ID {
			return l[i].SessionID < l[j].SessionID
		}
		return l[i].ID < l[j].ID
	}
	return l[i].Name < l[j].Name
}

// A Room is a nexus of communication. Users connect to a Room via
// Session and interact.
type Room interface {
	ID() string
	Title() string
	GetMessage(scope.Context, snowflake.Snowflake) (*Message, error)
	Latest(scope.Context, int, snowflake.Snowflake) ([]Message, error)
	Snapshot(ctx scope.Context, session Session, level PrivilegeLevel, numMessages int) (*SnapshotEvent, error)

	// Join inserts a Session into the Room's global presence.
	Join(scope.Context, Session) (virtualClientAddr string, err error)

	// Part removes a Session from the Room's global presence.
	Part(scope.Context, Session) error

	// IsValidParent checks whether the message with the given ID is able to be replied to.
	IsValidParent(id snowflake.Snowflake) (bool, error)

	// Send broadcasts a Message from a Session to the Room.
	Send(scope.Context, Session, Message) (Message, error)

	// Edit modifies or deletes a message.
	EditMessage(scope.Context, Session, EditMessageCommand) (EditMessageReply, error)

	// Listing returns the current global list of connected sessions to this
	// Room.
	Listing(ctx scope.Context, level PrivilegeLevel, exclude ...Session) (Listing, error)

	// RenameUser updates the nickname of a Session in this Room.
	RenameUser(ctx scope.Context, session Session, formerName string) (*NickEvent, error)

	// Version returns the version of the server hosting this Room.
	Version() string

	WaitForPart(sessionID string) error

	MessageKeyID(scope.Context) (string, bool, error)

	ResolveClientAddress(ctx scope.Context, addr string) (net.IP, error)

	ResolveNick(ctx scope.Context, userID UserID) (string, bool, error)
}

type ManagedRoom interface {
	Room

	// Ban adds an entry to the room's ban list. A zero value for until
	// indicates a permanent ban.
	Ban(ctx scope.Context, ban Ban, until time.Time) error

	// UnbanAgent removes an agent ban from the room.
	Unban(ctx scope.Context, ban Ban) error

	// GenerateMessageKey generates and stores a new key and nonce
	// for encrypting messages in the room. This invalidates all grants made with
	// the previous key.
	GenerateMessageKey(ctx scope.Context, kms security.KMS) (RoomMessageKey, error)

	// MessageKey returns the room's current message key, or nil if the room is
	// unencrypted.
	MessageKey(ctx scope.Context) (RoomMessageKey, error)

	// ManagerKey returns a handle to the room's manager key.
	ManagerKey(ctx scope.Context) (RoomManagerKey, error)

	// Managers returns the list of accounts managing the room.
	Managers(ctx scope.Context) ([]Account, error)

	// AddManager adds an account as a manager of the room. An unencrypted
	// client key and corresponding account are needed from the user taking
	// this action.
	AddManager(
		ctx scope.Context, kms security.KMS,
		actor Account, actorKey *security.ManagedKey,
		newManager Account) error

	// RemoveManager removes an account as manager. An unencrypted client key
	// and corresponding account are needed from the user taking this action.
	RemoveManager(
		ctx scope.Context, actor Account, actorKey *security.ManagedKey, formerManager Account) error

	// ManagerCapability returns the manager capablity for the given account.
	ManagerCapability(ctx scope.Context, manager Account) (security.Capability, error)

	MinAgentAge() time.Duration
}

type RoomMessageKey interface {
	AccountGrantable
	PasscodeGrantable

	// ID returns a unique identifier for the key.
	KeyID() string

	// Timestamp returns when the key was generated.
	Timestamp() time.Time

	// Nonce returns the current 128-bit nonce for the room.
	Nonce() []byte

	// ManagedKey returns the current encrypted ManagedKey for the room.
	ManagedKey() security.ManagedKey
}

type RoomManagerKey interface {
	AccountGrantable

	// KeyPair returns the current encrypted ManagedKeyPair for the room.
	KeyPair() security.ManagedKeyPair

	// StaffUnlock uses the KMS to decrypt the room's ManagedKeyPair and returns it.
	StaffUnlock(kms security.KMS) (*security.ManagedKeyPair, error)

	// Unlock decrypts the room's ManagedKeyPair with the given key and returns it.
	Unlock(managerKey *security.ManagedKey) (*security.ManagedKeyPair, error)

	// Nonce returns the current 128-bit nonce for the room.
	Nonce() []byte
}

func NewRoomSecurity(kms security.KMS, roomName string) (*RoomSecurity, error) {
	kpType := security.Curve25519

	// Use one KMS request to obtain all the randomness we need:
	//   - key-encrypting-key IV
	//   - private key for grants to accounts
	//   - nonce for manager grants to accounts
	randomData, err := kms.GenerateNonce(
		RoomManagerKeyType.BlockSize() + kpType.PrivateKeySize() + kpType.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("rng error: %s", err)
	}
	randomReader := bytes.NewReader(randomData)

	// Generate IV with random data.
	iv := make([]byte, RoomManagerKeyType.BlockSize())
	if _, err := io.ReadFull(randomReader, iv); err != nil {
		return nil, fmt.Errorf("rng error: %s", err)
	}

	// Generate private key using randomReader.
	keyPair, err := kpType.Generate(randomReader)
	if err != nil {
		return nil, fmt.Errorf("keypair generation error: %s", err)
	}

	// Generate nonce with random data.
	nonce := make([]byte, kpType.NonceSize())
	if _, err := io.ReadFull(randomReader, nonce); err != nil {
		return nil, fmt.Errorf("rng error: %s", err)
	}

	// Generate key-encrypting-key. This will be returned encrypted, using the
	// name of the room as its context.
	encryptedKek, err := kms.GenerateEncryptedKey(RoomManagerKeyType, "room", roomName)
	if err != nil {
		return nil, fmt.Errorf("key generation error: %s", err)
	}

	// Decrypt key-encrypting-key so we can encrypt keypair.
	kek := encryptedKek.Clone()
	if err = kms.DecryptKey(&kek); err != nil {
		return nil, fmt.Errorf("key decryption error: %s", err)
	}

	// Encrypt private key.
	keyPair.IV = iv
	if err = keyPair.Encrypt(&kek); err != nil {
		return nil, fmt.Errorf("keypair encryption error: %s", err)
	}

	// Generate message authentication code, for verifying a given key-encryption-key.
	var (
		mac [16]byte
		key [32]byte
	)
	copy(key[:], kek.Plaintext)
	poly1305.Sum(&mac, iv, &key)

	sec := &RoomSecurity{
		Nonce:            nonce,
		MAC:              mac[:],
		KeyEncryptingKey: *encryptedKek,
		KeyPair:          *keyPair,
	}
	return sec, nil
}

type RoomSecurity struct {
	Nonce            []byte
	MAC              []byte
	KeyEncryptingKey security.ManagedKey
	KeyPair          security.ManagedKeyPair
}

func (sec *RoomSecurity) Unlock(managerKey *security.ManagedKey) (*security.ManagedKeyPair, error) {
	if managerKey.Encrypted() {
		return nil, security.ErrKeyMustBeDecrypted
	}

	var (
		mac [16]byte
		key [32]byte
	)
	copy(mac[:], sec.MAC)
	copy(key[:], managerKey.Plaintext)
	if !poly1305.Verify(&mac, sec.KeyPair.IV, &key) {
		return nil, ErrAccessDenied
	}

	kp := sec.KeyPair.Clone()
	if err := kp.Decrypt(managerKey); err != nil {
		return nil, err
	}

	return &kp, nil
}
