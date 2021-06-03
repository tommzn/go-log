package log

import (
	"encoding/json"
	"fmt"
	"time"
)

// newDefaultFormatter returns a new DefaultFormatter.
func newDefaultFormatter() LogFormatter {
	return &DefaultFormatter{}
}

// format converts passed log level and context using Sprintf and return a complete string together with passed message.
func (formatter *DefaultFormatter) format(logLevel LogLevel, logContext LogContext, message string) string {
	return fmt.Sprintf("%s: %s, Context: %+v", logLevel, message, logContext)
}

// newLogzioJsonFormatter returns a new LogzioJsonFormatter.
func newLogzioJsonFormatter() LogFormatter {
	return &LogzioJsonFormatter{}
}

// format composes passed log level, context and message in a map and marshal it to JSON.
func (formatter *LogzioJsonFormatter) format(logLevel LogLevel, logContext LogContext, message string) string {

	ctxValues := logContext.values
	ctxValues[LogCtxLogLevel] = logLevel.String()
	ctxValues["@timestamp"] = time.Now().UTC().Format(LOGZIO_TIMESTAMP_FORMAT)
	ctxValues[LogCtxMessage] = message

	// Since we marshal string values only here, we'll omit the error
	logContent, _ := json.Marshal(ctxValues)
	return string(logContent)
}
