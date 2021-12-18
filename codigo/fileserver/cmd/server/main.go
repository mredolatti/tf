package main

import (
	"os"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/fileserver/api/client"
)

func main() {

	logger, err := log.New(os.Stdout, log.Error)
	if err != nil {
		panic(err.Error())
	}

	rtm, err := runtime.New(logger)
	if err != nil {
		panic(err.Error())
	}

	api, err := client.New(&client.Options{
		Logger:                   logger,
		Host:                     "file-server",
		Port:                     9877,
		ServerCertificateChainFN: "/home/martin/projects/tf/codigo/PKI/fileserver/certs/chain.pem",
		ServerPrivateKeyFN:       "/home/martin/projects/tf/codigo/PKI/fileserver/private/fs_server.key",
		RootCAFn:                 "/home/martin/projects/tf/codigo/PKI/root/certs/ca.crt",
	})
	if err != nil {
		panic(err.Error())
	}

	go api.Start()

	rtm.Block()
}
