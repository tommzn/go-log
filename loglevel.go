package log

import (
	"os"
	"strings"

	config "github.com/tommzn/go-config"
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

// SyslogLevel returns corresponding syslog(3) log level.
func (logLevel LogLevel) SyslogLevel() int {

	switch logLevel {
	case Error:
		return 3 // LOG_ERR
	case Info:
		return 6 // LOG_INFO
	case Debug:
		return 7 // LOG_DEBUG
	default:
		return 0 // LOG_EMERG
	}
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

// LogLevelFromConfig reads log level from config using "log.loglevel" as key.
// If there's no log level defined in passed config default Error will be returned.
func LogLevelFromConfig(conf config.Config) LogLevel {
	if logLevelName := conf.Get("log.loglevel", nil); logLevelName != nil {
		return LogLevelByName(*logLevelName)
	}
	return Error
}
