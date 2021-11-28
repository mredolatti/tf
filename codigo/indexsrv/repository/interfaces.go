package repository

import (
	"github.com/mredolatti/tf/codigo/indexsrv/models"
)

// UserRepository defines the interface for a user storage access class
type UserRepository interface {
	Get(id string) (models.User, error)
	Add(user models.User) (models.User, error)
	Remove(userID string) error
}

// OrganizationRepository defines the interface for an Organization storage access class
type OrganizationRepository interface {
	Get(user string, id string) (models.Organization, error)
	List(user string) ([]models.Organization, error)
	Add(source models.Organization) (models.Organization, error)
	Remove(id string) error
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
