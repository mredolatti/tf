package organizations

// DTO is a JSON-serializable representation of an organization
type DTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
