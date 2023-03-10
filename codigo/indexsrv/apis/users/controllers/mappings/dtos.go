package mappings

import "time"

// DTO is a JSON-serializable representation of a file
type DTO struct {
	UserIDField    string `json:"userId"`
	ServerIDField  string `json:"serverId"`
	PathField      string `json:"path"`
	RefField       string `json:"ref"`
	SizeBytesField int64  `json:"sizeBytes"`
	UpdatedField   int64  `json:"updated"`
	DeletedField   bool   `json:"deleted"`
}

// UserID returns the id of the user
func (d *DTO) UserID() string {
	return d.UserIDField
}

// FileServerID returns the userId of the mapping
func (d *DTO) FileServerID() string {
	return d.ServerIDField
}

func (d *DTO) SizeBytes() int64 {
	return d.SizeBytesField
}

// Path returns the path of the mapping
func (d *DTO) Path() string {
	return d.PathField
}

// Ref returns the internal reference to the file on the actual server
func (d *DTO) Ref() string {
	return d.RefField
}

// Updated returns the time when the mapping was last updated
func (d *DTO) Updated() time.Time {
	return time.Unix(d.UpdatedField, 0)
}

// Deleted returns true if the referenced file is no longer available
func (d *DTO) Deleted() bool {
	return d.DeletedField
}
