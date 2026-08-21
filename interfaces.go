package log

import (
	"context"
	"net/http"
)

// Logger is an infterface for different types of logger.
type Logger interface {

	// WithContext returns a new Logger with the log context from the given
	// context.Context merged into its existing context (values from ctx win
	// on key conflicts) - fields already attached via WithFields or similar
	// are preserved, not discarded. The receiver is left unchanged, so it's
	// safe to call concurrently on a logger shared across goroutines - assign
	// the result rather than relying on WithContext to mutate the logger in
	// place.
	WithContext(context.Context) Logger

	// WithFields returns a new Logger with the given key/value pairs merged
	// into its log context. The receiver is left unchanged.
	WithFields(fields map[string]string) Logger

	// Statusf logs a formated message with log level Status.
	Statusf(message string, v ...interface{})

	// Status logs given message with log level Status.
	Status(v ...interface{})

	// Errorf logs a formated message with log level Error.
	Errorf(message string, v ...interface{})

	// Error logs given message with log level Error.
	Error(v ...interface{})

	// Infof logs a formated message with log level Info.
	Infof(message string, v ...interface{})

	// Info logs given message with log level Info.
	Info(v ...interface{})

	// Debugf logs a formated message with log level Debug.
	Debugf(message string, v ...interface{})

	// Debug logs given message with log level Debug.
	Debug(v ...interface{})

	// Logs a formated message with given log level.
	Logf(logLevel LogLevel, message string, v ...interface{})

	// Logs passed message with given log level.
	Log(logLevel LogLevel, v ...interface{})

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
