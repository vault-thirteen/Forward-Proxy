package server

import (
	"log"
	"strings"

	"github.com/rs/zerolog"
)

const (
	LogLevelDebug         = "debug"
	LogLevelInformation   = "info"
	LogLevelWarning       = "warn"
	LogLevelError         = "error"
	LogLevelFatal         = "fatal"
	LogLevelPanic         = "panic"
	LogLevelNone          = "none"
	LogLevelDisableLogger = "disabled"

	LogLevelDefault = LogLevelError
)

type LogLevelSetting struct {
	TextParameter string
	ZerologValue  zerolog.Level
}

var logLevelSettings = []LogLevelSetting{
	{TextParameter: LogLevelDebug, ZerologValue: zerolog.DebugLevel},
	{TextParameter: LogLevelInformation, ZerologValue: zerolog.InfoLevel},
	{TextParameter: LogLevelWarning, ZerologValue: zerolog.WarnLevel},
	{TextParameter: LogLevelError, ZerologValue: zerolog.ErrorLevel},
	{TextParameter: LogLevelFatal, ZerologValue: zerolog.FatalLevel},
	{TextParameter: LogLevelPanic, ZerologValue: zerolog.PanicLevel},
	{TextParameter: LogLevelNone, ZerologValue: zerolog.NoLevel},
	{TextParameter: LogLevelDisableLogger, ZerologValue: zerolog.Disabled},
}

func SetLogLevel(logLevelParameter string) {
	llpLC := strings.ToLower(logLevelParameter)

	for _, lls := range logLevelSettings {
		if llpLC == strings.ToLower(lls.TextParameter) {
			zerolog.SetGlobalLevel(lls.ZerologValue)
			log.Println("Log level: ", lls.TextParameter)
			return
		}
	}

	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func possibleLogLevelsHint() string {
	var hint strings.Builder
	for _, lls := range logLevelSettings {
		hint.WriteString("[" + lls.TextParameter + "] ")
	}
	return hint.String()
}
