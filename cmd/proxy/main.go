package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vault-thirteen/Forward-Proxy/pkg/server"
	ver "github.com/vault-thirteen/Versioneer"
)

const ExitCodeBadParameters = 1

func main() {
	showIntro()

	var err error
	var parameters *server.Parameters
	parameters, err = server.ReadParameters()
	if err != nil {
		fmt.Println("Use '-h' command line argument to get help.")
		fmt.Println()
		fmt.Println(err)
		os.Exit(ExitCodeBadParameters)
	}

	server.SetLogLevel(parameters.LogLevel)

	log.Println("Server is starting ...")
	var srv *server.Server
	srv, err = server.NewServer(parameters)
	mustBeNoError(err)

	err = srv.Start()
	mustBeNoError(err)
	fmt.Println("HTTP Server: " + srv.GetListenDsn())

	serverMustBeStopped := srv.GetStopChannel()
	waitForQuitSignalFromOS(serverMustBeStopped)
	<-*serverMustBeStopped

	log.Println("Stopping the server ...")
	err = srv.Stop()
	if err != nil {
		log.Println(err)
	}
	log.Println("Server was stopped.")
	time.Sleep(time.Second)
}

func mustBeNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func showIntro() {
	versioneer, err := ver.New()
	mustBeNoError(err)
	versioneer.ShowIntroText("")
	versioneer.ShowComponentsInfoText()
	fmt.Println()
}

func waitForQuitSignalFromOS(serverMustBeStopped *chan bool) {
	osSignals := make(chan os.Signal, 16)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range osSignals {
			switch sig {
			case syscall.SIGINT,
				syscall.SIGTERM:
				log.Println("quit signal from OS has been received: ", sig)
				*serverMustBeStopped <- true
			}
		}
	}()
}
