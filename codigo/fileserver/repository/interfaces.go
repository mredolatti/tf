package repository

import (
	"context"
	"errors"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/mredolatti/tf/codigo/fileserver/models"
)

// Public errors
var (
	ErrNotFound = errors.New("not found")
)

// OAuth2ClientRepository defines the set of methods to create, retrieve and remove oauth2 client information
type OAuth2ClientRepository interface {
	oauth2.ClientStore
	Add(ctx context.Context, id string, secret string, domain string, userID string) (models.ClientInfo, error)
	Remove(ctx context.Context, clientID string) error
}

// OAuth2TokenRepository defines the set of methods to create, retreive and remove aauth2 tokens
type OAuth2TokenRepository = oauth2.TokenStore
