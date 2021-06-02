package log

import (
	"time"

	secrets "github.com/tommzn/go-secrets"
)

type LogLevel int

const (
	None LogLevel = iota * 100
	Error
	Info
	Debug
)

const (
	LogCtx_RequestId = "requestid"
	LogCtx_LogLevel  = "loglevel"
	LogCtx_Message   = "message"
	LogCtx_Timestamp = "timestamp"
	LogCtx_Namespace = "namespace"
	LogCtx_Domain    = "domain"
	LogCtx_Hostname  = "hostname"
	LogCtx_Ip        = "ip"
	LogCtx_K8s_Node  = "k8s_node"
	LogCtx_K8s_Pod   = "k8s_pod"
)

var logPrefix map[LogLevel]string

type LogMessage struct {
	Level   LogLevel
	Message string
}

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
