package login

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
)

// Controller serves endpoints that render ui pages
type Controller struct {
	logger         log.Interface
	userManager    authentication.UserManager
	sessionManager authentication.SessionManager
	authMW         *middleware.SessionAuth
}

func New(
	userManager authentication.UserManager,
	sessionManager authentication.SessionManager,
	authMW *middleware.SessionAuth,
	logger log.Interface,
) *Controller {
	return &Controller{
		userManager:    userManager,
		sessionManager: sessionManager,
		logger:         logger,
		authMW:         authMW,
	}
}

func (c *Controller) Register(router gin.IRouter) error {
	router.POST("/signup", c.signup)
	router.POST("/login", c.login)
	router.POST("/logout", c.authMW.Handle, c.logout) // TODO(mredolatti): Add middleware that checks for token
	return nil
}

func (c *Controller) signup(ctx *gin.Context) {
	var body userRegistrationDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.logger.Error("error parsing JSON in body: ", err)
		ctx.AbortWithStatus(400)
		return
	}

	_, err := c.userManager.Create(ctx.Request.Context(), body.NameField, body.EmailField, body.PasswordField)
	if err != nil {
		c.logger.Error("error creating user: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.Status(200)
}

func (c *Controller) login(ctx *gin.Context) {
	var body userLoginDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.logger.Error("error parsing JSON in body: ", err)
		ctx.AbortWithStatus(400)
		return
	}

	token, err := c.sessionManager.Create(ctx.Request.Context(), body.EmailField, body.PasswordField)
	if err != nil {
		if errors.Is(err, authentication.ErrInvalidCredentials) {
			ctx.AbortWithStatus(401)
			return
		}
		c.logger.Error("error creating session: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, &tokenDTO{token})
}

func (c *Controller) logout(ctx *gin.Context) {
	token, err := middleware.SessionTokenFromContext(ctx)
	fmt.Println("LOGGING OUT TOKEN: ", token)
	if err != nil {
		c.logger.Warning("Failed to get token from request when closing session: %s", err)
		ctx.AbortWithStatus(400)
	}

	if err := c.sessionManager.Revoke(ctx.Request.Context(), token); err != nil {
		c.logger.Error("error shutting down session: ", err)
		ctx.AbortWithStatus(500)
	}

	ctx.Status(200)
}
