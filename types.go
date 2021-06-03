package log

import (
	"time"

	secrets "github.com/tommzn/go-secrets"
)

// LogLevel defines the log level, e.g. error or debug.
type LogLevel int

const (
	// None diables logging.
	None LogLevel = iota * 100
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

// LogzioJsonFormatter will convert passed values to a JSON record suitlable for an import at Logz.io.
type LogzioJsonFormatter struct {
}

// StdoutShipper will print given log messages on stdout.
type StdoutShipper struct {
}

// LogzioShipper will deliver log messages to Logz.io.
type LogzioShipper struct {

	// LogzioUrl is the enpooint all logs will be shipped to.
	logzioUrl string

	// BatchSize defines the number of logs shipped together in a batch.
	batchSize int

	// ShipmentStack is a worker queue to restrict parallel shipment.
	shipmentStack chan bool

	// MessageStack is a channel to buffer log messages.
	messageStack chan string

	// ObtainShipmentTimeout defines the time the shipper will wait to get
	// a slot from shipmentStack.
	obtainShipmentTimeout time.Duration

	// MessageReadTimeout defines the time a shipper will wait for new messages
	// during reading from messageStack.
	messageReadTimeout time.Duration

	// HttpClient is used to send POST request to ship log messages.
	httpClient httpClient

	// SecretsManager is used to obtain Logz.io token for shipment requests.
	secretsManager secrets.SecretsManager
}
