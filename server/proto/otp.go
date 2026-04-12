package proto

import (
	"image"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
)

const (
	skew = 1
)

type OTPInfo interface {
	// GetURI returns the otpauth URI for enrollment into password managers.
	GetURI() string

	// QRImage returns a QR code of the enrollment URI.
	QRImage(width, height int) (image.Image, error)

	// HasBeenValidated returns whether the user has ever successfully submitted
	// an OTP for this OTP object. This *cannot* be used to check whether the
	// user has recently submitted a fresh one-time password.
	HasBeenValidated() bool
}

type OTP struct {
	URI           string
	LastValidated uint64
}

func NewOTP(issuer, name string) (*OTP, error) {
	opts := totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: name,
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return nil, err
	}
	return &OTP{URI: key.String()}, nil
}

func (o *OTP) GetURI() string         { return o.URI }
func (o *OTP) HasBeenValidated() bool { return o.LastValidated != 0 }

func (o *OTP) QRImage(width, height int) (image.Image, error) {
	key, err := otp.NewKeyFromURL(o.URI)
	if err != nil {
		return nil, err
	}
	return key.Image(width, height)
}

// DANGER: The LastValidated field of the receiver must be stored to persistent
// storage after calling to prevent replay attacks.
func (o *OTP) Validate(password string) error {
	key, err := otp.NewKeyFromURL(o.URI)
	if err != nil {
		return err
	}

	// The totp package does not provide a simple and correct way of preventing
	// replay attacks, and instead suggests that downstream developers implement
	// that themselves ("its [sic] just a simple lookup if a TOTP has been used
	// within the last few minutes"). Storing invalidated OTPs would require
	// either nontrivial garbage collection or a persistent audit log (which is
	// probably a good idea). We choose the more simple approach of enforcing
	// a monotonically increasing underlying counter. Fortunately, grabbing and
	// adapting the relevant totp code is easy because it is mostly a wrapper
	// around hotp.

	opts := hotp.ValidateOpts{
		Digits:    key.Digits(),
		Algorithm: key.Algorithm(),
	}
	period := int64(key.Period())
	if period == 0 {
		period = 30
	}

	now := uint64(time.Now().UTC().Unix() / period)
	counts := []uint64{now}
	for i := 0; i < skew; i++ {
		counts = append(counts, now+1, now-1)
	}

	for _, c := range counts {
		if c <= o.LastValidated {
			continue
		}
		rv, err := hotp.ValidateCustom(password, c, key.Secret(), opts)
		if err != nil {
			return ErrAccessDenied
		}
		if rv {
			o.LastValidated = c
			return nil
		}
	}

	return ErrAccessDenied
}
