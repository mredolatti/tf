package dtos

type ServerInfoDTO struct {
	OrgName         string `json:"orgName"`
	Name            string `json:"name"`
	AuthURL         string `json:"authUrl"`
	TokenURL        string `json:"tokenUrl"`
	FetchURL        string `json:"fetchUrl"`
	ControlEndpoint string `json:"controlEndpoint"`
}
