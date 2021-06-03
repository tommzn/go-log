package log

import (
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
)

// LOG_CONTEXT_KEY is used to set and get context values.
const LOG_CONTEXT_KEY = "log-context"

// newLogContext returns a new log context with passed key/value pairs.
func newLogContext(values map[string]string) LogContext {
	return LogContext{values: values}
}

// LogContextWithValues adds passed log values to passed context.
func LogContextWithValues(ctx context.Context, values map[string]string) context.Context {

	logContext := newLogContext(values)
	if currentLogContext, ok := ctx.Value(LOG_CONTEXT_KEY).(LogContext); ok {
		logContext = logContext.AppendValues(currentLogContext.values)
	}
	return context.WithValue(ctx, LOG_CONTEXT_KEY, logContext)
}

// getLogContext will extract log context from passed context.
// If there's no log context it will return an empty one.
func getLogContext(ctx context.Context) LogContext {

	if logContext, ok := ctx.Value(LOG_CONTEXT_KEY).(LogContext); ok {
		return logContext
	}
	return newLogContext(make(map[string]string))
}

// AppendValues reads values from current log context, append passed values and returns a new log context.
func (logContext LogContext) AppendValues(values map[string]string) LogContext {

	currentValues := logContext.values
	for key, value := range values {
		currentValues[key] = value
	}
	return newLogContext(currentValues)
}

// String creates a string representation of internal values map.
func (logContext LogContext) String() string {

	values := []string{}
	for key, val := range logContext.values {
		values = append(values, fmt.Sprintf("%s:%s", key, val))
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}

// DefaultContextForNodes returns a log context with hostname and ip (v4).
func DefaultContextForNodes() LogContext {

	values := make(map[string]string)
	host, _ := os.Hostname()
	values[LogCtx_Hostname] = host
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			values[LogCtx_Ip] = ipv4.String()
		}
	}
	return newLogContext(values)
}

// DefaultContextForK8s returns a log context with kubernetes node and pod name
// if both has benn set as K8S_NODE_NAME and K8S_POD_NAME during deployment.
func DefaultContextForK8s() LogContext {

	values := make(map[string]string)
	if node, ok := os.LookupEnv("K8S_NODE_NAME"); ok {
		values[LogCtx_K8s_Node] = node
	}
	if pod, ok := os.LookupEnv("K8S_POD_NAME"); ok {
		values[LogCtx_K8s_Pod] = pod
	}
	return newLogContext(values)
}
