package log

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// WithNameSpace returns a new Logger with passed namespace appended as log context.
// The passed logger is left unchanged.
func WithNameSpace(logger Logger, namespace string) Logger {
	return logger.WithFields(map[string]string{LogCtxNamespace: namespace})
}

// WithK8sContext returns a new Logger with kubernetes values from environment variables
// appended as context. The passed logger is left unchanged.
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
	return logger.WithFields(logContextValues)
}

// AppendContextValues returns a new Logger with passed values added to its context.
// The passed logger is left unchanged.
func AppendContextValues(logger Logger, values map[string]string) Logger {
	return logger.WithFields(values)
}

// AppendFromLambdaContext returns a new Logger with some values from given context,
// e.g. a request id, added to its log context if passed context is an AWS Lambda context.
// The passed logger is left unchanged.
func AppendFromLambdaContext(logger Logger, ctx context.Context) Logger {
	if lambdaCtx, ok := lambdacontext.FromContext(ctx); ok {
		return logger.WithFields(map[string]string{LogCtxRequestId: lambdaCtx.AwsRequestID})
	}
	return logger
}
