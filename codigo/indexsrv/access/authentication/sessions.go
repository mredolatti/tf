package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalid2FAPasscode = errors.New("invalid 2fa passcode")
	ErrNoSuchSession      = errors.New("no session for specified token")
)

type SessionManager interface {
	Create(ctx context.Context, userID string, tfaOK bool) (string, error)
	LookUp(ctx context.Context, token string) (models.Session, error)
	Revoke(ctx context.Context, token string) error
}

type SessionManagerImpl struct {
	sessions   repository.SessionRepository
	ttl        time.Duration
	keyFactory SessionKeyFactory
	skeyLen    int
}

func newSessionManager(
	repo repository.SessionRepository,
	ttl time.Duration,
	sessionKeyLength int,
) *SessionManagerImpl {
	return &SessionManagerImpl{
		sessions:   repo,
		ttl:        ttl,
		keyFactory: &CryptoBase64Generator{},
		skeyLen:    sessionKeyLength,
	}
}

// Create implements Manager
func (m *SessionManagerImpl) Create(ctx context.Context, userID string, tfaOK bool) (string, error) {

	token, err := m.keyFactory.Generate(m.skeyLen)
	if err != nil {
		return "", fmt.Errorf("error generating session token: %w", err)
	}

	if err := m.sessions.Put(ctx, token, userID, tfaOK, m.ttl); err != nil {
		return "", fmt.Errorf("error persisting session data: %w", err)
	}

	return token, nil
}

// LookUp implements Manager
func (m *SessionManagerImpl) LookUp(ctx context.Context, token string) (models.Session, error) {
	sess, err := m.sessions.Get(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNoSuchSession
		}
		return nil, fmt.Errorf("error retrieving session data: %w", err)
	}

	return sess, nil
}

// Revoke implements Manager
func (m *SessionManagerImpl) Revoke(ctx context.Context, token string) error {
	if err := m.sessions.Remove(ctx, token); err != nil {
		return fmt.Errorf("error removing session data: %w", err)
	}
	return nil
}

var _ SessionManager = (*SessionManagerImpl)(nil)
