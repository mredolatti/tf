package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoSuchSession = errors.New("no session for specified token")
)

type SessionManager interface {
	Create(ctx context.Context, email string, password string) (string, error)
	LookUp(ctx context.Context, token string) (models.Session, error)
	Revoke(ctx context.Context, token string) error
}

type SessionManagerImpl struct {
	users      repository.UserRepository
	sessions   repository.SessionRepository
	ttl        time.Duration
	keyFactory SessionKeyFactory
	skeyLen    int
}

func NewSessionManager(
	repo repository.SessionRepository,
	users repository.UserRepository,
	ttl time.Duration,
	sessionKeyLength int,
) *SessionManagerImpl {
	return &SessionManagerImpl{
		users:      users,
		sessions:   repo,
		ttl:        ttl,
		keyFactory: &CryptoBase64Generator{},
		skeyLen:    sessionKeyLength,
	}
}

// Create implements Manager
func (m *SessionManagerImpl) Create(ctx context.Context, email string, password string) (string, error) {
	user, err := m.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to retrieve user with email '%s': %s", email, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("error verifying user credentials: %w", err)
	}

	token, err := m.keyFactory.Generate(m.skeyLen)
	if err != nil {
		return "", fmt.Errorf("error generating session token: %w", err)
	}

	if err := m.sessions.Put(ctx, token, user.ID(), m.ttl); err != nil {
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
