package controllers

import (
	"github.com/gin-gonic/gin"
)

// GetUserOrAbort gets the user set by the middleware previously, or otherwise aborts the request
func GetUserOrAbort(ctx *gin.Context) (string, bool) {
	user := ctx.GetString("user")
	if user == "" {
		ctx.AbortWithStatus(500)
		return "", false
	}
	return user, true
}
