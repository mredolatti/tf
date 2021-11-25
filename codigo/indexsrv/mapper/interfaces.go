package mapper

import (
	"time"
)

type File interface {
	ID() string
	Ref() string
	Size() int64
	PatientID() string
	Updated() time.Time
}

type Mapping interface {
	ID() string
	UserId() string
	Path() string
	File() *File
	Updated() time.Time
}

type Query struct {
	FileID        *string
	PatientID     *string
	Path          *string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}

type Interface interface {
	Get(userID string, query *Query) ([]Mapping, error)
	Add(userID string, fileID string, path string) error
	Remove(userID, mappingID string) error
}
