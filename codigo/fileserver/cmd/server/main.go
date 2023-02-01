package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/fileserver/api/client"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"github.com/mredolatti/tf/codigo/fileserver/api/server"
	"github.com/mredolatti/tf/codigo/fileserver/authz"
	basicAuthz "github.com/mredolatti/tf/codigo/fileserver/authz/basic"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
	"github.com/mredolatti/tf/codigo/fileserver/repository/psql"
	"github.com/mredolatti/tf/codigo/fileserver/storage/basic"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {

	cfg := parseEnvVars()

	logLevel := log.Info
	if cfg.debug {
		logLevel = log.Debug
	}

	logger, err := log.New(os.Stdout, logLevel)
	if err != nil {
		panic(err.Error())
	}

	rtm, err := runtime.New(logger)
	if err != nil {
		panic(err.Error())
	}

	db, err := setupDB(cfg.postgresUser, cfg.postgresPassword, cfg.postgresHost, cfg.postgresPort, cfg.postgresDB)
	if err != nil {
		logger.Error("error setting up databse: %s", err)
		os.Exit(1)
	}

	oauth2W := setupOAuth2Wrapper(db, logger, cfg.jwtSecret)

	metaStore := basic.NewInMemoryFileMetadataStore()
	m1, _ := metaStore.Create("file1.jpg", "some note", "some_patient_id", "image", 1646394925714181390)
	m2, _ := metaStore.Create("file2.jpg", "some note", "some_patient_id", "image", 1646394925714181390)
	m3, _ := metaStore.Create("file3.jpg", "some note", "some_patient_id", "image", 1646394925714181390)
	m4, _ := metaStore.Create("file4.jpg", "some note", "some_patient_id", "image", 1646394925714181390)
	m5, _ := metaStore.Create("file5.jpg", "some note", "some_patient_id", "image", 1646394925714181390)

	fileStore := basic.NewInMemoryFileStore()
	fileStore.Write(m1.ID(), []byte("some data 1"), true)
	fileStore.Write(m2.ID(), []byte("some data 2"), true)
	fileStore.Write(m3.ID(), []byte("some data 3"), true)
	fileStore.Write(m4.ID(), []byte("some data 4"), true)
	fileStore.Write(m5.ID(), []byte("some data 5"), true)

	authorization := basicAuthz.NewInMemoryAuthz()
	authorization.Grant("martin.redolatti", authz.Create, authz.AnyObject)
	authorization.Grant("martin.redolatti", authz.Admin|authz.Write|authz.Read, m1.ID())
	authorization.Grant("martin.redolatti", authz.Admin|authz.Write|authz.Read, m2.ID())
	authorization.Grant("martin.redolatti", authz.Admin|authz.Write|authz.Read, m3.ID())
	authorization.Grant("martin.redolatti", authz.Admin|authz.Write|authz.Read, m4.ID())
	authorization.Grant("martin.redolatti", authz.Admin|authz.Write|authz.Read, m5.ID())

	fm := filemanager.New(fileStore, metaStore, authorization)

	// Client API -- consumed by end-users to interact with files
	clientAPI, err := client.New(&client.Options{
		Logger:                   logger,
		OAuht2Wrapper:            oauth2W,
		FileManager:              fm,
		Host:                     cfg.host,
		Port:                     cfg.clientAPIPort,
		ServerCertificateChainFN: cfg.serverCertChain,
		ServerPrivateKeyFN:       cfg.serverPrivateKey,
		RootCAFn:                 cfg.rootCA,
	})
	if err != nil {
		panic(err.Error())
	}

	go func() {
		time.Sleep(1 * time.Second)
		err := clientAPI.Start()
		if err != nil {
			fmt.Println("HTTP server error: ", err)
		}
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
	if err != nil {
		panic(err.Error())
	}

	go func() {
		time.Sleep(1 * time.Second)
		err := serverAPI.Start()
		if err != nil {
			fmt.Println("gRPC server error: ", err)
		}
	}()
	rtm.Block()
}

func setupDB(user string, password string, host string, port int, db string) (*sqlx.DB, error) {
	// TODO: parametrize this properly!
	return sqlx.Connect("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, db))
}

func setupOAuth2Wrapper(db *sqlx.DB, logger log.Interface, jwtSecret string) *oauth2.Impl {
	clientRepo, _ := psql.NewClientRepository(db)
	tokenRepo, _ := psql.NewTokenInfoRepository(db)

	oauth2W, err := oauth2.New(logger, "user", clientRepo, tokenRepo, []byte(jwtSecret))
	if err != nil {
		panic(err.Error())
	}

	return oauth2W
}

type config struct {
	debug            bool
	host             string
	clientAPIPort    int
	serverAPIPort    int
	serverCertChain  string
	serverPrivateKey string
	rootCA           string
	jwtSecret        string
	postgresHost     string
	postgresPort     int
	postgresUser     string
	postgresPassword string
	postgresDB       string
}

func parseEnvVars() *config {
	return &config{
		debug:            os.Getenv("FS_LOG_DEBUG") == "true",
		host:             os.Getenv("FS_HOST"),
		clientAPIPort:    intOr(os.Getenv("FS_CLIENT_PORT"), 9877),
		serverAPIPort:    intOr(os.Getenv("FS_SERVER_PORT"), 9000),
		serverCertChain:  os.Getenv("FS_SERVER_CERT_CHAIN"),
		serverPrivateKey: os.Getenv("FS_SERVER_PRIVATE_KEY"),
		rootCA:           os.Getenv("FS_ROOT_CA"),
		jwtSecret:        os.Getenv("FS_JWT_SECRET"),
		postgresHost:     os.Getenv("FS_PG_HOST"),
		postgresPort:     intOr(os.Getenv("FS_PG_PORT"), 5432),
		postgresUser:     os.Getenv("FS_PG_USER"),
		postgresPassword: os.Getenv("FS_PG_PWD"),
		postgresDB:       os.Getenv("FS_PG_DB"),
	}
}

func intOr(num string, fallback int) int {
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return fallback
	}
	return parsed
}
