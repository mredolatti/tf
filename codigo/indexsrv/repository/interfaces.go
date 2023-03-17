package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")

type Factory interface {
	Users() UserRepository
	Organizations() OrganizationRepository
	Mappings() MappingRepository
	FileServers() FileServerRepository
	Accounts() UserAccountRepository
	PendingOAuth() PendingOAuth2Repository
}

// UserRepository defines the interface for a user storage access class
type UserRepository interface {
	Get(ctx context.Context, id string) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	Add(ctx context.Context, name string, email string, passwordHash string) (models.User, error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) (models.User, error)
	Update2FA(ctx context.Context, userID string, totp string) error
	Remove(ctx context.Context, userID string) error
}

// OrganizationRepository defines the interface for an Organization storage access class
type OrganizationRepository interface {
	Get(ctx context.Context, id string) (models.Organization, error)
	GetByName(ctx context.Context, name string) (models.Organization, error)
	List(ctx context.Context) ([]models.Organization, error)
	Add(ctx context.Context, name string) (models.Organization, error)
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
	Add(ctx context.Context, name string, orgID string, authURL string, tokenURL string, fetchURL string, controlEndpoint string) (models.FileServer, error)
	Remove(ctx context.Context, id string) error
}

// UserAccountRepository defines the interface for interacting with user accounts on file servers
type UserAccountRepository interface {
	List(ctx context.Context, userID string) ([]models.UserAccount, error)
	Get(ctx context.Context, userID string, serverID string) (models.UserAccount, error)
	Add(ctx context.Context, userID, serverID, passwordHash, refreshToken string) (models.UserAccount, error)
	Remove(ctx context.Context, userID string, serverID string) error
	UpdateCheckpoint(ctx context.Context, userID string, serverID string, newCheckpoint int64) error
	UpdateTokens(ctx context.Context, userID, serverID, accessToken, refreshToken string) error
}

// PendingOAuth2Repository is used to store & retreive in-progress oauth2 flows metadata
type PendingOAuth2Repository interface {
	Put(ctx context.Context, userID string, serverID string, state string) (models.PendingOAuth2, error)
	Pop(ctx context.Context, state string) (models.PendingOAuth2, error)
}

type SessionRepository interface {
	Get(ctx context.Context, token string) (models.Session, error)
	Put(ctx context.Context, token string, userID string, tfaDone bool, TTL time.Duration) error
	Remove(ctx context.Context, token string) error
}
