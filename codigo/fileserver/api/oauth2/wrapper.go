package oauth2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

type ctxKey int

// Errors
var (
	ErrNoUserInContext = errors.New("no user found in request context")
)

const (
	ctxUser ctxKey = iota + 100
)

// Interface defines the set of methods required to handle oauth2 flows and validations
type Interface interface {
	HandleAuthCodeRequest(ctx *gin.Context) error
	HandleAuthCodeExchangeRequest(ctx *gin.Context) error
	HandleTokenRefreshRequest(ctx *gin.Context)

	// TODO(mredolatti): mover esto a un compoenente que encapsule todo lo de jwt
	ValidateAccess(ctx *gin.Context) (string, error)
	ValidateToken(token string) (*generates.JWTAccessClaims, error)
}

// Impl is a wrapper around a set of helpers that handle oauth2 auth code & token requests
type Impl struct {
	logger      log.Interface
	server      *server.Server
	userCtxKey  string
	manager     *manage.Manager
	tokenStore  oauth2.TokenStore
	clientStore oauth2.ClientStore
	jwtSecret   []byte
}

// New constructs a new OAuth2 wrapper
func New(
	logger log.Interface,
	userContextKey string,
	clientStore oauth2.ClientStore,
	tokenStore oauth2.TokenStore,
	jwtSecret []byte,
) (*Impl, error) {

	manager := manage.NewDefaultManager()
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", jwtSecret, jwt.SigningMethodHS512))
	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    6 * time.Hour,
		RefreshTokenExp:   720 * time.Hour, // 1 week
		IsGenerateRefresh: true,
	})
    manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    6 * time.Hour,
		RefreshTokenExp:   720 * time.Hour, // 1 week
		IsGenerateRefresh: true,
	})

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logger.Error("oauth2 internal error: %s", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		logger.Error("oauth2 response error:")
		logger.Error("%+v\n", re)
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		user, ok := r.Context().Value(ctxUser).(string)
		if !ok {
			return "", errors.ErrAccessDenied
		}
		return user, nil
	})
	return &Impl{
		logger:      logger,
		userCtxKey:  userContextKey,
		manager:     manager,
		clientStore: clientStore,
		server:      srv,
		jwtSecret:   jwtSecret,
	}, nil
}

// HandleAuthCodeRequest handles oauth flow initiation request
func (o *Impl) HandleAuthCodeRequest(ctx *gin.Context) error {
	user, ok := ctx.Get(o.userCtxKey)
	if !ok {
		return ErrNoUserInContext
	}

	ctxWithUser := context.WithValue(ctx.Request.Context(), ctxUser, user)
	if err := o.server.HandleAuthorizeRequest(ctx.Writer, ctx.Request.WithContext(ctxWithUser)); err != nil {
		return fmt.Errorf("error handling auth code request: %w", err)
	}

	return nil
}

// HandleAuthCodeExchangeRequest handles exchanging an authorization code for a token
func (o *Impl) HandleAuthCodeExchangeRequest(ctx *gin.Context) error {
	err := o.server.HandleTokenRequest(ctx.Writer, ctx.Request)
	if err != nil {
		return fmt.Errorf("error exchanging auth code for token: %w", err)
	}
	return nil
}

// HandleTokenRefreshRequest TODO
func (o *Impl) HandleTokenRefreshRequest(ctx *gin.Context) {
}

// ValidateAccess verifies the token supplied in the requests and either accepts it or rejects it
func (o *Impl) ValidateAccess(ctx *gin.Context) (string, error) {
	info, err := o.server.ValidationBearerToken(ctx.Request)
	if err != nil {
		return "", fmt.Errorf("error validanting access token: %w", err)
	}

	return info.GetUserID(), nil
}

// ValidateToken parses and verifies a token in string form
func (o *Impl) ValidateToken(token string) (*generates.JWTAccessClaims, error) {
	fmt.Println("parseando: ", token)
	parsed, err := jwt.ParseWithClaims(token, &generates.JWTAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("invalid signature method")
		}

		return o.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := parsed.Claims.(*generates.JWTAccessClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

var _ Interface = (*Impl)(nil)
