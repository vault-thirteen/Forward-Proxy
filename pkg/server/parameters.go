package server

import (
	"flag"
	"time"

	wm "github.com/vault-thirteen/Forward-Proxy/pkg/server/WorkMode"
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
	TargetConnectionDialTimeoutSec uint
	targetConnectionDialTimeout    time.Duration

	// Work mode.
	WorkModeString string
	WorkModeList   string
	workMode       *wm.WorkMode
}

const (
	HostDefault                           = "0.0.0.0"
	PortDefault                           = 8080
	TargetConnectionDialTimeoutSecDefault = 60
	MustDecodeGzipDefault                 = false
	MustRemoveBOMDefault                  = true
	MustUseSpeedLimiterDefault            = true

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
	mustRemoveBOMFlag := flag.Bool("bom", MustRemoveBOMDefault, "Remove BOM from content")
	mustDecodeGzipFlag := flag.Bool("gzip", MustDecodeGzipDefault, "Decode GZip content")
	hostFlag := flag.String("host", HostDefault, "Listen host name")
	workModeListFlag := flag.String("list", "", "Path to a list of IP addresses for the selected work mode")
	logLevelFlag := flag.String("loglevel", LogLevelDefault, "Log level; possible values: "+possibleLogLevelsHint())
	workModeStringFlag := flag.String("mode", wm.WorkModeStringDefault, "Work mode: public or private")
	portFlag := flag.Uint("port", PortDefault, "Listen port number")
	mustUseSpeedLimiterFlag := flag.Bool("sl", MustUseSpeedLimiterDefault, "Use speed limiter")
	speedLimiterBurstLimitBytesPerSec := flag.Int("slbl", SpeedLimiterBurstLimitBytesPerSecDefault, "Speed limiter's burst limit (b/sec)")
	speedLimiterMaxBNR := flag.Float64("slbnr", SpeedLimiterMaxBNRDefault, "Speed limiter's maximal burst-to-normal ratio")
	speedLimiterNormalLimitBytesPerSec := flag.Float64("slnl", SpeedLimiterNormalLimitBytesPerSecDefault, "Speed limiter's normal limit (b/sec)")
	targetConnectionDialTimeoutSecFlag := flag.Uint("tcdt", TargetConnectionDialTimeoutSecDefault, "Target connection dial timeout (sec)")

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
		WorkModeString:                     *workModeStringFlag,
		WorkModeList:                       *workModeListFlag,
	}

	// Timeouts.
	p.TargetConnectionDialTimeoutSec = *targetConnectionDialTimeoutSecFlag
	p.targetConnectionDialTimeout = time.Second * time.Duration(p.TargetConnectionDialTimeoutSec)

	// Work mode.
	p.workMode, err = wm.New(p.WorkModeString, p.WorkModeList)
	if err != nil {
		return nil, err
	}

	return p, nil
}
