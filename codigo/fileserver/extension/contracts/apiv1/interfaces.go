package apiv1

import (
	"errors"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts"
)

type CreateFunc = func(map[string]interface{}) (Plugin, error)

const V contracts.Version = 1

const CreateFuncName = "Create"

// ---------------
// File & Metadata
// ---------------

// Public errors
var (
	ErrFileDoesNotExist = errors.New("file does not exist")
	ErrFileExists       = errors.New("file exists")
)

// File methods
type File interface {
	ID() string
	Contents() []byte
}

// FileMetadata methods
type FileMetadata interface {
	ID() string
	Name() string
	Notes() string
	PatientID() string
	SizeBytes() int64
	Type() string
	ContentID() string
	LastUpdated() int64
	Deleted() bool
}

// Filter for retrieving files
type Filter struct {
	IDs          []string
	UpdatedAfter *int64
}

// FilesMetadata defines the set of operations to be performed on file metadata records
type FilesMetadata interface {
	Get(id string) (FileMetadata, error)
	GetMany(filter *Filter) (map[string]FileMetadata, error)
	Create(name string, notes string, patient string, typ string, whenNs int64) (FileMetadata, error)
	Update(id string, updated FileMetadata, whenNs int64) (FileMetadata, error)
	Remove(id string, whenNs int64) error
}

// Files defines the set of operations that can be performed on file contents
type Files interface {
	Read(id string) ([]byte, error)
	Write(id string, data []byte, force bool) error
	Del(id string) error
}

// -------------
// Authorization
// -------------

// AnyObject is used as an object for permissions that don't target a specific one (ie: Create)
const AnyObject = "__GLOBAL__"

// EveryOne is used as an object for permissions that don't target a specific user, but affects everyone
const EveryOne = "__EVERYONE__"

// Permission types bitmask constants
type Operation uint32

const (
	OperationRead   Operation = (1 << 0)
	OperationWrite  Operation = (1 << 1)
	OperationCreate Operation = (1 << 2)
	OperationAdmin  Operation = (1 << 31)
)

// Public errors
var (
	ErrNoSuchUser       = errors.New("no such user")
	ErrNosuchObject     = errors.New("no such object")
	ErrNoSuchPermission = errors.New("no such permission type")
)

type Permission interface {
	Can(operation Operation) (bool, error)
	Grant(operation Operation) error
	Revoke(operation Operation) error
}

// Authorization defines the interface of an authorization handling component
type Authorization interface {
	Can(subject string, operation Operation, object string) (bool, error)
	Grant(subject string, operation Operation, object string) error
	Revoke(subject string, operation Operation, object string) error
	AllForSubject(subject string) (map[string]Permission, error)
	AllForObject(object string) (map[string]Permission, error)
}

// ---------------------
// Main plugin interface
// ---------------------

type Plugin interface {
	GetFileStorage() Files
	GetFileMetadataStorage() FilesMetadata
	GetAuthorization() Authorization
}
