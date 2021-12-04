package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/jmoiron/sqlx"
)

const (
	userListQuery         = "SELECT * FROM users"
	userGetQuery          = "SELECT * FROM users WHERE id = $1"
	userAddQuery          = "INSERT INTO users(id, name,email,access_token,refresh_token) VALUES ($1, $2, $3, $4, $5) RETURNING *"
	userUpdateTokensQuery = "UPDATE users SET access_token = $2, refresh_token = $3 WHERE id = $1 RETURNING *"
	userDelQuery          = "DELETE FROM users WHERE id = $1"
)

// User is a postgres-compatible struct implementing models.User interface
type User struct {
	IDField           string `db:"id"`
	NameField         string `db:"name"`
	EmailField        string `db:"email"`
	AccessTokenField  string `db:"access_token"`
	RefreshTokenField string `db:"refresh_token"`
}

// ID returns the id of the user
func (f *User) ID() string {
	return f.IDField
}

// Name returns the name of the user
func (f *User) Name() string {
	return f.NameField
}

// Email returns the email of the user
func (f *User) Email() string {
	return f.EmailField
}

// AccessToken returns the last access token of the user
func (f *User) AccessToken() string {
	return f.AccessTokenField
}

// RefreshToken returns the refresh token for the user
func (f *User) RefreshToken() string {
	return f.RefreshTokenField
}

// UserRepository is a mapping to a table in postgres that allows enables operations on users
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a new postgresql-based user repository
func NewUserRepository(db *sqlx.DB) (*UserRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &UserRepository{db: db}, nil
}

// List returns a list of all the users in the system
func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	var users []User
	err := r.db.SelectContext(ctx, &users, userListQuery)
	if err != nil {
		return nil, fmt.Errorf("error executing users::list in postgres: %w", err)
	}

	ret := make([]models.User, len(users))
	for idx := range users {
		ret[idx] = &users[idx]
	}

	return ret, nil
}

// Get returns a user by ID
func (r *UserRepository) Get(ctx context.Context, id string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userGetQuery, id).StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing users::get in postgres: %w", err)
	}

	return &user, nil
}

// Add adds a new user to the system
func (r *UserRepository) Add(ctx context.Context, id string, name string, email string, accessToken string, refreshToken string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userAddQuery, id, name, email, accessToken, refreshToken).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("error executing users::add in postgres: %w", err)
	}
	return &user, nil
}

// Remove deletes a user matching the supplied id
func (r *UserRepository) Remove(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, userDelQuery, id)
	if err != nil {
		return fmt.Errorf("error executiong users::del in postgres: %w", err)
	}
	return nil
}

// UpdateTokens updates the access & refresh tokens for a specific user
func (r *UserRepository) UpdateTokens(ctx context.Context, id string, accessToken string, refreshToken string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userUpdateTokensQuery, id, accessToken, refreshToken).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("error executing users::update_tokens in postgres: %w", err)
	}
	return &user, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
