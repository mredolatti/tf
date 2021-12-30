package models

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
}
