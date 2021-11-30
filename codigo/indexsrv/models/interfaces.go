package models

import (
	"time"
)

// User defines the user model
type User interface {
	ID() string
	Name() string
}

// Organization defines the institute model
type Organization interface {
	ID() string
	Name() string
}

// FileServer defines the file-server model
type FileServer interface {
	ID() int64
	Name() string
	AuthURL() string
	FetchURL() string
	OrganizationID() string
}

// Patient defines the patient model
type Patient interface {
	ID() string
}

// UserAccount defines the `user account in file server` model
type UserAccount interface {
	UserID() string
	FileServerID() string
	Token() string
	RefreshToken() string
}

// File defines the File model
type File interface {
	ID() string
	ServerID() string
	Ref() string
	Size() int64
	PatientID() string
	Updated() time.Time
}

// Mapping defines the mapping model
type Mapping interface {
	UserID() string
	FileServerID() string
	Ref() string
	Path() string
	Updated() time.Time
}

// FileQuery has optional fields that can be set to narrow the search for a file/path
// filtering by the following criteria
type FileQuery struct {
	ID            *string
	OrgID         *string
	ServerID      *string
	PatientID     *string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}

// MappingQuery has optional fields that can be set to narrow the search for mapping
// filtering by several criteria
type MappingQuery struct {
	ID            *string
	FileID        *string
	PatientID     *string
	Path          *string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}
