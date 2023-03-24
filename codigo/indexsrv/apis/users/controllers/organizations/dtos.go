package organizations

type OrganizationViewDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FileServerViewDTO struct {
	ID                string `json:"id"`
	OrganizationName  string `json:"organizationName"`
	Name              string `json:"name"`
	AuthenticationURL string `json:"authenticationUrl"`
	TokenURL          string `json:"tokenUrl"`
	FileFetchURL      string `json:"fileFetchUrl"`
	ControlEndpoint   string `json:"controlEndpoint"`
}
