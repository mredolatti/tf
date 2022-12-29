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
	oauthPutQuery = "INSERT INTO oauth2_pending(user_id,server_id,state) VALUES ($1, $2, $3) RETURNING *"
	oauthPopQuery = "DELETE FROM oauth2_pending WHERE state = $1 RETURNING *"
)

// PendingOAuth2 represents an in-progress oauth2 flow
type PendingOAuth2 struct {
	FileServerIDField string `db:"server_id"`
	UserIDField       string `db:"user_id"`
	StateField        string `db:"state"`
}

// FileServerID returns the if of the file server we're trying au authenticate in
func (p *PendingOAuth2) FileServerID() string {
	return p.FileServerIDField
}

// UserID returns the user we're trying to authenticate
func (p *PendingOAuth2) UserID() string {
	return p.UserIDField
}

// State returns the randomized-code used to secure the request (and map to user_id, server_id)
func (p *PendingOAuth2) State() string {
	return p.StateField
}

// PendingOAuth2Repository is a postgres-based implementation of an in-progress oauth2 flow repository
type PendingOAuth2Repository struct {
	db *sqlx.DB
}

// NewPendingOAuth2Repository constructs a new PendingOAuth2Repository
func NewPendingOAuth2Repository(db *sqlx.DB) *PendingOAuth2Repository {
	return &PendingOAuth2Repository{db: db}
}

// Put starts tracking a new flow
func (r *PendingOAuth2Repository) Put(ctx context.Context, userID string, serverID string, state string) (models.PendingOAuth2, error) {
	var flow PendingOAuth2
	err := r.db.QueryRowxContext(ctx, oauthPutQuery, userID, serverID, state).StructScan(&flow)
	if err != nil {
		return nil, fmt.Errorf("error executing oauth2flow::put in postgres: %w", err)
	}
	return &flow, nil
}

// Pop fetches & deletes an oauth2 flow by state
func (r *PendingOAuth2Repository) Pop(ctx context.Context, state string) (models.PendingOAuth2, error) {
	var flow PendingOAuth2
	err := r.db.QueryRowxContext(ctx, oauthPopQuery, state).StructScan(&flow)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing oauth2::pop in postgres: %w", err)
	}
	return &flow, nil
}

var _ repository.PendingOAuth2Repository = (*PendingOAuth2Repository)(nil)
