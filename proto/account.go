package proto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"golang.org/x/crypto/poly1305"

	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"
	"euphoria.io/scope"
)

const (
	MinPasswordLength            = 6
	ClientKeyType                = security.AES128
	PasswordResetRequestLifetime = time.Hour
)

type OTP struct {
	URI       string
	Validated bool
}

func (o *OTP) QRImage(width, height int) (image.Image, error) {
	key, err := otp.NewKeyFromURL(o.URI)
	if err != nil {
		return nil, err
	}
	return key.Image(width, height)
}

func (o *OTP) Validate(password string) error {
	key, err := otp.NewKeyFromURL(o.URI)
	if err != nil {
		return err
	}
	if !totp.Validate(password, key.Secret()) {
		return ErrAccessDenied
	}
	return nil
}

type AccountManager interface {
	// GetAccount returns the account with the given ID.
	Get(ctx scope.Context, id snowflake.Snowflake) (Account, error)

	// RegisterAccount creates and returns a new, unverified account, along with
	// its (unencrypted) client key.
	Register(
		ctx scope.Context, kms security.KMS, namespace, id, password string,
		agentID string, agentKey *security.ManagedKey) (
		Account, *security.ManagedKey, error)

	// ResolveAccount returns any account registered under the given account identity.
	Resolve(ctx scope.Context, namespace, id string) (Account, error)

	// GrantStaff adds a StaffKMS capability to the identified account.
	GrantStaff(ctx scope.Context, accountID snowflake.Snowflake, kmsCred security.KMSCredential) error

	// RevokeStaff removes a StaffKMS capability from the identified account.
	RevokeStaff(ctx scope.Context, accountID snowflake.Snowflake) error

	// VerifyPersonalIdentity marks a personal identity as verified.
	VerifyPersonalIdentity(ctx scope.Context, namespace, id string) error

	// ChangeClientKey re-encrypts account keys with a new client key.
	// The correct former client key must also be given.
	ChangeClientKey(
		ctx scope.Context, accountID snowflake.Snowflake,
		oldClientKey, newClientKey *security.ManagedKey) error

	// RequestPasswordReset generates a temporary password-reset record.
	RequestPasswordReset(
		ctx scope.Context, kms security.KMS, namespace, id string) (Account, *PasswordResetRequest, error)

	// CheckPasswordResetRequest returns the account associated with
	// a password reset request, or an error if invalid or expired.
	GetPasswordResetAccount(ctx scope.Context, confirmation string) (Account, error)

	// ConfirmPasswordReset verifies a password reset confirmation code,
	// and applies the new password to the account referred to by the
	// confirmation code.
	ConfirmPasswordReset(ctx scope.Context, kms security.KMS, confirmation, password string) error

	// ChangeEmail changes an account's primary email address. It returns true if the email
	// address is verified for this account. If false is returned, then a verification
	// email will be sent out.
	ChangeEmail(ctx scope.Context, accountID snowflake.Snowflake, email string) (bool, error)

	// ChangeName changes an account's name.
	ChangeName(ctx scope.Context, accountID snowflake.Snowflake, name string) error

	// OTP unlocks and returns the user's enrolled OTP, or nil if one has never
	// been generated.
	OTP(ctx scope.Context, kms security.KMS, accountID snowflake.Snowflake) (*OTP, error)

	// GenerateOTP generates a new OTP secret for the user. If one has been generated
	// before, then it is replaced if it was never validated, or an error is returned.
	GenerateOTP(ctx scope.Context, heim *Heim, kms security.KMS, account Account) (*OTP, error)

	// ValidateOTP validates a one-time passcode according to the user's enrolled OTP.
	ValidateOTP(ctx scope.Context, kms security.KMS, accountID snowflake.Snowflake, passcode string) error
}

type PersonalIdentity interface {
	Namespace() string
	ID() string
	Verified() bool
}

func ValidatePersonalIdentity(namespace, id string) (bool, string) {
	switch namespace {
	case "email":
		return true, ""
	default:
		return false, fmt.Sprintf("invalid namespace: %s", namespace)
	}
}

func ValidateAccountPassword(password string) (bool, string) {
	if len(password) < MinPasswordLength {
		return false, fmt.Sprintf("password must be at least %d characters long", MinPasswordLength)
	}
	return true, ""
}

