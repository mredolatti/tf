package repository

import (
	"context"
	"errors"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
)

// ErrNotFound is an error to be returned when a requested item is not found
var ErrNotFound = errors.New("not found")

// UserRepository defines the interface for a user storage access class
type UserRepository interface {
	Get(ctx context.Context, id string) (models.User, error)
	Add(ctx context.Context, id string, name string, email string, accessToken string, refreshToken string) (models.User, error)
	UpdateTokens(ctx context.Context, id string, accessToken string, refreshToken string) (models.User, error)
	Remove(ctx context.Context, userID string) error
}

// OrganizationRepository defines the interface for an Organization storage access class
type OrganizationRepository interface {
	Get(ctx context.Context, id string) (models.Organization, error)
	List(ctx context.Context) ([]models.Organization, error)
	Add(ctx context.Context, source models.Organization) (models.Organization, error)
	Remove(ctx context.Context, id string) error
}

// MappingRepository defines the interface for a Mapping storage access class
type MappingRepository interface {
	List(ctx context.Context, userID string, query models.MappingQuery) ([]models.Mapping, error)
	Add(ctx context.Context, userID string, mapping models.Mapping) (models.Mapping, error)
	Update(ctx context.Context, userID string, mappingID string, mapping models.Mapping) (models.Mapping, error)
	Remove(ctx context.Context, userID string, mappingID string) error
	HandleServerUpdates(ctx context.Context, userID string, updates []models.Update) error
}

// FileServerRepository defines the interface for a file server collection
type FileServerRepository interface {
	List(ctx context.Context, orgID string) ([]models.FileServer, error)
	Get(ctx context.Context, id string) (models.FileServer, error)
	Add(ctx context.Context, id string, name string, orgID string, authURL string, tokenURL string, fetchURL string, controlEndpoint string) (models.FileServer, error)
	Remove(ctx context.Context, id string) error
}

// UserAccountRepository defines the interface for interacting with user accounts on file servers
type UserAccountRepository interface {
	List(ctx context.Context, userID string) ([]models.UserAccount, error)
	Get(ctx context.Context, userID string, serverID string) (models.UserAccount, error)
	Add(ctx context.Context, userID, serverID, accessToken, refreshToken string) (models.UserAccount, error)
	Remove(ctx context.Context, userID string, serverID string) error
	UpdateCheckpoint(ctx context.Context, userID string, serverID string, newCheckpoint int64) error
}

// PendingOAuth2Repository is used to store & retreive in-progress oauth2 flows metadata
type PendingOAuth2Repository interface {
	Put(ctx context.Context, userID string, serverID string, state string) (models.PendingOAuth2, error)
	Pop(ctx context.Context, state string) (models.PendingOAuth2, error)
}
