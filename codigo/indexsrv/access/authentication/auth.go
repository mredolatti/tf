package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/google/uuid"
)

// ErrUserInfoMismatch is returned when attempting to update a user by ID and the supplied e-mail doesn't match
var ErrUserInfoMismatch = errors.New("user name or email don't match the supplied id")
var ErrUserExists = errors.New("user already exists")

// UserManager defines the interface for a facade that allows the rest of the application to manage users
type UserManager interface {
	Create(ctx context.Context, name string, email string, password string) (models.User, error)
	GetByID(id string) (models.User, error)
}

// UserManagerImpl implements the UserManagerFacade by means of a UserRepository
type UserManagerImpl struct {
	passwordHasher PasswordHasher
	repo repository.UserRepository
}

// NewUserManager constructs a new user manager
func NewUserManager(repo repository.UserRepository) *UserManagerImpl {
	return &UserManagerImpl{
		passwordHasher: &BCryptHasher{},
		repo: repo,
	}
}

// CreateOrUpdate creates a user if it doesn't exist.
// If it does, it checks that the main fields match, and updates the tokens
func (m *UserManagerImpl) Create(ctx context.Context, name string, email string, rawPassword string) (models.User, error) {
	_, err := m.repo.GetByEmail(ctx, email)
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

	id := uuid.New().String()
	created, err := m.repo.Add(ctx, id, name, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("error creating user in db: %w", err)
	}

	return created, nil

}

// GetByID looks up a user by it's id
func (m *UserManagerImpl) GetByID(id string) (models.User, error) {
	return m.repo.Get(context.Background(), id)
}

var _ UserManager = (*UserManagerImpl)(nil)