type Account interface {
	ID() snowflake.Snowflake
	Name() string
	Email() (string, bool)
	KeyFromPassword(password string) *security.ManagedKey
	KeyPair() security.ManagedKeyPair
	Unlock(clientKey *security.ManagedKey) (*security.ManagedKeyPair, error)
	IsStaff() bool
	UnlockStaffKMS(clientKey *security.ManagedKey) (security.KMS, error)
	PersonalIdentities() []PersonalIdentity
	UserKey() security.ManagedKey
	SystemKey() security.ManagedKey
	View(roomName string) *AccountView
}

// AccountView describes an account and its preferred names.
type AccountView struct {
	ID   snowflake.Snowflake `json:"id"`   // the id of the account
	Name string              `json:"name"` // the name that the holder of the account goes by
}

// PersonalAccountView describes an account to its owner.
type PersonalAccountView struct {
	AccountView
	Email string `json:"email"` // the account's email address
}

// NewAccountSecurity initializes the nonce and account secrets for a new account
// with the given password. Returns an encrypted key-encrypting-key, encrypted
// key-pair, nonce, and error.
func NewAccountSecurity(
	kms security.KMS, password string) (*AccountSecurity, *security.ManagedKey, error) {

	kpType := security.Curve25519

	// Use one KMS request to obtain all the randomness we need:
	//   - nonce
	//   - private key
	randomData, err := kms.GenerateNonce(kpType.NonceSize() + kpType.PrivateKeySize())
	if err != nil {
		return nil, nil, fmt.Errorf("rng error: %s", err)
	}
	randomReader := bytes.NewReader(randomData)

	// Generate nonce with random data. Use to populate IV.
	nonce := make([]byte, kpType.NonceSize())
	if _, err := io.ReadFull(randomReader, nonce); err != nil {
		return nil, nil, fmt.Errorf("rng error: %s", err)
	}
	iv := make([]byte, ClientKeyType.BlockSize())
	copy(iv, nonce)

	// Generate key-encrypting-key using KMS. This will be returned encrypted,
	// using the base64 encoding of the nonce as its context.
	nonceBase64 := base64.URLEncoding.EncodeToString(nonce)
	systemKey, err := kms.GenerateEncryptedKey(ClientKeyType, "nonce", nonceBase64)
	if err != nil {
		return nil, nil, fmt.Errorf("key generation error: %s", err)
	}

	// Generate private key using randomReader.
	keyPair, err := kpType.Generate(randomReader)
	if err != nil {
		return nil, nil, fmt.Errorf("keypair generation error: %s", err)
	}

	// Decrypt key-encrypting-key so we can encrypt keypair, and so we can re-encrypt
	// it using the user's key.
	kek := systemKey.Clone()
	if err = kms.DecryptKey(&kek); err != nil {
		return nil, nil, fmt.Errorf("key decryption error: %s", err)
	}

	// Encrypt private key.
	keyPair.IV = iv
	if err = keyPair.Encrypt(&kek); err != nil {
		return nil, nil, fmt.Errorf("keypair encryption error: %s", err)
	}

	// Clone key-encrypting-key and encrypt with client key.
	clientKey := security.KeyFromPasscode([]byte(password), nonce, ClientKeyType)
	userKey := kek.Clone()
	userKey.IV = iv
	if err := userKey.Encrypt(clientKey); err != nil {
		return nil, nil, fmt.Errorf("key encryption error: %s", err)
	}

	// Generate message authentication code, for verifying passwords.
	var (
		mac [16]byte
		key [32]byte
	)
	copy(key[:], clientKey.Plaintext)
	poly1305.Sum(&mac, nonce, &key)

	sec := &AccountSecurity{
		Nonce:     nonce,
		MAC:       mac[:],
		SystemKey: *systemKey,
		UserKey:   userKey,
		KeyPair:   *keyPair,
	}
	return sec, clientKey, nil
}

type AccountSecurity struct {
	Nonce     []byte
	MAC       []byte
	SystemKey security.ManagedKey
	UserKey   security.ManagedKey
	KeyPair   security.ManagedKeyPair
}

