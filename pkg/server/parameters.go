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
	MustDecodeGzip                     bool
	MustRemoveBOM                      bool
	MustUseSpeedLimiter                bool
	SpeedLimiterNormalLimitBytesPerSec float64
	SpeedLimiterBurstLimitBytesPerSec  int
	SpeedLimiterMaxBNR                 float64

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
	MustUseSpeedLimiterDefault        = true

	// SpeedLimiterNormalLimitBytesPerSecDefault is a default value of a normal
	// (average) speed limit in bytes per second.
	// Golang's 'io' package in Go version 1.20.2 uses 32 KiB chunks for
	// reading with io.Copy. It means that this parameter should not be less
	// than 32 KiB which is 32'768 b/s. Also note that in future versions of
	// Go language this behaviour may change.
	SpeedLimiterNormalLimitBytesPerSecDefault = 50_000 // 50 kByte/s.

	// SpeedLimiterBurstLimitBytesPerSecDefault is a default value of a burst
	// (short-term) speed limit in bytes per second.
	// Golang's 'io' package in Go version 1.20.2 uses 32 KiB chunks for
	// reading with io.Copy. It means that this parameter should not be less
	// than 32 KiB which is 32'768 b/s. Also note that in future versions of
	// Go language this behaviour may change.
	SpeedLimiterBurstLimitBytesPerSecDefault = 50_000 // 50 kByte/s.

	// SpeedLimiterMaxBNRDefault is a default value of maximal normal-to-burst
	// ratio (coefficient).
	SpeedLimiterMaxBNRDefault = 2.0
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
	mustDecodeGzipFlag := flag.Bool("gzip", MustDecodeGzipDefault, "decode GZip content")
	mustRemoveBOMFlag := flag.Bool("bom", MustRemoveBOMDefault, "remove BOM from content")
	mustUseSpeedLimiterFlag := flag.Bool("speed", MustUseSpeedLimiterDefault, "use speed limiter")
	speedLimiterNormalLimitBytesPerSec := flag.Float64("slnl", SpeedLimiterNormalLimitBytesPerSecDefault, "speed limiter's normal limit (kb/s)")
	speedLimiterBurstLimitBytesPerSec := flag.Int("slbl", SpeedLimiterBurstLimitBytesPerSecDefault, "speed limiter's burst limit (kb/s)")
	speedLimiterMaxBNR := flag.Float64("slbnr", SpeedLimiterMaxBNRDefault, "speed limiter's maximal burst-to-normal ratio")
	flag.Parse()

	p = &Parameters{
		LogLevel: *logLevelFlag,
		Host:     *hostFlag,
		Port:     uint16(*portFlag),

		MustDecodeGzip:                     *mustDecodeGzipFlag,
		MustRemoveBOM:                      *mustRemoveBOMFlag,
		MustUseSpeedLimiter:                *mustUseSpeedLimiterFlag,
		SpeedLimiterNormalLimitBytesPerSec: *speedLimiterNormalLimitBytesPerSec,
		SpeedLimiterBurstLimitBytesPerSec:  *speedLimiterBurstLimitBytesPerSec,
		SpeedLimiterMaxBNR:                 *speedLimiterMaxBNR,
	}

	// Timeouts.
	p.TargetConnectionTimeoutSec = *targetConnectionTimeoutSecFlag
	p.targetConnectionTimeout = time.Second * time.Duration(p.TargetConnectionTimeoutSec)

	return p, nil
}
