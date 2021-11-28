package files

// DTO is a JSON-serializable representation of a file
type DTO struct {
	ID        string `json:"id"`
	ServerID  string `json:"serverId"`
	Ref       string `json:"ref"`
	Size      int64  `json:"size"`
	PatientID string `json:"patientId"`
	Updated   int64  `json:"updated"`
}
