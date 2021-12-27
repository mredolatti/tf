package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
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
	ValidateAccess(ctx *gin.Context) (string, error)
}

// Impl is a wrapper around a set of helpers that handle oauth2 auth code & token requests
type Impl struct {
	logger      log.Interface
	userCtxKey  string
	manager     *manage.Manager
	tokenStore  oauth2.TokenStore
	clientStore oauth2.ClientStore
	server      *server.Server
}

// New constructs a new OAuth2 wrapper
func New(logger log.Interface, userContextKey string) (*Impl, error) {
	manager := manage.NewDefaultManager()
	tokenStore, err := store.NewMemoryTokenStore()
	if err != nil {
		// TODO
		return nil, fmt.Errorf("error instantiating token storage: %w", err)
	}
	manager.MapTokenStorage(tokenStore)

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		fmt.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		fmt.Println("Response Error:", re.Error.Error())
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

var _ Interface = (*Impl)(nil)
