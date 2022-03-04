package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/fslinks"
)

// Controller bundling endpoints for file-server check-in
type Controller struct {
	logger  log.Interface
	servers fslinks.Interface
}

// Register mounts controller endpoints in supplied gin router
func (c *Controller) Register(router gin.IRouter) {
	router.POST("/status", c.updateStatus)
}

func (c *Controller) updateStatus(ctx *gin.Context) {
	var status StatusDTO
	err := ctx.BindJSON(&status)
	if err != nil {
		ctx.JSON(400, "unable to parse JSON in request body")
		c.logger.Error("error parsing request body: ", err)
		return
	}

	err = c.servers.NotifyServerUp(ctx.Request.Context(), status.ServerID)
	if err != nil {
		ctx.JSON(500, "error updating server status")
		c.logger.Error("error updating server status: ", err)
	}
}
