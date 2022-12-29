package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	conf "github.com/mredolatti/tf/codigo/common/config"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"

	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users"
	"github.com/mredolatti/tf/codigo/indexsrv/config"
	"github.com/mredolatti/tf/codigo/indexsrv/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/mongodb"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/psql"
)

func main() {

	cfg := parseEnvVars()

	fmt.Printf("%+v\n", cfg)

	var logLevel = log.Info
	if cfg.Debug {
		logLevel = log.Debug
	}
	logger, err := log.New(os.Stdout, logLevel)
	if err != nil {
		fmt.Println("Error inicializando logger: ", err)
		os.Exit(1)
	}

	rtm, err := runtime.New(logger)
	if err != nil {
		logger.Error("Error inicializando runtime: %s", err)
		os.Exit(1)
	}

	repo, err := setupRepositories(cfg)
	if err != nil {
		logger.Error("Error setting up repositories: %s", err)
		os.Exit(1)
	}

	tlsConfig := parseTLSConfig(&cfg.Server)

	serverRegistrar := registrar.New(repo.FileServers(), repo.Accounts(), repo.PendingOAuth(), tlsConfig)

	fsLinks, err := fslinks.New(logger, repo.Users(), repo.Organizations(), repo.FileServers(), serverRegistrar, cfg.Server.RootCAFn)
	if err != nil {
		logger.Error("error setting up file-server links: %s", err)
		os.Exit(1)
	}

	userAPI, err := users.New(&users.Options{
		Host:                cfg.Server.Host,
		Port:                cfg.Server.Port,
		GoogleCredentialsFn: cfg.GoogleCredentialsFn,
		Logger:              logger,
		UserManager:         authentication.NewUserManager(repo.Users()),
		Mapper: mapper.New(mapper.Config{
			LastUpdateTolerance: 1 * time.Hour,
			Repo:                repo.Mappings(),
			Accounts:            repo.Accounts(),
			ServerLinks:         fsLinks,
		}),
		ServerRegistrar: serverRegistrar,
	})
	if err != nil {
		logger.Error("error constructing user-facing API: %s", err)
		os.Exit(1)
	}

	go func() {
		time.Sleep(1 * time.Second)
		err := userAPI.Start()
		if err != nil {
			fmt.Println("HTTP server error: ", err)
		}
		rtm.Unblock()
	}()

	setupShutdown(rtm)
	rtm.Block()
}

func setupRepositories(cfg *config.Main) (repository.Factory, error) {
	switch strings.ToLower(cfg.DBEngine) {
	case "mongo":
		return mongodb.NewFactory(&cfg.Mongo)
	case "postgres":
		return psql.NewFactory(&cfg.Postgres)
	default:
		return nil, fmt.Errorf("unknown db-engine: %s", cfg.DBEngine)
	}

}

func setupShutdown(rtm runtime.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		rtm.Unblock()
	}()
}

func parseTLSConfig(cfg *conf.Server) *tls.Config {
	certBytes, err := ioutil.ReadFile(cfg.RootCAFn)
	if err != nil {
		panic("cannot read root certificate file: " + err.Error())
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(certBytes)

	certs, err := tls.LoadX509KeyPair(cfg.CertChainFn, cfg.PrivateKeyFn)
	if err != nil {
		panic("cannot read server certficate chain / private key files: " + err.Error())
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
	}
}

func parseEnvVars() *config.Main {
	return &config.Main{
		Debug:               os.Getenv("IS_LOG_DEBUG") == "true",
		DBEngine:            os.Getenv("IS_DB_ENGINE"),
		GoogleCredentialsFn: os.Getenv("IS_GOOGLE_CREDS_FN"),
		Server: conf.Server{
			Host:         os.Getenv("IS_HOST"),
			Port:         conf.IntOr(os.Getenv("IS_PORT"), 9876),
			RootCAFn:     os.Getenv("IS_ROOT_CA"),
			CertChainFn:  os.Getenv("IS_SERVER_CERT_CHAIN"),
			PrivateKeyFn: os.Getenv("IS_SERVER_PRIVATE_KEY"),
		},
		Mongo: conf.Mongo{
			Hosts:    conf.StringListOr(os.Getenv("IS_MONGO_HOSTS"), nil),
			User:     os.Getenv("IS_MONGO_USERNAME"),
			Password: os.Getenv("IS_MONGO_PASSWORD"),
			DB:       os.Getenv("IS_MONGO_DATABASE"),
		},
		Postgres: conf.Postgres{
			Host:     os.Getenv("IS_PG_HOST"),
			Port:     conf.IntOr(os.Getenv("IS_PG_PORT"), 5432),
			User:     os.Getenv("IS_PG_USER"),
			Password: os.Getenv("IS_PG_PWD"),
			DB:       os.Getenv("IS_PG_DB"),
		},
	}
}

/*
	return &config{
		debug:               os.Getenv("IS_LOG_DEBUG") == "true",
		host:                os.Getenv("IS_HOST"),
		port:                intOr(os.Getenv("IS_PORT"), 9876),
		dbEngine:            os.Getenv("IS_DB_ENGINE"),
		mongoHost:           os.Getenv("IS_HOST"),
		mongoPort:           intOr(os.Getenv("IS_PORT"), 27017),
		mongoUser:           os.Getenv("IS_USERNAME"),
		mongoPassword:       os.Getenv("IS_PASSWORD"),
		mongoDB:             os.Getenv("IS_MONGO_DATABASE"),
		postgresHost:        os.Getenv("IS_PG_HOST"),
		postgresPort:        intOr(os.Getenv("IS_PG_PORT"), 5432),
		postgresUser:        os.Getenv("IS_PG_USER"),
		postgresPassword:    os.Getenv("IS_PG_PWD"),
		postgresDB:          os.Getenv("IS_PG_DB"),
		googleCredentialsFn: os.Getenv("IS_GOOGLE_CREDS_FN"),
		rootCAFn:            os.Getenv("IS_ROOT_CA"),
		certChainFn:         os.Getenv("IS_SERVER_CERT_CHAIN"),
		privateKeyFn:        os.Getenv("IS_SERVER_PRIVATE_KEY"),
	}
}
*/
