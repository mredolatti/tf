package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// PKAuth is a public-key authentication middleware
type PKAuth struct{}

// NewPkAuth returns a new instance of a PKAuth middleware
func NewPkAuth() *PKAuth {
	return &PKAuth{}
}

// Handle is the function to be called by gin to validate provided PK
func (a *PKAuth) Handle(ctx *gin.Context) {
	fmt.Println("certificates: ", ctx.Request.TLS.PeerCertificates)
	ctx.Next()
}
