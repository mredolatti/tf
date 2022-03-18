package users

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/login"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/ui"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"

	"github.com/gin-gonic/gin"
)

// Options contins user-api configuration parameters
type Options struct {
	Host                string
	Port                int
	GoogleCredentialsFn string
	UserManager         authentication.UserManager
	Mapper              mapper.Interface
	ServerRegistrar     registrar.Interface
	Logger              log.Interface
}

// API is the user-facing API serving the frontend assets and incoming client api calls
type API struct {
	ui     *ui.Controller
	login  *login.Controller
	server http.Server
}

// New instantiates a new user-api
func New(options *Options) (*API, error) {

	router := gin.New()
	router.Use(gin.Recovery())

	// TODO: Cambiar esto a postgres o redis
	store := memstore.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	loginController, err := login.New(options.UserManager, options.Logger, options.GoogleCredentialsFn)
	if err != nil {
		return nil, fmt.Errorf("error instantiating login controller: %w", err)
	}
	loginController.Register(router)

	uiController := ui.New(options.Logger, options.Mapper)
	uiController.Register(router)

	oauth2Controller := fslinks.New(options.Logger, options.ServerRegistrar)
	oauth2Controller.Register(router)

	return &API{
		ui:    uiController,
		login: loginController,
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
