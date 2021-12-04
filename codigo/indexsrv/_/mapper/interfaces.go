package mapper

import (
	"errors"
	"time"
)

// Mapping / Filesource module exported errors
var (
	ErrNotAllowed    = errors.New("permission denied")
	ErrFileNotFound  = errors.New("file not found")
	ErrUnknownUser   = errors.New("unknown user")
	ErrMappingExists = errors.New("mapping exists")
)

// File defines the read-only properties that can be retrieved from a file
type File interface {
	ID() string
	Ref() string
	Size() int64
	PatientID() string
	Updated() time.Time
}

// FileSource defines the set of methods used to interact with a file source
type FileSource interface {
	List(userID string) ([]File, error)
	GetByID(userID string, id string) (File, error)
}

// Mapping defines the properties that can be queries from a Mapping object
type Mapping interface {
	ID() string
	UserID() string
	Path() string
	File() File
	Updated() time.Time
}

// Query has optional fields that can be set to narrow the serach for a file/path
type Query struct {
	FileID        *string
	PatientID     *string
	Path          *string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}

// Interface defines the methods to be used when interacting with a mapper
type Interface interface {
	Get(userID string, query *Query) ([]Mapping, error)
	Add(userID string, fileID string, path string) error
	Remove(userID, mappingID string) error
}
