package login

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ErrInvalidClientSecret is returned when trying to construct a controller with invalid credentials
var ErrInvalidClientSecret = errors.New("oauth2 config cannot be nil")

// Controller serves endpoints that render ui pages
type Controller struct {
	logger      log.Interface
	userManager authentication.UserManager
	oauth2Conf  *oauth2.Config
	gpkFetcher  *googleKeyFetcher
	clientID    string
}

// New instantiates a new controller
func New(userManager authentication.UserManager, logger log.Interface, clientID string, secret string) (*Controller, error) {
	if clientID == "" || secret == "" {
		return nil, ErrInvalidClientSecret
	}

	googlePubKeyFetcher := newGoogleKeyFetcher(logger)
	go googlePubKeyFetcher.Run() // start fetching & updating google pub keys in BG

	return &Controller{
		userManager: userManager,
		logger:      logger,
		clientID:    clientID,
		oauth2Conf: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: secret,
			Endpoint:     google.Endpoint,
			// TODO: aceptar el host como parametro y buildear la URL
			RedirectURL: "http://localhost:9876/login/callback",
			Scopes:      []string{"openid", "email", "profile"},
		},
		gpkFetcher: googlePubKeyFetcher,
	}, nil
}

// Register mounts the endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/login", c.login)
	router.GET("/login/callback", c.loginCallback)
}

func (c *Controller) login(ctx *gin.Context) {
	// TODO: consider passing something other than "state"
	ctx.Redirect(301, c.oauth2Conf.AuthCodeURL("state", oauth2.AccessTypeOnline))
}

func (c *Controller) loginCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	token, err := c.oauth2Conf.Exchange(ctx.Request.Context(), code)
	if err != nil {
		c.logger.Error("error exchanging code for token: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.logger.Error("failed to extract id_token from raw token.")
		ctx.AbortWithStatus(500)
		return
	}

	claims, err := validateJWT(idToken, c.clientID, c.gpkFetcher)
	if err != nil {
		c.logger.Error("error validating JWT token: ", err)
		ctx.AbortWithStatus(500)
		return
	}

	// TODO: Ajustar esto
	// No se necesita refresh token mientras usemos online access type al autenticar
	// aun si usara offline access type, el refresh token no deberia cambiar

	_, err = c.userManager.CreateOrUpdate(
		context.Background(),
		claims.Subject,
		claims.FirstName+" "+claims.LastName,
		claims.Email,
		token.AccessToken,
		token.RefreshToken,
	)

	session := sessions.Default(ctx)
	session.Set("id", claims.Subject)
	session.Options(sessions.Options{Path: "/", MaxAge: 1800})
	session.Save()

	if err != nil {
		c.logger.Error("error creating/updating user: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	ctx.Redirect(301, "/main")
}
