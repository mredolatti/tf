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
	accountListQuery = "SELECT * FROM user_accounts WHERE user_id = $1"
	accountGetQuery  = "SELECT * FROM user_accounts WHERE id = $1"
	accountAddQuery  = "INSERT INTO user_accounts(user_id, server_id, token, refresh_token) VALUES ($1, $2, $3, $4) RETURNING *"
	accountDelQuery  = "DELETE FROM user_accounts WHERE user_id = $1 AND server_id = $2"
)

// UserAccount is a postgres-compatible struct implementing models.UserAccount interface
type UserAccount struct {
	UserIDField       string
	ServerIDField     string
	TokenFIeld        string
	RefreshTokenField string
}

// UserID returns the if of the user who has an account in a file server
func (u *UserAccount) UserID() string {
	return ""
}

// FileServerID returns the id of the server in which the user has the account
func (u *UserAccount) FileServerID() string {
	return ""
}

// Token returns the token used to make request on behalf of this user to the server
func (u *UserAccount) Token() string {
	return ""
}

// RefreshToken returns the token used to get new tokens when the current one expires
func (u *UserAccount) RefreshToken() string {
	return ""
}

// UserAccountRepository is a mapping to a table in postgres that allows enables operations on user accounts
type UserAccountRepository struct {
	db *sqlx.DB
}

// NewUserAccountRepository constructs a new postgresql-based file server repository
func NewUserAccountRepository(db *sqlx.DB) (*UserAccountRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &UserAccountRepository{db: db}, nil
}

// List returns a list of all the user accounts
func (r *UserAccountRepository) List(ctx context.Context, userID string) ([]models.UserAccount, error) {
	var accounts []UserAccount
	err := r.db.SelectContext(ctx, &accounts, accountListQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing accounts::list in postgres: %w", err)
	}

	ret := make([]models.UserAccount, len(accounts))
	for idx := range accounts {
		ret[idx] = &accounts[idx]
	}

	return ret, nil
}

// Get returns a user account matching the supplied userID and serverID
func (r *UserAccountRepository) Get(ctx context.Context, userID string, serverID string) (models.UserAccount, error) {
	var account UserAccount
	err := r.db.QueryRowxContext(ctx, accountGetQuery, userID, serverID).StructScan(&account)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing accounts::get in postgres: %w", err)
	}

	return &account, nil
}

// Add adds a new user account
func (r *UserAccountRepository) Add(ctx context.Context, id string, name string) (models.UserAccount, error) {
	var account UserAccount
	err := r.db.QueryRowxContext(ctx, accountAddQuery, id, name).StructScan(&account)
	if err != nil {
		return nil, fmt.Errorf("error executing accounts::add in postgres: %w", err)
	}
	return &account, nil
}

// Remove deletes a user account
func (r *UserAccountRepository) Remove(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, accountDelQuery, id)
	if err != nil {
		return fmt.Errorf("error executiong accounts::del in postgres: %w", err)
	}
	return nil
}
