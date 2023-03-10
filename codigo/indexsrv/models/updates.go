package models

import "fmt"

// Update encapsulates the update DTO sent as sent from a file server with extra information
type Update struct {
	OrganizationID string
	ServerID       string
	FileRef        string
	Checkpoint     int64
	SizeBytes      int64
	ChangeType     UpdateType
}

func (u *Update) UnmappedPath() string {
	return fmt.Sprintf("unassigned/%s/%s", u.ServerID, u.FileRef)
}

// UpdateType is the enumeration type used for update types
type UpdateType int

// UpdateType constants
const (
	UpdateTypeFileAdd    UpdateType = 0
	UpdateTypeFileDelete UpdateType = 1
	UpdateTypeFileUpdate UpdateType = 2
)
