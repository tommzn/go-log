// Package log provides a logger with different formatter and shipper.
package log

import (
	"strings"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

// NewLogger returns a new logger with passed log level, formatter and shipper.
// If you omit formatter and shipper the DefaultFormatter and StdoutShipper will be used.
func NewLogger(logLevel LogLevel, formatter LogFormatter, shipper LogShipper) Logger {

	if formatter == nil {
		formatter = newDefaultFormatter()
	}
	if shipper == nil {
		shipper = newStdoutShipper()
	}
	return &LogHandler{
		logLevel:  logLevel,
		context:   LogContext{},
		formatter: formatter,
		shipper:   shipper,
	}
}

// NewLoggerFromConfig returns a new logger created depending on passed config.
func NewLoggerFromConfig(conf config.Config, secretsManager secrets.SecretsManager) Logger {

	var logLevel LogLevel
	var formatter LogFormatter
	var shipper LogShipper

	shipperType := conf.Get("log.shipper", nil)
	if shipperType != nil && strings.ToLower(*shipperType) == "logzio" {
		formatter = newLogzioJsonFormatter()
		shipper = newLogzioShipper(conf, secretsManager)
	} else {
		formatter = newDefaultFormatter()
		shipper = newStdoutShipper()
	}

	if logLevelName := conf.Get("log.loglevel", nil); logLevelName != nil {
		logLevel = LogLevelByName(*logLevelName)
	} else {
		logLevel = Error
	}
	return &LogHandler{
		logLevel:  logLevel,
		context:   newEmptyLogContext(),
		formatter: formatter,
		shipper:   shipper,
	}
}
