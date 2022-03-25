package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/repository"

	"github.com/jmoiron/sqlx"
)

const (
	clientGetByID = "SELECT * FROM clients WHERE id = $1"
	clientAdd     = "INSERT INTO clients(id, secret, domain, user_id) VALUES ($1, $2, $3, $4) RETURNING *"
	clientDelete  = "DELETE FROM clients WHERE id = $1"
)

// Client is a postgres-compatible struct implementing models.Client interface
type Client struct {
	IDField     string `db:"id"`
	SecretField string `db:"secret"`
	DomainField string `db:"domain"`
	UserIDField string `db:"user_id"`
}

func (c *Client) GetID() string {
	return c.IDField
}

func (c *Client) GetSecret() string {
	return c.SecretField
}

func (c *Client) GetDomain() string {
	return c.DomainField
}

func (c *Client) GetUserID() string {
	return c.UserIDField
}

// ClientRepository is a mapping to a table in postgres that allows enables operations
// on file server
type ClientRepository struct {
	db *sqlx.DB
}

// NewClientRepository constructs a new postgresql-based file server repository
func NewClientRepository(db *sqlx.DB) (*ClientRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &ClientRepository{db: db}, nil
}

// GetByID returns an file server that matches the supplied id
func (r *ClientRepository) GetByID(ctx context.Context, id string) (models.ClientInfo, error) {
	var client Client
	err := r.db.QueryRowxContext(ctx, clientGetByID, id).StructScan(&client)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing client::get_by_id in postgres: %w", err)
	}
	return &client, nil
}

// Add a new client
func (r *ClientRepository) Add(ctx context.Context, id string, secret string, domain string, userID string) (models.ClientInfo, error) {
	var client Client
	err := r.db.QueryRowxContext(ctx, clientAdd, id, secret, domain, userID).StructScan(&client)
	if err != nil {
		return nil, fmt.Errorf("error executing client::add in postgres: %w", err)
	}
	return &client, nil
}

// Remove a client
func (r *ClientRepository) Remove(ctx context.Context, clientID string) error {
	res := r.db.QueryRowxContext(ctx, clientDelete, clientID)
	if err := res.Err(); err != nil {
		return fmt.Errorf("error executing client::delete in postgres: %w", err)
	}
	return nil
}

var _ models.ClientInfo = (*Client)(nil)
var _ repository.OAuth2ClientRepository = (*ClientRepository)(nil)
