package controllers

// StatusDTO is the object posted by a server when it wants to announce itself
type StatusDTO struct {
	ServerID string `json:"serverId"`
	Healthy  bool   `json:"healthy"`
	Uptime   int64  `json:"uptime"`
}
