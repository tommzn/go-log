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
	// LogCtx_RequestId is a context key used for a request id.
	LogCtx_RequestId = "requestid"
	// LogCtx_LogLevel is a context key for used log level.
	LogCtx_LogLevel = "loglevel"
	// LogCtx_Message is a context key for log messages.
	LogCtx_Message = "message"
	// LogCtx_Timestamp is a context key for a timestamp.
	LogCtx_Timestamp = "timestamp"
	// LogCtx_Namespace is a context key for a namespace.
	LogCtx_Namespace = "namespace"
	// LogCtx_Domain is a context key for a domain.
	LogCtx_Domain = "domain"
	// LogCtx_Hostname is a context keyfor a hostname.
	LogCtx_Hostname = "hostname"
	// LogCtx_Ip is a context key for an Ip (v4).
	LogCtx_Ip = "ip"
	// LogCtx_K8s_Node is a context key for kubernetes node name.
	LogCtx_K8s_Node = "k8s_node"
	// LogCtx_K8s_Pod is a context key for a kubernetes pod name.
	LogCtx_K8s_Pod = "k8s_pod"
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
