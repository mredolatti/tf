package models

import (
	"github.com/go-oauth2/oauth2/v4"
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
	Type() string
	ContentID() string
	LastUpdated() int64
	Deleted() bool
}

// TokenInfo type alias
type TokenInfo = oauth2.TokenInfo

// ClientInfo type alias
type ClientInfo = oauth2.ClientInfo
