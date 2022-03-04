package models

// Update encapsulates the update DTO sent as sent from a file server with extra information
type Update struct {
	OrganizationID string
	ServerID       string
	FileRef        string
	Checkpoint     int64
	ChangeType     UpdateType
}

// UpdateType is the enumeration type used for update types
type UpdateType int

// UpdateType constants
const (
	UpdateTypeFileAdd    UpdateType = 0
	UpdateTypeFileDelete UpdateType = 1
	UpdateTypeFileUpdate UpdateType = 2
)
