package main

import (
	"os"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
)

func main() {
	logger := log.New(os.Stdout, log.Debug)
	rtm, err := runtime.New(logger)
	if err != nil {
		logger.Error("Error inicializando runtime: %s", err)
		os.Exit(1)
	}

	rtm.Block()
}
