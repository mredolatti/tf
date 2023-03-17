package middleware

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"

	"github.com/gin-gonic/gin"
)


const (
	ServerCNKey = "SERVER_CN"
)

var (
	ErrNoServerID = errors.New("no server id in request context")
	ErrInvalidServerID = errors.New("invalid server id in request context")
)



type TLSClientCertValidator struct {
	logger    log.Interface
	tlsConfig *tls.Config
}

func NewTLSClientValidator(logger log.Interface, tlsConfig *tls.Config) *TLSClientCertValidator {
	return &TLSClientCertValidator{
		logger:    logger,
		tlsConfig: tlsConfig,
	}
}

func (v *TLSClientCertValidator) Handle(ctx *gin.Context) {

	certs := ctx.Request.TLS.PeerCertificates
	if len(certs) == 0 {
		v.logger.Error("connecting server did not send a certificate. aborting")
		ctx.AbortWithStatusJSON(400, gin.H{"error": "this endpoint requires a client certificate to be sent"})
		return
	}

	opts := x509.VerifyOptions{
		Roots:         v.tlsConfig.ClientCAs,
		CurrentTime:   time.Now(),
		Intermediates: x509.NewCertPool(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	_, err := certs[0].Verify(opts)
	if err != nil {
		v.logger.Error("failed to verify client certificate")
		ctx.AbortWithStatusJSON(401, gin.H{"error": "client certificate could not be verified"})
		return
	}

	switch certs[0].PublicKey.(type) {
	case *ecdsa.PublicKey, *rsa.PublicKey, ed25519.PublicKey:
	default:
		v.logger.Error("client certificate contains an unsupported public key of type %T", certs[0].PublicKey)
		ctx.AbortWithStatusJSON(401, gin.H{"error": "unsupported public key type"})
		return
	}

	ctx.Set(ServerCNKey, certs[0].Subject.CommonName)
}

func ServerCommonNameFromContext(ctx *gin.Context) (string, error) {
	raw, exists := ctx.Get(ServerCNKey);
	if !exists {
		return "", ErrNoServerID
	}

	id, ok := raw.(string)
	if !ok {
		return "", ErrInvalidServerID
	}

	return id, nil
}
