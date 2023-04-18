package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
)

// GetUserOrAbort gets the user set by the middleware previously, or otherwise aborts the request
func GetUserOrAbort(ctx *gin.Context) (string, bool) {
	user := ctx.GetString("user")
	if user == "" {
		ctx.AbortWithStatusJSON(500, jsend.NewErrorResponse("internal error while identifying user"))
		return "", false
	}
	return user, true
}
