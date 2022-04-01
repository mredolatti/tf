package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mredolatti/tf/codigo/fileserver/api/client/files"
	"github.com/mredolatti/tf/codigo/fileserver/api/client/login"
	"github.com/mredolatti/tf/codigo/fileserver/api/client/middleware"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"

	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
)

// Options contains user-api configuration parameters
type Options struct {
	Host                     string
	Port                     int
	ServerCertificateChainFN string
	ServerPrivateKeyFN       string
	RootCAFn                 string
	OAuht2Wrapper            oauth2.Interface
	FileManager              filemanager.Interface
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
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.NewPkAuth(options.Logger).Handle)

	login := login.New(options.Logger, options.OAuht2Wrapper)
	login.Register(router)

	files := files.New(options.Logger, options.FileManager)
	files.Register(router)

	certBytes, err := ioutil.ReadFile(options.RootCAFn)
	if err != nil {
		panic(err.Error())
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certBytes)

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
				ClientCAs:  certPool,
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
		},
	}, nil
}

// Start blocks while accepting incoming connections. returns an error when done
func (a *API) Start() error {
	return a.server.ListenAndServeTLS(a.serverCertificateChainFN, a.serverPrivateKeyFN)
}
