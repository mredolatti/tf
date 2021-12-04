package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

// ErrUserInfoMismatch is returned when attempting to update a user by ID and the supplied e-mail doesn't match
var ErrUserInfoMismatch = errors.New("user name or email don't match the supplied id")

// UserManager defines the interface for a facade that allows the rest of the application to manage users
type UserManager interface {
	CreateOrUpdate(ctx context.Context, id string, name string, email string, accessToken string, refreshToken string) (models.User, error)
	GetByID(id string) (models.User, error)
}

// UserManagerImpl implements the UserManagerFacade by means of a UserRepository
type UserManagerImpl struct {
	repo repository.UserRepository
}

// NewUserManager constructs a new user manager
func NewUserManager(repo repository.UserRepository) *UserManagerImpl {
	return &UserManagerImpl{repo: repo}
}

// CreateOrUpdate creates a user if it doesn't exist.
// If it does, it checks that the main fields match, and updates the tokens
func (m *UserManagerImpl) CreateOrUpdate(ctx context.Context, id string, name string, email string, accessToken string, refreshToken string) (models.User, error) {
	fetched, err := m.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return m.repo.Add(ctx, id, name, email, accessToken, refreshToken)
		}
		return nil, fmt.Errorf("error fetching user from db: %w", err)
	}

	if fetched.Name() != name || fetched.Email() != email {
		return nil, ErrUserInfoMismatch
	}

	updated, err := m.repo.UpdateTokens(ctx, id, accessToken, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error updating user's tokens: %w", err)
	}

	return updated, nil
}

// GetByID looks up a user by it's id
func (m *UserManagerImpl) GetByID(id string) (models.User, error) {
	return m.repo.Get(context.Background(), id)
}

var _ UserManager = (*UserManagerImpl)(nil)
