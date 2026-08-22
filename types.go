package log

// LogLevel defines the log level, e.g. error or debug.
type LogLevel int

const (
	// None diables logging.
	None LogLevel = iota * 100
	// Status can be used to log messages.
	Status
	// Error is a log level for errors.
	Error
	// Info is a log level for status information.
	Info
	// Debug is a log level for dev outputs.
	Debug
)

const (
	// LogCtxRequestId is a context key used for a request id.
	LogCtxRequestId = "requestid"
	// LogCtxLogLevel is a context key for used log level.
	LogCtxLogLevel = "loglevel"
	// LogCtxMessage is a context key for log messages.
	LogCtxMessage = "message"
	// LogCtxTimestamp is a context key for a timestamp.
	LogCtxTimestamp = "timestamp"
	// LogCtxNamespace is a context key for a namespace.
	LogCtxNamespace = "namespace"
	// LogCtxDomain is a context key for a domain.
	LogCtxDomain = "domain"
	// LogCtxHostname is a context keyfor a hostname.
	LogCtxHostname = "hostname"
	// LogCtxIp is a context key for an Ip (v4).
	LogCtxIp = "ip"
	// LogCtxK8sNode is a context key for kubernetes node name.
	LogCtxK8sNode = "k8s_node"
	// LogCtxK8sPod is a context key for a kubernetes pod name.
	LogCtxK8sPod = "k8s_pod"
)

// LogContext provides context values for logging.
type LogContext struct {
	values map[string]string
}

// DefaultFormatter is a fallback formatter to convert log values into a message.
type DefaultFormatter struct {
}

// StdoutShipper will print given log messages on stdout.
type StdoutShipper struct {
}
