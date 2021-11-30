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
	Get(id string) (models.User, error)
	Add(user models.User) (models.User, error)
	Remove(userID string) error
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
	List(userID string, query models.MappingQuery) ([]models.Mapping, error)
	Add(userID string, mapping models.Mapping) (models.Mapping, error)
	Update(userID string, mappingID string, mapping models.Mapping) (models.Mapping, error)
	Delete(userID string, mappingID string) error
}
