package client

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/fileserver/api/client/login"
	"github.com/mredolatti/tf/codigo/fileserver/api/client/middleware"

	"github.com/gin-gonic/gin"
)

// Options contains user-api configuration parameters
type Options struct {
	Host                     string
	Port                     int
	ServerCertificateChainFN string
	ServerPrivateKeyFN       string
	Logger                   log.Interface
}

// API is the user-facing API serving the frontend assets and incoming client api calls
type API struct {
	server                   http.Server
	logger                   log.Interface
	serverCertificateChainFN string
	serverPrivateKeyFN       string
}

// New instantiates a new user-api
func New(options *Options) (*API, error) {

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.NewPkAuth().Handle)

	login := login.New(options.Logger)
	login.Register(router)

	return &API{
		logger:                   options.Logger,
		serverCertificateChainFN: options.ServerCertificateChainFN,
		serverPrivateKeyFN:       options.ServerPrivateKeyFN,
		server: http.Server{
			Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
			Handler: router,
			TLSConfig: &tls.Config{
				ServerName: options.Host,
				MinVersion: tls.VersionTLS13,
			},
		},
	}, nil
}

// Start blocks while accepting incoming connections. returns an error when done
func (a *API) Start() error {
	return a.server.ListenAndServeTLS(a.serverCertificateChainFN, a.serverPrivateKeyFN)
}