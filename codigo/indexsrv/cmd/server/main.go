package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users"
	"github.com/mredolatti/tf/codigo/indexsrv/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/psql"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {

	config := parseEnvVars()

	fmt.Printf("%+v\n", config)

	var logLevel = log.Info
	if config.debug {
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

	db, err := setupDB(config.postgresUser, config.postgresPassword, config.postgresHost, config.postgresPort, config.postgresDB)
	if err != nil {
		logger.Error("error setting up databse: %s", err)
		os.Exit(1)
	}

	fsLinks := setupFSLinks(logger, db, config.rootCAFn)

	tlsConfig := parseTLSConfig(config)
	serverRegistrar := setupRegistrar(logger, db, tlsConfig)

	userAPI, err := users.New(&users.Options{
		Host:                config.host,
		Port:                config.port,
		GoogleCredentialsFn: config.googleCredentialsFn,
		Logger:              logger,
		UserManager:         setupUserManager(db),
		Mapper:              setupMappingManager(db, fsLinks),
		ServerRegistrar:     serverRegistrar,
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

func setupDB(user string, password string, host string, port int, db string) (*sqlx.DB, error) {
	// TODO: parametrize this properly!
	return sqlx.Connect("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, db))
}

func setupUserManager(db *sqlx.DB) authentication.UserManager {
	repo, _ := psql.NewUserRepository(db)
	return authentication.NewUserManager(repo)
}

func setupFSLinks(logger log.Interface, db *sqlx.DB, rootCAFn string) fslinks.Interface {
	userRepo, _ := psql.NewUserRepository(db)
	orgRepo, _ := psql.NewOrganizationRepository(db)
	accountRepo, _ := psql.NewUserAccountRepository(db)
	serversRepo, _ := psql.NewFileServerRepository(db)
	toRet, _ := fslinks.New(logger, userRepo, orgRepo, serversRepo, accountRepo, rootCAFn)
	return toRet
}

func setupMappingManager(db *sqlx.DB, fsLinks fslinks.Interface) *mapper.Impl {
	mappingRepo, _ := psql.NewMappingRepository(db)
	accountRepo, _ := psql.NewUserAccountRepository(db)

	return mapper.New(mapper.Config{
		LastUpdateTolerance: 1 * time.Hour,
		Repo:                mappingRepo,
		Accounts:            accountRepo,
		ServerLinks:         fsLinks,
	})
}

func setupRegistrar(logger log.Interface, db *sqlx.DB, tlsConfig *tls.Config) *registrar.Impl {
	serversRepo, _ := psql.NewFileServerRepository(db)
	accountRepo, _ := psql.NewUserAccountRepository(db)
	oauth2Flows, _ := psql.NewPendingOAuth2Repository(db)
	return registrar.New(serversRepo, accountRepo, oauth2Flows, tlsConfig)
}

func setupShutdown(rtm runtime.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		rtm.Unblock()
	}()
}

type config struct {
	debug               bool
	host                string
	port                int
	postgresHost        string
	postgresPort        int
	postgresUser        string
	postgresPassword    string
	postgresDB          string
	googleCredentialsFn string
	rootCAFn            string
	certChainFn         string
	privateKeyFn        string
}

func parseTLSConfig(c *config) *tls.Config {
	certBytes, err := ioutil.ReadFile(c.rootCAFn)
	if err != nil {
		panic("cannot read root certificate file: " + err.Error())
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(certBytes)

	certs, err := tls.LoadX509KeyPair(c.certChainFn, c.privateKeyFn)
	if err != nil {
		panic("cannot read server certficate chain / private key files: " + err.Error())
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
	}
}

func parseEnvVars() *config {
	return &config{
		debug:               os.Getenv("IS_LOG_DEBUG") == "true",
		host:                os.Getenv("IS_HOST"),
		port:                intOr(os.Getenv("IS_PORT"), 9876),
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

// TODO(mredolatti): mover a commons
func intOr(num string, fallback int) int {
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return fallback
	}
	return parsed
}
