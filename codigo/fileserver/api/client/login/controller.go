package login

import (
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"

	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
)

// Controller implements authorization/token fetching endpoints for offline oauth2 login
type Controller struct {
	logger        log.Interface
	oauth2Wrapper oauth2.Interface
}

// New constructs a new controller
func New(logger log.Interface, oauth2Wrapper oauth2.Interface) *Controller {
	return &Controller{
		logger:        logger,
		oauth2Wrapper: oauth2Wrapper,
	}
}

// Register mounts the login endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/authorize", c.authorize)
	router.GET("/token", c.token)
	router.POST("/token", c.token)
}

func (c *Controller) authorize(ctx *gin.Context) {
	err := c.oauth2Wrapper.HandleAuthCodeRequest(ctx)
	if err != nil {
		c.logger.Error("error handling oauth2 authorization request: %s", err)
		ctx.AbortWithStatus(500)
	}
}

func (c *Controller) token(ctx *gin.Context) {

    // there's a specific error being returned internally for expired refresh tokens,
    // but it's internally replaced by a more generic before the function returns.
    // clients will have treat any refresh error as fatal and request a re-link of the
    // user account if the refresh token ever expires
	err := c.oauth2Wrapper.HandleAuthCodeExchangeRequest(ctx)
	if err != nil {
		c.logger.Error(err.Error())
		ctx.AbortWithStatus(401)
	}
}
