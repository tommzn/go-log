package log

import (
	"context"
	"fmt"
)

type LogHandler struct {
	logLevel  LogLevel
	context   LogContext
	formatter LogFormatter
	shipper   LogShipper
}

func (logger *LogHandler) logf(logLevel LogLevel, message string, v ...interface{}) {
	logger.log(logLevel, fmt.Sprintf(message, v...))
}

func (logger *LogHandler) log(logLevel LogLevel, v ...interface{}) {

	if logger.logLevel >= logLevel {
		logger.shipper.send(logger.formatter.format(logLevel, logger.context, fmt.Sprint(v...)))
	}
}

func (logger *LogHandler) WithContext(ctx context.Context) {
	logger.context = getLogContext(ctx)
}

func (logger *LogHandler) Errorf(message string, v ...interface{}) {
	logger.logf(Error, message, v...)
}

func (logger *LogHandler) Error(v ...interface{}) {
	logger.log(Error, v...)
}

func (logger *LogHandler) Infof(message string, v ...interface{}) {
	logger.logf(Info, message, v...)
}

func (logger *LogHandler) Info(v ...interface{}) {
	logger.log(Info, v...)
}

func (logger *LogHandler) Debugf(message string, v ...interface{}) {
	logger.logf(Debug, message, v...)
}

func (logger *LogHandler) Debug(v ...interface{}) {
	logger.log(Debug, v...)
}

func (logger *LogHandler) Flush() {
	logger.shipper.flush()
}
