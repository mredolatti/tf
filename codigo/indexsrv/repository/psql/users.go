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
	userListQuery = "SELECT * FROM users"
	userGetQuery  = "SELECT * FROM users WHERE id = $1"
	userAddQuery  = "INSERT INTO users(id, name) VALUES ($1, $2) RETURNING *"
	userDelQuery  = "DELETE FROM users WHERE id = $1"
)

// User is a postgres-compatible struct implementing models.User interface
type User struct {
	IDField   string `db:"id"`
	NameField string `db:"name"`
}

// ID returns the id of the user
func (f *User) ID() string {
	return f.IDField
}

// Name returns the name of the user
func (f *User) Name() string {
	return f.NameField
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
func (r *UserRepository) Add(ctx context.Context, id string, name string) (models.User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, userAddQuery, id, name).StructScan(&user)
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
