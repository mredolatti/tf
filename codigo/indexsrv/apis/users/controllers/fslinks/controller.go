package fslinks

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
)

// Controller used to initiale an oauth2 flow to setup a user account in this server with one in a file server
type Controller struct {
	logger log.Interface
	reg    registrar.Interface
}

// New constructs a controller
func New(logger log.Interface, reg registrar.Interface) *Controller {
	return &Controller{
		logger: logger,
		reg:    reg,
	}
}

// Register mounts the provided endpoints
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/accounts/server/:serverId/authorize", c.initialRedirect)
	router.GET("/accounts/auth_callback", c.callback)
}

func (c *Controller) initialRedirect(ctx *gin.Context) {

	// TODO(mredolatti): get proper userID here
	userID := "107156877088323945674"

	serverID := ctx.Param("serverId")
	force := ctx.Query("force") == "true"

	url, err := c.reg.InitiateLinkProcess(ctx.Request.Context(), userID, serverID, force)
	if err != nil {
		if errors.Is(err, registrar.ErrAccountExists) {
			ctx.JSON(400, "account already exists")
			c.logger.Error("requested initial link with an already existing account (%s/%s)", userID, serverID)
			return
		}
		ctx.JSON(500, "unable to initiate oauth2 flow")
		c.logger.Error("error initiating oauth2 flow: %s", err)
		return
	}

	ctx.Redirect(301, url)
}

func (c *Controller) callback(ctx *gin.Context) {
	state := ctx.Query("state")
	code := ctx.Query("code")

	err := c.reg.CompleteLinkProcess(ctx.Request.Context(), state, code)
	if err != nil {
		c.logger.Error("error handling auth code: %w", err)
		ctx.JSON(500, "internal error when handling auth code")
		return
	}

	ctx.JSON(200, "account linked successfully")
}
