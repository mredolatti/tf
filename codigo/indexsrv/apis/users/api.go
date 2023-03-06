package users

import (
	"fmt"
	"net/http"

	//	"github.com/gin-contrib/sessions"
	//	"github.com/gin-contrib/sessions/memstore"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/login"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/mappings"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/ui"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
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
	SessionManager      authentication.SessionManager
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

	// Setup authentication middleware
	samw := middleware.NewSessionAuth(options.SessionManager, options.Logger)

	// Setup login controller
	loginController := login.New(options.UserManager, options.SessionManager, samw, options.Logger)
	loginController.Register(router)


	// Setup session-protected api group 
	protected := router.Group("/")
	protected.Use(samw.Handle)
	mappingController := mappings.New(options.Logger, options.Mapper)
	mappingController.Register(protected)
	oauth2Controller := fslinks.New(options.Logger, options.ServerRegistrar)
	oauth2Controller.Register(protected)

	return &API{
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
