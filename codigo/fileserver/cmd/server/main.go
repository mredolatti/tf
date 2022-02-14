package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/fileserver/api/client"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"github.com/mredolatti/tf/codigo/fileserver/authz"
	basicAuthz "github.com/mredolatti/tf/codigo/fileserver/authz/basic"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
	"github.com/mredolatti/tf/codigo/fileserver/storage/basic"
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

	oauth2W, err := oauth2.New(logger, "user")
	if err != nil {
		panic(err.Error())
	}

	fileStore := basic.NewInMemoryFileStore()
	metaStore := basic.NewInMemoryFileMetadataStore()
	authorization := basicAuthz.NewInMemoryAuthz()
	authorization.Grant("martin.redolatti", authz.Create, authz.AnyObject)
	fm := filemanager.New(fileStore, metaStore, authorization)

	api, err := client.New(&client.Options{
		Logger:                   logger,
		OAuht2Wrapper:            oauth2W,
		FileManager:              fm,
		Host:                     cfg.host,
		Port:                     cfg.port,
		ServerCertificateChainFN: cfg.serverCertChain,
		ServerPrivateKeyFN:       cfg.serverPrivateKey,
		RootCAFn:                 cfg.rootCA,
	})
	if err != nil {
		panic(err.Error())
	}

	go func() {
		time.Sleep(1 * time.Second)
		err := api.Start()
		if err != nil {
			fmt.Println("HTTP server error: ", err)
		}
		rtm.Unblock()
	}()

	rtm.Block()
}

type config struct {
	debug            bool
	host             string
	port             int
	serverCertChain  string
	serverPrivateKey string
	rootCA           string
}

func parseEnvVars() *config {
	return &config{
		debug:            os.Getenv("FS_LOG_DEBUG") == "true",
		host:             os.Getenv("FS_HOST"),
		port:             intOr(os.Getenv("FS_PORT"), 9877),
		serverCertChain:  os.Getenv("FS_SERVER_CERT_CHAIN"),
		serverPrivateKey: os.Getenv("FS_SERVER_PRIVATE_KEY"),
		rootCA:           os.Getenv("FS_ROOT_CA"),
	}
}

func intOr(num string, fallback int) int {
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return fallback
	}
	return parsed
}
