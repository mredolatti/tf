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

// FilesMetadata defines the set of operations to be performed on file metadata records
type FilesMetadata interface {
	Get(id string) (models.FileMetadata, error)
	GetMany(ids []string) (map[string]models.FileMetadata, error)
	Create(name string, notes string, patient string, typ string) (models.FileMetadata, error)
	Update(id string, updated models.FileMetadata) (models.FileMetadata, error)
	Remove(id string) error
}

// Files defines the set of operations that can be performed on file contents
type Files interface {
	Read(id string) ([]byte, error)
	Write(id string, data []byte, force bool) error
	Del(id string) error
}
