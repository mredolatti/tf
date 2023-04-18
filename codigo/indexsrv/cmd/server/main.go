package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	conf "github.com/mredolatti/tf/codigo/common/config"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"

	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/apis"
	"github.com/mredolatti/tf/codigo/indexsrv/config"
	"github.com/mredolatti/tf/codigo/indexsrv/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/mongodb"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/psql"
	"github.com/mredolatti/tf/codigo/indexsrv/repository/redis"

	goredis "github.com/redis/go-redis/v9"
)

func main() {

	cfg := parseEnvVars()

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

	serverRegistrar := registrar.New(&registrar.Config{
		FileServers:        repo.FileServers(),
		UserAccounts:       repo.Accounts(),
		Organizations:      repo.Organizations(),
		Pauth2Flows:        repo.PendingOAuth(),
		RootCAFN:           cfg.Server.RootCAFn,
		ServerCertFN:       cfg.Server.CertChainFn,
		ServerPrivateKeyFN: cfg.Server.PrivateKeyFn,
		BaseURL:            fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
	})

	fsLinks, err := fslinks.New(logger, repo.Users(), repo.Organizations(), repo.FileServers(), serverRegistrar, cfg.Server.RootCAFn)
	if err != nil {
		logger.Error("error setting up file-server links: %s", err)
		os.Exit(1)
	}

	sessionCache, err := setupSessionCache(repo.Users(), &cfg.Redis)
	if err != nil {
		logger.Error("error setting up session cache: %s", err)
		os.Exit(1)
	}

	apiBundle, err := apis.Setup(&apis.Config{
		Logger:      logger,
		UserManager: authentication.NewUserManager(repo.Users(), sessionCache, logger),
		Mapper: mapper.New(mapper.Config{
			LastUpdateTolerance: 1 * time.Hour,
			Repo:                repo.Mappings(),
			Users:               repo.Users(),
			Accounts:            repo.Accounts(),
			ServerLinks:         fsLinks,
		}),
		Server:          cfg.Server,
		ServerRegistrar: serverRegistrar,
	})

	if err != nil {
		logger.Error("error constructing user-facing API: %s", err)
		os.Exit(1)
	}

	go func() {
		time.Sleep(1 * time.Second)
		err := apiBundle.ListenAndServe()
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

func setupSessionCache(usersRepo repository.UserRepository, redisCfg *conf.Redis) (repository.SessionRepository, error) {
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
		DB:   redisCfg.DB,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("error setting up redis connection: %w", err)
	}

	return redis.NewSessionRepository(redisClient), nil
}

func setupShutdown(rtm runtime.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		rtm.Unblock()
	}()
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
		Redis: conf.Redis{
			Host: os.Getenv("IS_REDIS_HOST"),
			Port: conf.IntOr(os.Getenv("IS_REDIS_PORT"), 6379),
			DB:   conf.IntOr(os.Getenv("IS_REDIS_DB"), 0),
		},
	}
}
