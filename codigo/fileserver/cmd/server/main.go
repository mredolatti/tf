package main

import (
	"fmt"
	"os"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/fileserver/api/client"
	"github.com/mredolatti/tf/codigo/fileserver/api/oauth2"
	"github.com/mredolatti/tf/codigo/fileserver/authz"
	basicAuthz "github.com/mredolatti/tf/codigo/fileserver/authz/basic"
	"github.com/mredolatti/tf/codigo/fileserver/storage/basic"
)

func main() {

	logger, err := log.New(os.Stdout, log.Debug)
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

	can, err := authorization.Can("martin.redolatti", authz.Create, authz.AnyObject)
	fmt.Println("AAAA ", can)

	api, err := client.New(&client.Options{
		Logger:                   logger,
		OAuht2Wrapper:            oauth2W,
		Authorization:            authorization,
		FileStorage:              fileStore,
		FileMetaStorage:          metaStore,
		Host:                     "file-server",
		Port:                     9877,
		ServerCertificateChainFN: "./PKI/fileserver/certs/chain.pem",
		ServerPrivateKeyFN:       "./PKI/fileserver/private/fs_server.key",
		RootCAFn:                 "./PKI/root/certs/ca.crt",
	})
	if err != nil {
		panic(err.Error())
	}

	go api.Start()

	rtm.Block()
}
