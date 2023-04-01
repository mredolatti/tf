package filemanager

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/storage"
)

// Public errors
var (
	ErrUnauthorized = errors.New("unauthorized")
)

// ListQuery specifies paramateres that can be used to firther FileMetadatas
type ListQuery struct {
	UpdatedAfter *int64
}

// Interface defines the set of methods that can be used to interact with the virtual FS
type Interface interface {
	// Metadata
	ListFileMetadata(user string, query *ListQuery) ([]models.FileMetadata, error)
	GetFileMetadata(user string, id string) (models.FileMetadata, error)
	CreateFileMetadata(user string, data models.FileMetadata) (models.FileMetadata, error)
	UpdateFileMetadata(user string, id string, data models.FileMetadata) (models.FileMetadata, error)
	DeleteFileMetadata(user string, id string) error

	// Contents
	GetFileContents(user string, id string) ([]byte, error)
	UpdateFileContents(user string, id string, data []byte) error
	DeleteFileContents(user string, id string) error

	// Permission
	Grant(user string, id string, operation authz.Operation) error
	Revoke(user string, id string, permission authz.Operation) error

	// Listeners
	AddListener(l ChangeListener)
}

// Impl implements the FileManager interface
type Impl struct {
	metadatas      storage.FilesMetadata
	files          storage.Files
	authorization  authz.Authorization
	listeners      []ChangeListener
	listenersMutex sync.RWMutex
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
func (i *Impl) ListFileMetadata(user string, query *ListQuery) ([]models.FileMetadata, error) {
	objsWithAuth, err := i.authorization.AllForSubject(user)
	if err != nil {
		return nil, fmt.Errorf("error permissions for user '%s': %w", user, err)
	}
	if len(objsWithAuth) == 0 { // supplied user doesn't have access to any file
		return nil, nil
	}

	fileIDList := make([]string, 0, len(objsWithAuth))
	for id := range objsWithAuth {
		if id != authz.AnyObject {
			fileIDList = append(fileIDList, id)
		}
	}

	if query == nil { // para que no falle
		query = &ListQuery{}
	}

	if query == nil {
		query = &ListQuery{}
	}

	metas, err := i.metadatas.GetMany(&storage.Filter{
		IDs:          fileIDList,
		UpdatedAfter: query.UpdatedAfter,
	})
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
	allowed, err := i.authorization.Can(user, authz.OperationRead, id)
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
	allowed, err := i.authorization.Can(user, authz.OperationCreate, authz.AnyObject)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	meta, err := i.metadatas.Create(data.Name(), data.Notes(), data.PatientID(), data.Type(), time.Now().UnixNano())
	if err != nil {
		return nil, fmt.Errorf("error storing new file-meta: %w", err)
	}

	i.authorization.Grant(user, authz.OperationRead, meta.ID())
	i.authorization.Grant(user, authz.OperationWrite, meta.ID())
	i.authorization.Grant(user, authz.OperationAdmin, meta.ID())

	i.notify(Change{EventType: EventFileAvailable, FileRef: meta.ID(), User: user})

	return meta, nil
}

// UpdateFileMetadata updates an already existing file-metadata record
func (i *Impl) UpdateFileMetadata(user string, id string, data models.FileMetadata) (models.FileMetadata, error) {
	allowed, err := i.authorization.Can(user, authz.OperationWrite, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if !allowed {
		return nil, ErrUnauthorized
	}

	meta, err := i.metadatas.Update(id, data, time.Now().UnixNano())
	if err != nil {
		return nil, fmt.Errorf("error updating file-meta: %w", err)
	}

	i.notify(Change{EventType: EventFileAvailable, FileRef: meta.ID(), User: user})

	return meta, nil
}

// DeleteFileMetadata removes a file-metadata record
func (i *Impl) DeleteFileMetadata(user string, id string) error {
	allowed, err := i.authorization.Can(user, authz.OperationWrite, id)
	if err != nil {
		return fmt.Errorf("error reading permissions: %w", err)
	}

	if !allowed {
		return ErrUnauthorized
	}

	err = i.metadatas.Remove(id, time.Now().UnixNano())
	if err != nil {
		return err
	}

	i.notify(Change{EventType: EventFileNotAvailable, FileRef: id, User: authz.EveryOne})
	return nil
}

// GetFileContents returns the contents of a file
func (i *Impl) GetFileContents(user string, id string) ([]byte, error) {
	// TODO(mredolatti): Fix and re-enable!
	allowed, err := i.authorization.Can(user, authz.OperationRead, id)
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
	allowed, err := i.authorization.Can(user, authz.OperationCreate, authz.AnyObject)
	if err != nil {
		return fmt.Errorf("error reading permissions: %s", err)
	}

	if !allowed {
		return ErrUnauthorized
	}

	err = i.files.Write(id, data, true)
	if err != nil {
		return err
	}

	i.notify(Change{EventType: EventFileAvailable, FileRef: id, User: authz.EveryOne})
	return nil

}

// DeleteFileContents deletest he contents of a file
func (i *Impl) DeleteFileContents(user string, id string) error {
	allowed, err := i.authorization.Can(user, authz.OperationWrite, id)
	if err != nil {
		return fmt.Errorf("error reading permissions: %s", err)
	}

	if !allowed {
		return ErrUnauthorized
	}

	err = i.files.Del(id)
	if err != nil {
		return err
	}

	i.notify(Change{EventType: EventFileNotAvailable, FileRef: id, User: authz.EveryOne})
	return nil
}

// AddListener registers a new listener that will be notified on every change
func (i *Impl) AddListener(l ChangeListener) {
	i.listenersMutex.Lock()
	i.listeners = append(i.listeners, l)
	i.listenersMutex.Unlock()
}

// Grant enables user to execute `permission` on id
func (i *Impl) Grant(user string, id string, operation authz.Operation) error {
	return i.authorization.Grant(user, operation, id)
}

// Revoke prevents user from executing `permission` on id
func (i *Impl) Revoke(user string, id string, operation authz.Operation) error {
	return i.authorization.Revoke(user, operation, id)
}

func (i *Impl) notify(c Change) {
	i.listenersMutex.RLock()
	for _, listener := range i.listeners {
		listener(c)
	}
	i.listenersMutex.RUnlock()
}

var _ Interface = (*Impl)(nil)
