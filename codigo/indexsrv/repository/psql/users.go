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
	userListQuery           = "SELECT * FROM users"
	userGetQuery            = "SELECT * FROM users WHERE id = $1"
	userGetByEmailQuery     = "SELECT * FROM users WHERE email = $1"
	userAddQuery            = "INSERT INTO users(name,email,password_hash) VALUES ($1, $2, $3) RETURNING *"
	userUpdatePasswordQuery = "UPDATE users SET password_hash = $2 WHERE id = $1 RETURNING *"
	userUpdate2FAQuery      = "UPDATE users SET tfa_secret = $2 WHERE id = $1 RETURNING *"
	userDelQuery            = "DELETE FROM users WHERE id = $1"
)

// User is a postgres-compatible struct implementing models.User interface
type User struct {
	IDField           string `db:"id"`
	NameField         string `db:"name"`
	EmailField        string `db:"email"`
	PasswordHashField string `db:"password_hash"`
	TFASecretField    *string `db:"tfa_secret"`
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
func (f *User) PasswordHash() string {
	return f.PasswordHashField
}

// TFASecret implements models.User
func (f *User) TFASecret() string {
	if f.TFASecretField == nil {
		return ""
	}

	return *f.TFASecretField
}

// UserRepository is a mapping to a table in postgres that allows enables operations on users
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a new postgresql-based user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
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

// GetByEmail returns a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userGetByEmailQuery, email).StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing users::get in postgres: %w", err)
	}

	return &user, nil
}

// Add adds a new user to the system
func (r *UserRepository) Add(ctx context.Context, name string, email string, passwordHash string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userAddQuery, name, email, passwordHash).StructScan(&user)
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
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userUpdatePasswordQuery, id, passwordHash).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("error executing users::update_password in postgres: %w", err)
	}
	return &user, nil
}

// Update2FA implements repository.UserRepository
func (r *UserRepository) Update2FA(ctx context.Context, id string, tfaSecret string) error {
	if err := r.db.QueryRowxContext(ctx, userUpdate2FAQuery, id, tfaSecret).Err(); err != nil {
		return fmt.Errorf("error executing users::update_2faSecret in postgres: %w", err)
	}
	return nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
var _ models.User = (*User)(nil)
