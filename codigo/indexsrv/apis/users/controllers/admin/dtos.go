package admin

// DTO is a JSON-serializable representation of an organization
type DTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	X509Certificate string `json:"x509Certificate"`
}
