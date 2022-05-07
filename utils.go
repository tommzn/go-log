package log

import (
	"context"
	"os"
)

// WithNameSpace appends passed namespace as log context
func WithNameSpace(logger Logger, namespace string) Logger {
	return appendContextValues(logger, map[string]string{LogCtxNamespace: namespace})
}

// WithK8sContext appends kubernetes values from environment variables as context
// At the moment following environment variables are supported:
//	K8S_NODE_NAME 	- Node name
//	K8S_POD_NAME	- Pod name
func WithK8sContext(logger Logger) Logger {

	logContextValues := make(map[string]string)
	if node, ok := os.LookupEnv("K8S_NODE_NAME"); ok {
		logContextValues[LogCtxK8sNode] = node
	}
	if pod, ok := os.LookupEnv("K8S_POD_NAME"); ok {
		logContextValues[LogCtxK8sPod] = pod
	}
	return appendContextValues(logger, logContextValues)
}

// WithNameSpace appends passed passed values to log context
func appendContextValues(logger Logger, values map[string]string) Logger {
	if logHandler, ok := logger.(*LogHandler); ok {
		logHandler.context.AppendValues(values)
		return logHandler
	}
	logger.WithContext(LogContextWithValues(context.Background(), values))
	return logger
}
