package dtos

type ServerInfoStatus int

const (
	StatusServerReady         ServerInfoStatus = 0
	StatusServerNotRegistered ServerInfoStatus = 1
	StatusServerDisabled      ServerInfoStatus = 2
)

type RegistrationResult int

const (
	ResultOK                RegistrationResult = 0
	ResultAlreadyRegistered RegistrationResult = 1
	ResultFail              RegistrationResult = 2
)

type ServerStatusDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type RegistrationResultDTO struct {
	ServerInfo ServerStatusDTO    `json:"serverInfo"`
	Result     RegistrationResult `json:"result"`
}
