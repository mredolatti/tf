package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	logger, err := log.New(os.Stdout, log.Debug)
	if err != nil {
		fmt.Println("Error inicializando logger: ", err)
		os.Exit(1)
	}

	rtm, err := runtime.New(logger)
	if err != nil {
		logger.Error("Error inicializando runtime: %s", err)
		os.Exit(1)
	}

	db, err := setupDB()
	if err != nil {
		logger.Error("error setting up databse: %s", err)
		os.Exit(1)
	}

	userAPI, err := users.New(&users.Options{
		Host:              "0.0.0.0",
		Port:              9876,
		OAuthClientID:     os.Getenv("GOOGLE_LOGIN_CLIENT_ID"),
		OAuthClientSecret: os.Getenv("GOOGLE_LOGIN_CLIENT_SECRET"),
		Logger:            logger,
		UserManager:       setupUserManager(db),
		Mapper:            setupMappingManager(db),
	})
	if err != nil {
		logger.Error("error constructing user-facing API: %s", err)
		os.Exit(1)
	}

	go userAPI.Start()

	setupShutdown(rtm)
	rtm.Block()
}

func setupDB() (*sqlx.DB, error) {
	// TODO: parametrize this properly!
	return sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
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
