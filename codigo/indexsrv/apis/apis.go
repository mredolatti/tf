package apis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mredolatti/tf/codigo/common/config"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/fileservers"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Server          config.Server
	UserManager     authentication.UserManager
	Mapper          mapper.Interface
	ServerRegistrar registrar.Interface
	Logger          log.Interface
}

type Bundle struct {
	server http.Server
}

func (b *Bundle) ListenAndServe() error {
	return b.server.ListenAndServeTLS("", "") // cert & key provided in tls.Config
}

func Setup(config *Config) (*Bundle, error) {

	router := gin.New()
	router.Use(gin.Recovery())
    router.Use(gin.Logger())
	clientAPI := router.Group("/api/clients/v1")
	users.Mount(clientAPI, &users.Config{
		UserManager:     config.UserManager,
		Mapper:          config.Mapper,
		ServerRegistrar: config.ServerRegistrar,
		Logger:          config.Logger,
	})

	tlsConfig := setupTLSConfig(&config.Server)
	fileServerAPI := router.Group("/api/fileservers/v1")
	fileservers.Mount(fileServerAPI, &fileservers.Config{
		ServerRegistrar: config.ServerRegistrar,
		Logger:          config.Logger,
		TLSConfig:       tlsConfig,
	})

	return &Bundle{
		server: http.Server{
			Addr:      fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
			Handler:   router,
			TLSConfig: tlsConfig,
		},
	}, nil
}

func setupTLSConfig(config *config.Server) *tls.Config {
	certBytes, err := ioutil.ReadFile(config.RootCAFn)
	if err != nil {
		panic("cannot read root certificate file: " + err.Error())
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(certBytes)

	certs, err := tls.LoadX509KeyPair(config.CertChainFn, config.PrivateKeyFn)
	if err != nil {
		panic("cannot read server certficate chain / private key files: " + err.Error())
	}

	return &tls.Config{
		ServerName:   config.Host,
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
		ClientAuth:   tls.RequestClientCert,
		ClientCAs:    caPool,
	}
}
