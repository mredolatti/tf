package ui

import (
	"github.com/mredolatti/tf/codigo/indexsrv/frontend"

	"github.com/gin-gonic/gin"
)

// Controller serves endpoints that render ui pages
type Controller struct {
}

// Register mounts the endpoints onto the supplied router
func (c *Controller) Register(router gin.IRouter) {
	router.GET("/", c.main)
}

func (c *Controller) main(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")
	ctx.String(200, string(frontend.Index()))
}
