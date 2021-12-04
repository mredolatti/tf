package repository

import (
	"context"
	"errors"
	"time"

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

// FileRepository defines the interface for a File
type FileRepository interface {
	List(userID string) ([]models.File, error)
	ListByOrg(userID string, OrganizationID string) ([]models.File, error)
	Get(userID string, OrganizationID string, fileID string) (models.File, error)
}

// MappingRepository defines the interface for a Mapping storage access class
type MappingRepository interface {
	List(ctx context.Context, userID string, query models.MappingQuery) ([]models.Mapping, error)
	Add(ctx context.Context, userID string, serverID string, ref string, path string, created time.Time) (models.Mapping, error)
	Update(ctx context.Context, userID string, mappingID string, mapping models.Mapping) (models.Mapping, error)
	Remove(ctx context.Context, userID string, mappingID string) error
}
