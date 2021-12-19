package middleware

import (
	"crypto/x509"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
)

// PKAuth is a public-key authentication middleware
type PKAuth struct {
	logger log.Interface
}

// NewPkAuth returns a new instance of a PKAuth middleware
func NewPkAuth(logger log.Interface) *PKAuth {
	return &PKAuth{logger: logger}
}

// Handle is the function to be called by gin to validate provided PK
func (a *PKAuth) Handle(ctx *gin.Context) {

	var clientCertficate *x509.Certificate
	for _, cert := range ctx.Request.TLS.PeerCertificates {
		if cert != nil || !cert.IsCA {
			clientCertficate = cert
			break
		}
	}

	if clientCertficate == nil {
		a.logger.Error("no valid certificate provided by the client")
		ctx.AbortWithStatus(401)
	}

	a.logger.Debug("found valid certificate for: ", clientCertficate.Subject.CommonName)
	ctx.Set("user", clientCertficate.Subject.CommonName)

	ctx.Next()
}
