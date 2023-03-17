package users

import (
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/admin"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/login"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/controllers/mappings"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"

	"github.com/gin-gonic/gin"
)

// Config contins user-api configuration parameters
type Config struct {
	UserManager         authentication.UserManager
	Mapper              mapper.Interface
	ServerRegistrar     registrar.Interface
	Logger              log.Interface
}


func Mount(router gin.IRouter, config *Config) {
	// Setup authentication middlewares
	samw := middleware.NewSessionAuth(config.UserManager, config.Logger)
	tfamw := middleware.NewTFACheck(config.Logger)

	// Setup login controller
	loginController := login.New(config.UserManager, samw, config.Logger)
	loginController.Register(router)

	// Setup session-protected api group
	protected := router.Group("/")
	protected.Use(samw.Handle, tfamw.Handle)
	mappingController := mappings.New(config.Logger, config.Mapper)
	mappingController.Register(protected)
	fsLinksController := fslinks.New(config.Logger, config.ServerRegistrar)
	fsLinksController.Register(protected)

	// Setup admin-token protected endpoints
	// TODO(mredolatti): create and use middleware that authenticates an admin
	adminEndpoints := router.Group("/admin")
	adminController := admin.New(config.ServerRegistrar, config.Logger)
	adminController.Register(adminEndpoints)
}

