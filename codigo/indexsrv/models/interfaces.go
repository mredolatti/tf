package models

import (
	"time"
)

// User defines the user model
type User interface {
	ID() string
	Name() string
	Email() string
	PasswordHash() string
	TFASecret() string
}

// Organization defines the institute model
type Organization interface {
	ID() string
	Name() string
}

// FileServer defines the file-server model
type FileServer interface {
	ID() string
	OrganizationID() string
	Name() string
	AuthURL() string
	TokenURL() string
	FetchURL() string
	ControlEndpoint() string
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
	Checkpoint() int64
}

// Mapping defines the mapping model
type Mapping interface {
	UserID() string
	FileServerID() string
	Ref() string
	Path() string
	SizeBytes() int64
	Updated() time.Time
	Deleted() bool
}

// PendingOAuth2 is user to model in-progress oauth2 flows between initial redirect & auth code arrival
type PendingOAuth2 interface {
	State() string
	UserID() string
	FileServerID() string
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

type Session interface {
	User() string
	TFADone() bool
	ValidUntil() time.Time
}
