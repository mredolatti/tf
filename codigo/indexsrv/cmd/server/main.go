package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/runtime"
	"github.com/mredolatti/tf/codigo/indexsrv/apis/users"
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

	userAPI, err := users.New(&users.Options{
		Host: "0.0.0.0",
		Port: 9876,
	})
	if err != nil {
		logger.Error("error constructing user-facing API: %s", err)
		os.Exit(1)
	}

	go userAPI.Start()

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
