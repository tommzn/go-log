package log

import (
	"os"
	"strings"
)

// ENV_LOGLEVEL defines the environment variable which can be used to set a log level.
const ENV_LOGLEVEL = "LOGLEVEL"

// logLevelNames is a map of log level names set at init.
var logLevelNames map[LogLevel]string

// Init create a map with log level names.
func init() {
	logLevelNames = make(map[LogLevel]string)
	logLevelNames[None] = "None"
	logLevelNames[Status] = "Status"
	logLevelNames[Error] = "Error"
	logLevelNames[Info] = "Info"
	logLevelNames[Debug] = "Debug"
}

// String returns the name of a log level.
func (logLevel LogLevel) String() string {
	return logLevelNames[logLevel]
}

// LogLevelByName will try to convert passed name of a log level
// into a log level.
// If there's no suitable log level for a given name, log level None is returned, which disables logging.
func LogLevelByName(logLevelName string) LogLevel {

	switch strings.ToLower(logLevelName) {
	case "status":
		return Status
	case "error":
		return Error
	case "info":
		return Info
	case "debug":
		return Debug
	default:
		return None
	}
}

// LogLevelFromEnv will lookup for environment variable defined by ENV_LOGLEVEL and if
// it exists call LogLevelByName to convert it's value to a log level.
func LogLevelFromEnv() LogLevel {
	return LogLevelByName(os.Getenv(ENV_LOGLEVEL))
}
