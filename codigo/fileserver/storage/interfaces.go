package storage

import (
	"errors"

	"github.com/mredolatti/tf/codigo/fileserver/models"
)

// Public errors
var (
	ErrNoSuchFile = errors.New("file not found")
	ErrFileExists = errors.New("file exists")
)

// Filter for retrieving files
type Filter struct {
	IDs          []string
	UpdatedAfter *int64
}

// FilesMetadata defines the set of operations to be performed on file metadata records
type FilesMetadata interface {
	Get(id string) (models.FileMetadata, error)
	GetMany(filter *Filter) (map[string]models.FileMetadata, error)
	Create(name string, notes string, patient string, typ string, whenNs int64) (models.FileMetadata, error)
	Update(id string, updated models.FileMetadata, whenNs int64) (models.FileMetadata, error)
	Remove(id string, whenNs int64) error
}

// Files defines the set of operations that can be performed on file contents
type Files interface {
	Read(id string) ([]byte, error)
	Write(id string, data []byte, force bool) error
	Del(id string) error
}
