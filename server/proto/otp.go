package proto

import (
	"image"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type OTP struct {
	URI       string
	Validated bool
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
