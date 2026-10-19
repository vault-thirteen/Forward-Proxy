package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	ver "github.com/vault-thirteen/auxie/Versioneer/classes/Versioneer"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"

	"github.com/vault-thirteen/Forward-Proxy/pkg/server"
)

const (
	ExitCode_BadParameters = 1
	ExitCode_Init          = 2
)

func main() {
	showIntro()

	var err error
	var parameters *server.Parameters
	parameters, err = server.ReadParameters()
	if err != nil {
		cio.PrintMessage("Use '-h' command line argument to get help.")
		cio.PrintMessage()
		errors.LogError(err)
		os.Exit(ExitCode_BadParameters)
	}

	server.SetLogLevel(parameters.LogLevel)

	cio.LogMessage("Server is starting ...")
	var srv *server.Server
	srv, err = server.NewServer(parameters)
	mustBeNoError(err)

	err = srv.Start()
	mustBeNoError(err)
	cio.PrintMessage("HTTP Server: " + srv.GetListenDsn())

	serverMustBeStopped := srv.GetStopChannel()
	waitForQuitSignalFromOS(serverMustBeStopped)
	<-*serverMustBeStopped

	cio.LogMessage("Stopping the server ...")
	err = srv.Stop()
	if err != nil {
		errors.LogError(err)
	}
	cio.LogMessage("Server was stopped.")
	time.Sleep(time.Second)
}

func mustBeNoError(err error) {
	if err != nil {
		errors.LogError(err)
		os.Exit(ExitCode_Init)
	}
}

func showIntro() {
	versioneer, err := ver.New(false)
	mustBeNoError(err)
	versioneer.ShowIntroText("")
	versioneer.ShowComponentsInfoText()
	cio.PrintMessage()
}

func waitForQuitSignalFromOS(serverMustBeStopped *chan bool) {
	osSignals := make(chan os.Signal, 16)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range osSignals {
			switch sig {
			case syscall.SIGINT,
				syscall.SIGTERM:
				cio.LogMessage("quit signal from OS has been received: ", sig)
				*serverMustBeStopped <- true
			}
		}
	}()
}
