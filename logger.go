package log

import (
	"context"
	"fmt"
)

// LogHandler provides methods to log messges with different log level
// and takes care about formatting and shipping logs.
type LogHandler struct {
	logLevel  LogLevel
	context   LogContext
	formatter LogFormatter
	shipper   LogShipper
}

// logf will format given log message.
func (logger *LogHandler) logf(logLevel LogLevel, message string, v ...interface{}) {
	logger.log(logLevel, fmt.Sprintf(message, v...))
}

// log will create a log message with given values.
func (logger *LogHandler) log(logLevel LogLevel, v ...interface{}) {

	if logger.logLevel >= logLevel {
		logger.shipper.send(logger.formatter.format(logLevel, logger.context, fmt.Sprint(v...)))
	}
}

// WithContext applies the log context.
func (logger *LogHandler) WithContext(ctx context.Context) {
	logger.context = getLogContext(ctx)
}

// Errorf format given log message for log level Error.
func (logger *LogHandler) Errorf(message string, v ...interface{}) {
	logger.logf(Error, message, v...)
}

// Error will create a log message with given values for log level Error.
func (logger *LogHandler) Error(v ...interface{}) {
	logger.log(Error, v...)
}

// Infof format given log message for log level Info.
func (logger *LogHandler) Infof(message string, v ...interface{}) {
	logger.logf(Info, message, v...)
}

// Info will create a log message with given values for log level Info.
func (logger *LogHandler) Info(v ...interface{}) {
	logger.log(Info, v...)
}

// Debugf format given log message for log level Debug.
func (logger *LogHandler) Debugf(message string, v ...interface{}) {
	logger.logf(Debug, message, v...)
}

// Debug will create a log message with given values for log level Debug.
func (logger *LogHandler) Debug(v ...interface{}) {
	logger.log(Debug, v...)
}

// Flush will force it's log shipper to deliver all remaining log messages.
func (logger *LogHandler) Flush() {
	logger.shipper.flush()
}
