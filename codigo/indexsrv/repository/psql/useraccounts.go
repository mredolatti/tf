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
	accountListQuery         = "SELECT * FROM user_accounts WHERE user_id = $1"
	accountGetQuery          = "SELECT * FROM user_accounts WHERE user_id = $1 and organization_name = $2 AND server_name = $3"
	accountAddQuery          = "INSERT INTO user_accounts(user_id, organization_name, server_name, token, refresh_token) VALUES ($1, $2, $3, $4, $5) RETURNING *"
	accountCheckpointUpdate  = "UPDATE user_accounts set checkpoint = $3 WHERE user_id = $1 AND organization_name = $2 and server_name = $3 returning *"
	accountAccessTokenUpdate = "UPDATE user_accounts set token = $3 WHERE user_id = $1 AND organization_name = $2 AND server_name = $3 returning *"
	accountTokensUpdate      = "UPDATE user_accounts set token = $3, refresh_token = $4 WHERE user_id = $1 AND organization_name = $2 server_name = $3 returning *"
	accountDelQuery          = "DELETE FROM user_accounts WHERE user_id = $1 AND organization_name = $2 and server_name = $3"
)

// UserAccount is a postgres-compatible struct implementing models.UserAccount interface
type UserAccount struct {
	UserIDField           string `db:"user_id"`
	OrganizationNameField string `db:"organization_name"`
	ServerNameField       string `db:"server_name"`
	TokenFIeld            string `db:"token"`
	RefreshTokenField     string `db:"refresh_token"`
	CheckpointField       int64  `db:"checkpoint"`
}

// UserID returns the if of the user who has an account in a file server
func (u *UserAccount) UserID() string {
	return u.UserIDField
}

// FileServerID returns the id of the server in which the user has the account
func (u *UserAccount) OrganizationName() string {
	return u.OrganizationNameField
}

// FileServerID returns the id of the server in which the user has the account
func (u *UserAccount) FileServerName() string {
	return u.ServerNameField
}

// Token returns the token used to make request on behalf of this user to the server
func (u *UserAccount) Token() string {
	return u.TokenFIeld
}

// RefreshToken returns the token used to get new tokens when the current one expires
func (u *UserAccount) RefreshToken() string {
	return u.RefreshTokenField
}

// Checkpoint returns a nanosecond-granularity timestamp of the last update (or zero y if has never happend)
func (u *UserAccount) Checkpoint() int64 {
	return u.CheckpointField
}

// UserAccountRepository is a mapping to a table in postgres that allows enables operations on user accounts
type UserAccountRepository struct {
	db *sqlx.DB
}

// NewUserAccountRepository constructs a new postgresql-based file server repository
func NewUserAccountRepository(db *sqlx.DB) *UserAccountRepository {
	return &UserAccountRepository{db: db}
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
func (r *UserAccountRepository) Get(ctx context.Context, userID string, orgName string, serverName string) (models.UserAccount, error) {
	var account UserAccount
	err := r.db.QueryRowxContext(ctx, accountGetQuery, userID, orgName, serverName).StructScan(&account)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing accounts::get in postgres: %w", err)
	}

	return &account, nil
}

// Add adds a new user account
func (r *UserAccountRepository) Add(ctx context.Context, userID, orgName, serverName, accessToken, refreshToken string) (models.UserAccount, error) {
	var account UserAccount
	err := r.db.QueryRowxContext(ctx, accountAddQuery, userID, orgName, serverName, accessToken, refreshToken).StructScan(&account)
	if err != nil {
		return nil, fmt.Errorf("error executing accounts::add in postgres: %w", err)
	}
	return &account, nil
}

// Remove deletes a user account
func (r *UserAccountRepository) Remove(ctx context.Context, userID string, orgName string, serverName string) error {
	_, err := r.db.ExecContext(ctx, accountDelQuery, userID, orgName, serverName)
	if err != nil {
		return fmt.Errorf("error executiong accounts::del in postgres: %w", err)
	}
	return nil
}

// UpdateCheckpoint updates the checkpoint on an account
func (r *UserAccountRepository) UpdateCheckpoint(ctx context.Context, userID string, orgName string, serveName string, newCheckpoint int64) error {
	_, err := r.db.ExecContext(ctx, accountCheckpointUpdate, userID, orgName, serveName, newCheckpoint)
	if err != nil {
		return fmt.Errorf("error executing accounts::update_checkpoint in postgres: %w", err)
	}
	return nil

}

// UpdateTokens updates access (& maybe refresh) tokens for a user account
func (r *UserAccountRepository) UpdateTokens(ctx context.Context, userID, orgName, serverName, accessToken, refreshToken string) error {

	if refreshToken == "" { // update access token only
		_, err := r.db.ExecContext(ctx, accountAccessTokenUpdate, userID, orgName, serverName, accessToken)
		if err != nil {
			return fmt.Errorf("error executing accounts::update_access_token in postgres: %w", err)
		}
		return nil
	}

	// update both access & refresh tokens
	_, err := r.db.ExecContext(ctx, accountTokensUpdate, userID, orgName, serverName, accessToken, refreshToken)
	if err != nil {
		return fmt.Errorf("error executing accounts::update_all_tokens in postgres: %w", err)
	}
	return nil
}

var _ repository.UserAccountRepository = (*UserAccountRepository)(nil)
var _ models.UserAccount = (*UserAccount)(nil)
