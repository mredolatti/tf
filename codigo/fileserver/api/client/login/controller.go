package login

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

const (
	ctxUser ctxKey = iota + 100
)

// Controller implements authorization/token fetching endpoints for offline oauth2 login
type Controller struct {
	logger            log.Interface
	oauth2Manager     *manage.Manager
	oauth2TokenStore  oauth2.TokenStore
	oauth2ClientStore oauth2.ClientStore
	oauth2Server      *server.Server
}

// New constructs a new controller
func New(logger log.Interface) *Controller {
	manager := manage.NewDefaultManager()
	tokenStore, err := store.NewMemoryTokenStore()
	if err != nil {
		// TODO
		panic(err.Error())
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

	return &Controller{
		logger:            logger,
		oauth2Manager:     manager,
		oauth2TokenStore:  tokenStore,
		oauth2ClientStore: clientStore,
		oauth2Server:      srv,
	}
}

// Register mounts the login endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/authorize", c.authorize)
	router.GET("/token", c.token)
}

func (c *Controller) authorize(ctx *gin.Context) {
	user, _ := ctx.Get("user") // will be valited later down the chain
	err := c.oauth2Server.HandleAuthorizeRequest(
		ctx.Writer,
		ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxUser, user)),
	)
	if err != nil {
		c.logger.Error("error handling oauth2 authorization request: %s", err)
		ctx.AbortWithStatus(500)
	}
}

func (c *Controller) token(ctx *gin.Context) {
	gt, tkr, err := c.oauth2Server.ValidationTokenRequest(ctx.Request)
	fmt.Println(gt, err)
	fmt.Println("code:", tkr)
	c.oauth2Server.HandleTokenRequest(ctx.Writer, ctx.Request)
}
