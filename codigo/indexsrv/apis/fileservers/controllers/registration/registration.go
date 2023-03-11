package registration

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/fileservers/middleware"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
)

var (
	errServerCNMismatch = errors.New("server common-name mismatch")
)

type Controller struct {
	logger   log.Interface
	registry registrar.Interface
}

func New(logger log.Interface, registry registrar.Interface) *Controller {
	return &Controller{
		logger:   logger,
		registry: registry,
	}
}

func (c *Controller) Register(router gin.IRouter) {
	router.GET("/test", func(*gin.Context) {})
	router.POST("/register", c.registerServer)
}

func (c *Controller) registerServer(ctx *gin.Context) {
	var dto dtos.ServerInfoDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("error reading body from service registration request: %s", err.Error())
		ctx.AbortWithStatus(400)
		return
	}

	commonName, err := middleware.ServerCommonNameFromContext(ctx)
	if err != nil {
		c.logger.Error("failed to get common-name from TLS params: ", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	if err := c.registry.RegisterServer(
		ctx.Request.Context(),
		dto.OrgID,
		commonName,
		dto.AuthURL,
		dto.TokenURL,
		dto.FetchURL,
		dto.ControlEndpoint,
	); err != nil {
		c.logger.Error("error registering server: %s", err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.Status(200)
}
