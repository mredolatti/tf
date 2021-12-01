package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/ui"
)

// Options contins user-api configuration parameters
type Options struct {
	Host string
	Port int
}

// API is the user-facing API serving the frontend assets and incoming client api calls
type API struct {
	ui     *ui.Controller
	server http.Server
}

// New instantiates a new user-api
func New(options *Options) (*API, error) {

	router := gin.New()
	router.Use(gin.Recovery())

	uiController := &ui.Controller{}
	uiController.Register(router)

	return &API{
		ui: uiController,
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
