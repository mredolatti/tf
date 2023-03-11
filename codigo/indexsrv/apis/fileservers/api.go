package fileservers

import (
	"crypto/tls"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/fileservers/controllers/registration"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/fileservers/middleware"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"

	"github.com/gin-gonic/gin"
)

// Config contins user-api configuration parameters
type Config struct {
	Logger          log.Interface
	TLSConfig       *tls.Config
	ServerRegistrar registrar.Interface
}

func Mount(router gin.IRouter, config *Config) {
	tlsMW := middleware.NewTLSClientValidator(config.Logger, config.TLSConfig)
	router.Use(tlsMW.Handle)
	servregController := registration.New(config.Logger, config.ServerRegistrar)
	servregController.Register(router)
}