func (sec *AccountSecurity) unlock(clientKey *security.ManagedKey) (
	*security.ManagedKey, *security.ManagedKeyPair, error) {

	if clientKey.Encrypted() {
		return nil, nil, security.ErrKeyMustBeDecrypted
	}

	var (
		mac [16]byte
		key [32]byte
	)
	copy(mac[:], sec.MAC)
	copy(key[:], clientKey.Plaintext)
	if !poly1305.Verify(&mac, sec.Nonce, &key) {
		return nil, nil, ErrAccessDenied
	}

	kek := sec.UserKey.Clone()
	if err := kek.Decrypt(clientKey); err != nil {
		return nil, nil, err
	}

	kp := sec.KeyPair.Clone()
	if err := kp.Decrypt(&kek); err != nil {
		return nil, nil, err
	}

	return &kek, &kp, nil
}

func (sec *AccountSecurity) Unlock(clientKey *security.ManagedKey) (*security.ManagedKeyPair, error) {
	_, kp, err := sec.unlock(clientKey)
	return kp, err
}

func (sec *AccountSecurity) ChangeClientKey(oldKey, newKey *security.ManagedKey) error {
	if oldKey.Encrypted() || newKey.Encrypted() {
		return security.ErrKeyMustBeDecrypted
	}

	// Extract decrypted UserKey and verify correctness of oldKey.
	kek, _, err := sec.unlock(oldKey)
	if err != nil {
		return err
	}

	// Encrypt new UserKey.
	if err := kek.Encrypt(newKey); err != nil {
		return err
	}

	// Update MAC and encrypted UserKey.
	var (
		mac [16]byte
		key [32]byte
	)
	copy(key[:], newKey.Plaintext)
	poly1305.Sum(&mac, sec.Nonce, &key)
	sec.MAC = mac[:]
	sec.UserKey = *kek

	return nil
}

func (sec *AccountSecurity) ResetPassword(kms security.KMS, password string) (*AccountSecurity, error) {
	kek := sec.SystemKey.Clone()
	if err := kms.DecryptKey(&kek); err != nil {
		return nil, fmt.Errorf("key decryption error: %s", err)
	}
	kek.IV = make([]byte, ClientKeyType.BlockSize())
	copy(kek.IV, sec.Nonce)

	clientKey := security.KeyFromPasscode([]byte(password), sec.Nonce, sec.UserKey.KeyType)
	if err := kek.Encrypt(clientKey); err != nil {
		return nil, fmt.Errorf("key encryption error: %s", err)
	}

	var (
		mac [16]byte
		key [32]byte
	)
	copy(key[:], clientKey.Plaintext)
	poly1305.Sum(&mac, sec.Nonce, &key)

	nsec := &AccountSecurity{
		Nonce:     sec.Nonce,
		MAC:       mac[:],
		SystemKey: sec.SystemKey,
		UserKey:   kek,
		KeyPair:   sec.KeyPair,
	}
	return nsec, nil
}

type PasswordResetRequest struct {
	ID        snowflake.Snowflake
	AccountID snowflake.Snowflake
	Key       []byte
	Requested time.Time
	Expires   time.Time
}

func GeneratePasswordResetRequest(
	kms security.KMS, accountID snowflake.Snowflake) (*PasswordResetRequest, error) {

	id, err := snowflake.New()
	if err != nil {
		return nil, err
	}

	key, err := kms.GenerateNonce(sha256.BlockSize)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	req := &PasswordResetRequest{
		ID:        id,
		AccountID: accountID,
		Key:       key,
		Requested: now,
		Expires:   now.Add(PasswordResetRequestLifetime),
	}
	return req, nil
}

func (req *PasswordResetRequest) MAC() []byte {
	mac := hmac.New(sha256.New, req.Key)
	mac.Write([]byte(req.ID.String()))
	return mac.Sum(nil)
}

func (req *PasswordResetRequest) VerifyMAC(mac []byte) bool { return hmac.Equal(mac, req.MAC()) }

func (req *PasswordResetRequest) String() string {
	return fmt.Sprintf("%s-%s", req.ID, hex.EncodeToString(req.MAC()))
}

func ParsePasswordResetConfirmation(confirmation string) (snowflake.Snowflake, []byte, error) {
	var id snowflake.Snowflake

	idx := strings.IndexRune(confirmation, '-')
	if idx < 0 {
		return id, nil, ErrInvalidConfirmationCode
	}

	mac, err := hex.DecodeString(confirmation[idx+1:])
	if err != nil {
		return id, nil, ErrInvalidConfirmationCode
	}

	if err := id.FromString(confirmation[:idx]); err != nil {
		return id, nil, ErrInvalidConfirmationCode
	}

	return id, mac, nil
}
