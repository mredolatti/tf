package fileservers

import (
	"fmt"
	"net/http"

	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
)

// Options contins file-server-api configuration parameters
type Options struct {
	Host   string
	Port   int
	Logger log.Interface
}

// API is the server-facing API exposing file-server registration endpoints
type API struct {
	server http.Server
}

// New instantiates a new file-server-api
func New(options *Options) (*API, error) {

	router := gin.New()
	router.Use(gin.Recovery())

	return &API{
		server: http.Server{
			Addr:    fmt.Sprintf("%s:%d", options.Host, options.Port),
			Handler: router,
		},
	}, nil
}

// Start blocks while accepting incoming connections. returns an error when done
func (a *API) Start() error {
	return a.server.ListenAndServe()
}
