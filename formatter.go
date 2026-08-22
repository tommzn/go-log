package log

import (
	"fmt"
)

// newDefaultFormatter returns a new DefaultFormatter.
func newDefaultFormatter() LogFormatter {
	return &DefaultFormatter{}
}

// Format converts passed log level and context using Sprintf and return a complete string together with passed message.
func (formatter *DefaultFormatter) Format(logLevel LogLevel, logContext LogContext, message string) string {
	return fmt.Sprintf("%s: %s, Context: %+v", logLevel, message, logContext)
}
