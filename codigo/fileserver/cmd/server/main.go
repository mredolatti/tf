package main

import (
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/fileserver/api/client"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"github.com/mredolatti/tf/codigo/fileserver/api/server"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
	"github.com/mredolatti/tf/codigo/fileserver/repository/psql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {

	cfg := parseEnvVars()

	logLevel := log.Info
	if cfg.debug {
		logLevel = log.Debug
	}

	logger, err := log.New(os.Stdout, logLevel)
	mustBeNil(err)

	rtm, err := runtime.New(logger)
	mustBeNil(err)

	fm, err := filemanager.Setup(cfg.storagePlugin, cfg.storagePluginConf)
	mustBeNil(err)

	oauth2W := setupOAuth2Wrapper(cfg.psqlURI, logger, cfg.jwtSecret)
	clientAPI, err := client.New(&client.Options{ // Client API -- consumed by end-users to interact with files
		Logger:                   logger,
		OAuht2Wrapper:            oauth2W,
		FileManager:              fm,
		Host:                     cfg.host,
		Port:                     cfg.clientAPIPort,
		ServerCertificateChainFN: cfg.serverCertChain,
		ServerPrivateKeyFN:       cfg.serverPrivateKey,
		RootCAFn:                 cfg.rootCA,
	})
	mustBeNil(err)

	go func() {
		time.Sleep(1 * time.Second)
		mustBeNil(clientAPI.Start())
		rtm.Unblock()
	}()

	serverAPI, err := server.New(&server.Options{
		Logger:                   logger,
		Port:                     cfg.serverAPIPort,
		FileManager:              fm,
		ServerCertificateChainFN: cfg.serverCertChain,
		ServerPrivateKeyFN:       cfg.serverPrivateKey,
		RootCAFn:                 cfg.rootCA,
		OAuth2Wrapper:            oauth2W,
	})
	mustBeNil(err)

	go func() {
		time.Sleep(1 * time.Second)
		mustBeNil(serverAPI.Start())
	}()

	rtm.Block() // block the main thread
}

func setupOAuth2Wrapper(dbURI string, logger log.Interface, jwtSecret string) *oauth2.Impl {
	db, err := sqlx.Connect("pgx", dbURI)
	mustBeNil(err)

	clientRepo, _ := psql.NewClientRepository(db)
	tokenRepo, _ := psql.NewTokenInfoRepository(db)

	oauth2W, err := oauth2.New(logger, "user", clientRepo, tokenRepo, []byte(jwtSecret))
	mustBeNil(err)

	return oauth2W
}

type config struct {
	debug             bool
	host              string
	clientAPIPort     int
	serverAPIPort     int
	serverCertChain   string
	serverPrivateKey  string
	rootCA            string
	jwtSecret         string
	psqlURI           string
	storagePlugin     string
	storagePluginConf string
}

func parseEnvVars() *config {
	return &config{
		debug:             os.Getenv("FS_LOG_DEBUG") == "true",
		host:              os.Getenv("FS_HOST"),
		clientAPIPort:     intOr(os.Getenv("FS_CLIENT_PORT"), 9877),
		serverAPIPort:     intOr(os.Getenv("FS_SERVER_PORT"), 9000),
		serverCertChain:   os.Getenv("FS_SERVER_CERT_CHAIN"),
		serverPrivateKey:  os.Getenv("FS_SERVER_PRIVATE_KEY"),
		rootCA:            os.Getenv("FS_ROOT_CA"),
		jwtSecret:         os.Getenv("FS_JWT_SECRET"),
		psqlURI:           os.Getenv("FS_PSQL_URI"),
		storagePlugin:     os.Getenv("FS_STORAGE_PLUGIN"),
		storagePluginConf: os.Getenv("FS_STORAGE_PLUGIN_CONF"),
	}
}

func intOr(num string, fallback int) int {
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return fallback
	}
	return parsed
}

func mustBeNil(e error) {
	if e != nil {
		panic(e.Error())
	}
}
