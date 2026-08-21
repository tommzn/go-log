package log

import (
	"encoding/json"
	"fmt"
	"maps"
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
// Builds its own copy of the context values rather than writing into logContext.values
// directly - that map is the same instance a *LogHandler keeps for every message it logs,
// so mutating it here would race with concurrent log calls on the same logger.
func (formatter *LogzioJsonFormatter) format(logLevel LogLevel, logContext LogContext, message string) string {

	ctxValues := make(map[string]string, len(logContext.values)+3)
	maps.Copy(ctxValues, logContext.values)
	ctxValues[LogCtxLogLevel] = logLevel.String()
	ctxValues["@timestamp"] = time.Now().UTC().Format(LOGZIO_TIMESTAMP_FORMAT)
	ctxValues[LogCtxMessage] = message

	logContent, err := json.Marshal(ctxValues)
	if err != nil {
		return fmt.Sprintf(`{"message":%q,"loglevel":%q,"error":"json marshal failed"}`, message, logLevel.String())
	}
	return string(logContent)
}
