package mappings

import "time"

// DTO is a JSON-serializable representation of a file
type DTO struct {
	IDField      string `json:"id"`
	UserIDField  string `json:"userId"`
	PathField    string `json:"path"`
	FileIDField  string `json:"fileId"`
	UpdatedField int64  `json:"updated"`
}

// ID returns the id of the mapping
func (d *DTO) ID() string {
	return d.IDField
}

// UserID returns the userId of the mapping
func (d *DTO) UserID() string {
	return d.UserIDField
}

// Path returns the path of the mapping
func (d *DTO) Path() string {
	return d.PathField
}

// FileID returns the id of the mapped file
func (d *DTO) FileID() string {
	return d.FileIDField
}

// Updated returns the time when the mapping was last updated
func (d *DTO) Updated() time.Time {
	return time.Unix(d.UpdatedField, 0)
}
