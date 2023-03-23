package server

import (
	"flag"
	"time"
)

type Parameters struct {
	LogLevel string
	Host     string
	Port     uint16

	// Content.
	MustDecodeGzip bool
	MustRemoveBOM  bool

	// Timeouts.
	TargetConnectionTimeoutSec uint
	targetConnectionTimeout    time.Duration
}

const (
	HostDefault                       = "0.0.0.0"
	PortDefault                       = 8080
	TargetConnectionTimeoutSecDefault = 60
	MustDecodeGzipDefault             = false
	MustRemoveBOMDefault              = true
)

func ReadParameters() (p *Parameters, err error) {
	logLevelFlag := flag.String(
		"loglevel",
		LogLevelDefault,
		"log level; possible values: "+possibleLogLevelsHint(),
	)

	hostFlag := flag.String("host", HostDefault, "listen host name")
	portFlag := flag.Uint("port", PortDefault, "listen port number")
	mustDecodeGzipFlag := flag.Bool("gzip", MustDecodeGzipDefault, "decode GZip content")
	mustRemoveBOMFlag := flag.Bool("bom", MustRemoveBOMDefault, "remove BOM from content")
	targetConnectionTimeoutSecFlag := flag.Uint("tct", TargetConnectionTimeoutSecDefault, "target connection timeout in seconds")
	flag.Parse()

	p = &Parameters{
		LogLevel:       *logLevelFlag,
		Host:           *hostFlag,
		Port:           uint16(*portFlag),
		MustDecodeGzip: *mustDecodeGzipFlag,
		MustRemoveBOM:  *mustRemoveBOMFlag,
	}

	// Timeouts.
	p.TargetConnectionTimeoutSec = *targetConnectionTimeoutSecFlag
	p.targetConnectionTimeout = time.Second * time.Duration(p.TargetConnectionTimeoutSec)

	return p, nil
}
