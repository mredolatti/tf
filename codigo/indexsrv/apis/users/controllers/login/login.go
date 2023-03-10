package login

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users/middleware"
)

// Controller serves endpoints that render ui pages
type Controller struct {
	logger      log.Interface
	userManager authentication.UserManager
	authMW      *middleware.SessionAuth
}

func New(
	userManager authentication.UserManager,
	authMW *middleware.SessionAuth,
	logger log.Interface,
) *Controller {
	return &Controller{
		userManager: userManager,
		logger:      logger,
		authMW:      authMW,
	}
}

func (c *Controller) Register(router gin.IRouter) error {
	router.POST("/signup", c.signup)
	router.POST("/login", c.login)
	router.POST("/logout", c.authMW.Handle, c.logout)
	router.POST("/2fa", c.authMW.Handle, c.setup2FA)
	return nil
}

func (c *Controller) signup(ctx *gin.Context) {
	var body userRegistrationDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.logger.Error("error parsing JSON in body: ", err)
		ctx.AbortWithStatus(400)
		return
	}

	_, err := c.userManager.Signup(ctx.Request.Context(), body.NameField, body.EmailField, body.PasswordField)
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

	token, err := c.userManager.Login(ctx.Request.Context(), body.EmailField, body.PasswordField, body.OTP)
	if err != nil {
		if errors.Is(err, authentication.ErrInvalidCredentials) || errors.Is(err, authentication.ErrInvalid2FAPasscode) {
			ctx.AbortWithStatus(401)
			return
		}
		c.logger.Error("error creating session: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.JSON(200, &tokenDTO{token})
}

func (c *Controller) setup2FA(ctx *gin.Context) {
	session, err := middleware.SessionFromContext(ctx)
	if err != nil {
		c.logger.Warning("Failed to get token from request when closing session: %s", err)
		ctx.AbortWithStatus(400)
		return
	}

	// TODO(mredolatti): offer recovery codes as well
	qr, _, err := c.userManager.Setup2FA(ctx, session.User())
	if err != nil {
		c.logger.Error("error setting up 2fa for user='%s': %s", session.User(), err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.Data(200, "image/png", qr.Bytes())

}

func (c *Controller) logout(ctx *gin.Context) {
	token, err := middleware.SessionTokenFromContext(ctx)
	if err != nil {
		c.logger.Warning("Failed to get token from request when closing session: %s", err)
		ctx.AbortWithStatus(400)
		return
	}

	if err := c.userManager.Logout(ctx.Request.Context(), token); err != nil {
		c.logger.Error("error shutting down session: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.Status(200)
}
