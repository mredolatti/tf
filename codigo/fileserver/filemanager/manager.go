package filemanager

import (
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/storage"
)

// Public errors
var (
	ErrUnauthorized = errors.New("unauthorized")
)

// Interface defines the set of methods that can be used to interact with the virtual FS
type Interface interface {
	// Metadata
	ListFileMetadata(user string) ([]models.FileMetadata, error)
	GetFileMetadata(user string, id string) (models.FileMetadata, error)
	CreateFileMetadata(user string, data models.FileMetadata) (models.FileMetadata, error)
	UpdateFileMetadata(user string, id string, data models.FileMetadata) (models.FileMetadata, error)
	DeleteFileMetadata(user string, id string) error

	// Contents
	GetFileContents(user string, id string) ([]byte, error)
	UpdateFileContents(user string, id string, data []byte) error
	DeleteFileContents(user string, id string) error
}

// Impl implements the FileManager interface
type Impl struct {
	metadatas     storage.FilesMetadata
	files         storage.Files
	authorization authz.Authorization
}

// New constructs a new file manager
func New(files storage.Files, metadatas storage.FilesMetadata, authorization authz.Authorization) *Impl {
	return &Impl{
		files:         files,
		metadatas:     metadatas,
		authorization: authorization,
	}
}

// ListFileMetadata lists all known file (metas) that a user has access to
func (i *Impl) ListFileMetadata(user string) ([]models.FileMetadata, error) {
	objsWithAuth := i.authorization.AllForSubject(user)
	fileIDList := make([]string, 0, len(objsWithAuth))
	for id := range objsWithAuth {
		fileIDList = append(fileIDList, id)
	}

	metas, err := i.metadatas.GetMany(fileIDList)
	if err != nil {
		return nil, err
	}

	result := make([]models.FileMetadata, 0, len(metas))
	for _, meta := range metas {
		result = append(result, meta)
	}

	return result, nil

}

// GetFileMetadata fetches a single file-metadata record
func (i *Impl) GetFileMetadata(user string, id string) (models.FileMetadata, error) {
	allowed, err := i.authorization.Can(user, authz.Read, id)
	if err != nil {
		return nil, fmt.Errorf("error reading permissions: %w", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	meta, err := i.metadatas.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error reading file metadata: %w", err)
	}

	return meta, err
}

// CreateFileMetadata creates a file-metadata record
func (i *Impl) CreateFileMetadata(user string, data models.FileMetadata) (models.FileMetadata, error) {
	allowed, err := i.authorization.Can(user, authz.Create, authz.AnyObject)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	meta, err := i.metadatas.Create(data.Name(), data.Notes(), data.PatientID(), data.Type())
	if err != nil {
		return nil, fmt.Errorf("error storing new file-meta: %w", err)
	}

	i.authorization.Grant(user, authz.Read, meta.ID())
	i.authorization.Grant(user, authz.Write, meta.ID())
	i.authorization.Grant(user, authz.Admin, meta.ID())
	return meta, nil
}

// UpdateFileMetadata updates an already existing file-metadata record
func (i *Impl) UpdateFileMetadata(user string, id string, data models.FileMetadata) (models.FileMetadata, error) {
	allowed, err := i.authorization.Can(user, authz.Write, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	meta, err := i.metadatas.Update(id, data)
	if err != nil {
		return nil, fmt.Errorf("error updating file-meta: %w", err)
	}

	return meta, nil
}

// DeleteFileMetadata removes a file-metadata record
func (i *Impl) DeleteFileMetadata(user string, id string) error {
	allowed, err := i.authorization.Can(user, authz.Write, id)
	if err != nil {
		return fmt.Errorf("error reading permissions: %w", err)
	}

	if !allowed {
		return ErrUnauthorized
	}

	return i.metadatas.Remove(id)
}

// GetFileContents returns the contents of a file
func (i *Impl) GetFileContents(user string, id string) ([]byte, error) {
	allowed, err := i.authorization.Can(user, authz.Read, id)
	if err != nil {
		return nil, fmt.Errorf("error reading permissions: %s", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	return i.files.Read(id)
}

// UpdateFileContents updates the contents of a file
func (i *Impl) UpdateFileContents(user string, id string, data []byte) error {
	allowed, err := i.authorization.Can(user, authz.Create, authz.AnyObject)
	if err != nil {
		return fmt.Errorf("error reading permissions: %s", err)
	}

	if !allowed {
		return ErrUnauthorized
	}

	return i.files.Write(id, data, true)
}

// DeleteFileContents deletest he contents of a file
func (i *Impl) DeleteFileContents(user string, id string) error {
	allowed, err := i.authorization.Can(user, authz.Write, id)
	if err != nil {
		return fmt.Errorf("error reading permissions: %s", err)
	}

	if !allowed {
		return ErrUnauthorized
	}
	return i.files.Del(id)
}

var _ Interface = (*Impl)(nil)
