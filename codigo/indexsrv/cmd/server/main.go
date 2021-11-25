package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
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

	setupShutdown(rtm)
	rtm.Block()
}

func setupShutdown(rtm runtime.Interface) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigs
		rtm.Unblock()
	}()
}
