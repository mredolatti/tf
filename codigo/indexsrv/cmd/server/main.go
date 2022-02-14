package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
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

	userAPI, err := users.New(&users.Options{
		Host:                config.host,
		Port:                config.port,
		GoogleCredentialsFn: config.googleCredentialsFn,
		Logger:              logger,
		UserManager:         setupUserManager(db),
		Mapper:              setupMappingManager(db),
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

func setupMappingManager(db *sqlx.DB) *mapper.Impl {
	repo, _ := psql.NewMappingRepository(db)
	return mapper.New(repo)
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
