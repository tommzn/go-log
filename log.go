// Package log provides a logger with different formatter and shipper.
package log

import (
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
		context:   newEmptyLogContext(),
		formatter: formatter,
		shipper:   shipper,
	}
}

// NewLoggerFromConfig returns a new logger created depending on passed config.
// The shipper is picked based on the "log.shipper" config value - see
// RegisterShipper for how non-stdout shippers (e.g. Logz.io) plug in.
func NewLoggerFromConfig(conf config.Config, secretsManager secrets.SecretsManager) Logger {

	formatter, shipper := resolveShipperFromConfig(conf, secretsManager)
	return &LogHandler{
		logLevel:  LogLevelFromConfig(conf),
		context:   newEmptyLogContext(),
		formatter: formatter,
		shipper:   shipper,
	}
}
