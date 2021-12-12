package login

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
)

type Controller struct {
	logger log.Interface
}

func New(logger log.Interface) *Controller {
	return &Controller{logger: logger}
}

func (c *Controller) Register(router gin.IRouter) {
	router.GET("/authorize", c.authorize)
}

func (c *Controller) authorize(ctx *gin.Context) {
	fmt.Println("URL: ", *ctx.Request.URL)
}
