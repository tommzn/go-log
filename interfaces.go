package log

import (
	"context"
	"net/http"
)

// Logger is an infterface for different types of logger.
type Logger interface {

	// WithContext sets a given log context.
	WithContext(context.Context)

	// Statusf logs a formated message with log level Status.
	Statusf(message string, v ...interface{})

	// Status logs given message with log level Status.
	Status(v ...interface{})

	// Errorf logs a formated message with log level Error.
	Errorf(message string, v ...interface{})

	// Error logs given message with log level Error.
	Error(v ...interface{})

	// Errorf logs a formated message with log level Info.
	Infof(message string, v ...interface{})

	// Error logs given message with log level Info.
	Info(v ...interface{})

	// Errorf logs a formated message with log level Debug.
	Debugf(message string, v ...interface{})

	// Error logs given message with log level Debug.
	Debug(v ...interface{})

	// FLush tells the log shipper to cleat it's internal message queue.
	Flush()
}

// LogShipper will take care of sending logs to a defined target.
type LogShipper interface {

	// Send will process given message. Depending on log shipper implementation
	// this can lead to an immediate shippment or a shiiper can queue messages
	// to deliver them in a batch.
	send(string)

	// Flush clear internal buffer.
	// Depending on the logger this can include writing to a remote destination.
	flush()
}

// LogFormatter will convert passed log values into a suitable log message.
type LogFormatter interface {

	// Format create a log message from given values.
	format(LogLevel, LogContext, string) string
}

// httpClient is an interface for a HTTP client.
type httpClient interface {

	// Do will send a http request.
	Do(req *http.Request) (*http.Response, error)
}
