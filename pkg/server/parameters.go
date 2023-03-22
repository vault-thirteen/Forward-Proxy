package server

import (
	"flag"
	"time"
)

type Parameters struct {
	LogLevel string
	Host     string
	Port     uint16

	// Timeouts.
	TargetConnectionTimeoutSec uint
	targetConnectionTimeout    time.Duration
}

const (
	HostDefault                       = "0.0.0.0"
	PortDefault                       = 8080
	TargetConnectionTimeoutSecDefault = 60
)

func ReadParameters() (p *Parameters, err error) {
	logLevelFlag := flag.String(
		"loglevel",
		LogLevelDefault,
		"log level; possible values: "+possibleLogLevelsHint(),
	)

	hostFlag := flag.String("host", HostDefault, "listen host name")
	portFlag := flag.Uint("port", PortDefault, "listen port number")
	targetConnectionTimeoutSecFlag := flag.Uint("tct", TargetConnectionTimeoutSecDefault, "target connection timeout in seconds")
	flag.Parse()

	p = &Parameters{
		LogLevel: *logLevelFlag,
		Host:     *hostFlag,
		Port:     uint16(*portFlag),
	}

	// Timeouts.
	p.TargetConnectionTimeoutSec = *targetConnectionTimeoutSecFlag
	p.targetConnectionTimeout = time.Second * time.Duration(p.TargetConnectionTimeoutSec)

	return p, nil
}
