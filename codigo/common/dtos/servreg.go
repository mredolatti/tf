package dtos

type ServerInfoDTO struct {
	ID              string `json:"id"`
	OrgID           string `json:"orgId"`
	Name            string `json:"name"`
	AuthURL         string `json:"authUrl"`
	TokenURL        string `json:"tokenUrl"`
	FetchURL        string `json:"fetchUrl"`
	ControlEndpoint string `json:"controlEndpoint"`
}
