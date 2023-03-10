package authentication

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/pquerna/otp/totp"
)

const (
	recoveryCodeCount = 6
)

var (
	ErrInvalidPasscode = errors.New("invalid passcode")
)

type TFA interface {
	Setup(ctx context.Context, email string) (secret string, qrCode bytes.Buffer, recoveryCodes []string, err error)
	Verify(ctx context.Context, user models.User, passcode string) error
}

type TFAImpl struct {
	issuer string
	logger log.Interface
}

func newTFA(issuer string, logger log.Interface) *TFAImpl {
	return &TFAImpl{
		logger: logger,
		issuer: issuer,
	}
}

// Generate implements Interface
func (i *TFAImpl) Setup(ctx context.Context, email string) (string, bytes.Buffer, []string, error) {

	secret, qr, recoveryCodes, err := i.generate(email)
	if err != nil {
		return "", bytes.Buffer{}, nil, err // descriptive enough, no need to wrap
	}

	return secret, qr, recoveryCodes, err
}

func (i *TFAImpl) generate(email string) (string, bytes.Buffer, []string, error) {
	key, err := totp.Generate(
		totp.GenerateOpts{
			Issuer:      i.issuer,
			AccountName: email,
		},
	)
	if err != nil {
		return "", bytes.Buffer{}, nil, fmt.Errorf("error generating totp: %w", err)
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", bytes.Buffer{}, nil, fmt.Errorf("error generating qr code: %w", err)
	}

	var qr bytes.Buffer
	if err := png.Encode(&qr, img); err != nil {
		return "", bytes.Buffer{}, nil, fmt.Errorf("error png-encoding qr code: %w", err)
	}

	recoveryCodes := make([]string, recoveryCodeCount)
	now := time.Now()
	for idx := 0; idx < len(recoveryCodes); idx++ {
		code, err := totp.GenerateCode(key.Secret(), now)
		if err != nil {
			return "", bytes.Buffer{}, nil, fmt.Errorf("error generating recovery codes: %w", err)
		}
		recoveryCodes[idx] = code
	}
	return key.Secret(), qr, recoveryCodes, nil
}

// Verify implements Interface
func (i *TFAImpl) Verify(ctx context.Context, user models.User , passcode string) error {

	if !totp.Validate(passcode, user.TFASecret()) {
		return ErrInvalidPasscode
	}

	return nil
}

var _ TFA = (*TFAImpl)(nil)
