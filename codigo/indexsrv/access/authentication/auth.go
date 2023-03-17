package authentication

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	tfaIssuer          = "mifs_is"
	sessionTTL         = 12 * time.Hour
	sessionTokenLength = 50
)

var (
	ErrUserInfoMismatch = errors.New("user name or email don't match the supplied id")
	ErrUserExists       = errors.New("user already exists")
)

// UserManager defines the interface for a facade that allows the rest of the application to manage users
type UserManager interface {
	Signup(ctx context.Context, name string, email string, password string) (models.User, error)
	Login(ctx context.Context, email string, password string, passCode string) (token string, err error)
	Logout(ctx context.Context, token string) error
	GetByID(ctx context.Context, id string) (models.User, error)
	Setup2FA(ctx context.Context, userID string) (qr bytes.Buffer, codes []string, err error)
	GetSession(ctx context.Context, token string) (models.Session, error)
}

// UserManagerImpl implements the UserManagerFacade by means of a UserRepository
type UserManagerImpl struct {
	passwordHasher PasswordHasher
	users          repository.UserRepository
	twofa          TFA
	sessions       SessionManager
}

// CheckSession implements UserManager
func (m *UserManagerImpl) GetSession(ctx context.Context, token string) (models.Session, error) {
	return m.sessions.LookUp(ctx, token)
}

// NewUserManager constructs a new user manager
func NewUserManager(users repository.UserRepository, sessions repository.SessionRepository, logger log.Interface) *UserManagerImpl {
	return &UserManagerImpl{
		passwordHasher: &BCryptHasher{},
		users:          users,
		twofa:          newTFA(tfaIssuer, logger),
		sessions:       newSessionManager(sessions, sessionTTL, sessionTokenLength),
	}
}

// Login implements UserManager
func (m *UserManagerImpl) Login(ctx context.Context, email string, password string, tfaPasscode string) (string, error) {

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

	var tfaOK bool
	if tfaPasscode != "" {
		if err := m.twofa.Verify(ctx, user, tfaPasscode); err != nil {
			return "", ErrInvalid2FAPasscode
		}
		tfaOK = true
	}

	return m.sessions.Create(ctx, user.ID(), tfaOK)
}

// Logout implements UserManager
func (m *UserManagerImpl) Logout(ctx context.Context, token string) error {
	return m.sessions.Revoke(ctx, token)
}

// Setup2FA implements UserManager
func (m *UserManagerImpl) Setup2FA(ctx context.Context, userID string) (bytes.Buffer, []string, error) {
	user, err := m.users.Get(ctx, userID)
	if err != nil {
		return bytes.Buffer{}, nil, fmt.Errorf("failed to fetch user from repo: %w", err)
	}

	secret, qrCode, recoveryCodes, err := m.twofa.Setup(ctx, user.Email())
	if err != nil {
		return bytes.Buffer{}, nil, err
	}

	if err := m.users.Update2FA(ctx, user.ID(), secret); err != nil {
		return bytes.Buffer{}, nil, fmt.Errorf("error updating 2fa secret in database: %w", err)
	}

	return qrCode, recoveryCodes, nil
}

// CreateOrUpdate creates a user if it doesn't exist.
// If it does, it checks that the main fields match, and updates the tokens
func (m *UserManagerImpl) Signup(ctx context.Context, name string, email string, rawPassword string) (models.User, error) {
	_, err := m.users.GetByEmail(ctx, email)
	if !errors.Is(err, repository.ErrNotFound) {
		if err != nil {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("error checking user existance in db: %w", err)
	}

	passwordHash, err := m.passwordHasher.Hash(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	created, err := m.users.Add(ctx, name, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("error creating user in db: %w", err)
	}

	return created, nil

}

// GetByID looks up a user by it's id
func (m *UserManagerImpl) GetByID(ctx context.Context, id string) (models.User, error) {
	return m.users.Get(ctx, id)
}

var _ UserManager = (*UserManagerImpl)(nil)
