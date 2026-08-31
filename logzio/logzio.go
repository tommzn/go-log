// Package logzio provides a LogShipper/LogFormatter pair that ship log
// messages to Logz.io. Import this package only when Logz.io support is
// needed; it brings in net/http (and everything that pulls in - crypto/tls,
// IDNA, Unicode normalization) as a dependency. Everything else in go-log
// stays dependency-light without it.
//
// Blank-import this package to activate it for config-driven dispatch via
// log.NewLoggerFromConfig ("log.shipper: logzio" in config):
//
//	import _ "github.com/tommzn/go-log/v2/logzio"
//
// Or construct a shipper/formatter explicitly, without going through config
// at all:
//
//	logger := log.NewLogger(log.Debug, logzio.NewFormatter(), logzio.NewShipper(conf, secretsManager))
package logzio

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	golog "github.com/tommzn/go-log/v2"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

// LOGZIO_URL define the endpoint all logs will be shipped to.
// Can be set by config: log.logzio.url
const LOGZIO_URL = "https://listener.logz.io:8071/"

// LOGZIO_TOKEN_KEY defines the key which will be used to obtain
// the Logz.io token from secrets mananger.
const LOGZIO_TOKEN_KEY = "LOGZIO_TOKEN"

// LOGZIO_TIMESTAMP_FORMAT is Logz.io timestamp format which will be used
// for @timestamp value in a log record.
const LOGZIO_TIMESTAMP_FORMAT = "2006-01-02T15:04:05.999Z"

// LOGZIO_BATCH_SIZE is the default batch size the Logz.io shipper will use.
// Can be set by config: log.logzio.batchsize
const LOGZIO_BATCH_SIZE = 10

// MESSAGE_STACK_SIZE defines the capacity of internal message buffer.
// Can be set by config: log.logzio.messagestacksize
const MESSAGE_STACK_SIZE = 500

// SHIPMENT_STACK_SIZE defines the number of worker which can ship logs in parallel to Logz.io.
// Can be set by config: log.logzio.shipmentstacksize
const SHIPMENT_STACK_SIZE = 2

// SHIPMENT_WAIT_TIMEOUT defines the time the shipper will wait to get a slot from shipment stack.
// Can be set by config: log.logzio.shipmenttimeout
const SHIPMENT_WAIT_TIMEOUT = 1 * time.Second

// MESSAGE_READ_TIMEOUT defines the time a shipper will wait for new messages during reading from message stack.
// Can be set by config: log.logzio.messagereadtimeout
const MESSAGE_READ_TIMEOUT = 50 * time.Millisecond

// HTTP_CLIENT_TIMEOUT bounds a single Logz.io request (connection, TLS
// handshake, request write and response read combined). Without this, the
// default http.Client has no deadline at all: a single request that hangs
// (a network hiccup, a stalled TLS handshake, Logz.io itself being slow)
// blocks its goroutine forever, permanently holding one of shipmentStack's
// slots. Two hangs on two shipments is enough to exhaust shipmentstacksize
// (2 by default) and jam the shipper for good - every future Send/Flush
// call then times out in obtainShipment and silently drops, with no error
// ever logged, since obtainShipment's failure path isn't itself an error.
// Can be set by config: log.logzio.timeout
const HTTP_CLIENT_TIMEOUT = 10 * time.Second

// init registers this shipper with go-log's config-driven dispatch, so
// log.NewLoggerFromConfig picks it up for "log.shipper: logzio" once this
// package has been blank-imported.
func init() {
	golog.RegisterShipper("logzio", func(conf config.Config, secretsManager secrets.SecretsManager) (golog.LogFormatter, golog.LogShipper) {
		return NewFormatter(), NewShipper(conf, secretsManager)
	})
}

// httpClient is an interface for a HTTP client.
type httpClient interface {

	// Do will send a http request.
	Do(req *http.Request) (*http.Response, error)
}

// Formatter converts passed values to a JSON record suitable for an import at Logz.io.
type Formatter struct {
}

// NewFormatter returns a new Formatter.
func NewFormatter() golog.LogFormatter {
	return &Formatter{}
}

// Format composes passed log level, context and message in a map and marshals it to JSON.
// Builds its own copy of the context values (via LogContext.Values()) rather than
// mutating anything owned by logContext - that value may be the same instance a
// *log.LogHandler keeps for every message it logs, so mutating it here would race
// with concurrent log calls on the same logger.
func (formatter *Formatter) Format(logLevel golog.LogLevel, logContext golog.LogContext, message string) string {

	ctxValues := logContext.Values()
	if ctxValues == nil {
		ctxValues = make(map[string]string, 3)
	}
	ctxValues[golog.LogCtxLogLevel] = logLevel.String()
	ctxValues["@timestamp"] = time.Now().UTC().Format(LOGZIO_TIMESTAMP_FORMAT)
	ctxValues[golog.LogCtxMessage] = message

	logContent, err := json.Marshal(ctxValues)
	if err != nil {
		return fmt.Sprintf(`{"message":%q,"loglevel":%q,"error":"json marshal failed"}`, message, logLevel.String())
	}
	return string(logContent)
}

// Shipper delivers log messages to Logz.io.
type Shipper struct {

	// logzioUrl is the endpoint all logs will be shipped to.
	logzioUrl string

	// batchSize defines the number of logs shipped together in a batch.
	batchSize int

	// shipmentStack is a worker queue to restrict parallel shipment.
	shipmentStack chan bool

	// messageStack is a channel to buffer log messages.
	messageStack chan string

	// obtainShipmentTimeout defines the time the shipper will wait to get
	// a slot from shipmentStack.
	obtainShipmentTimeout time.Duration

	// messageReadTimeout defines the time a shipper will wait for new messages
	// during reading from messageStack.
	messageReadTimeout time.Duration

	// httpClient is used to send POST request to ship log messages.
	httpClient httpClient

	// secretsManager is used to obtain Logz.io token for shipment requests.
	secretsManager secrets.SecretsManager
}

// NewShipper returns a new Shipper, configured from conf. If secretsManager
// is nil, a default secrets.NewSecretsManager() (environment-variable-backed)
// is used instead - shipping a message calls secretsManager.Obtain to get the
// Logz.io token, which would otherwise panic on first shipment rather than
// failing fast at construction time.
func NewShipper(conf config.Config, secretsManager secrets.SecretsManager) golog.LogShipper {

	if secretsManager == nil {
		secretsManager = secrets.NewSecretsManager()
	}

	logzioUrl := conf.Get("log.logzio.url", config.AsStringPtr(LOGZIO_URL))
	batchSize := conf.GetAsInt("log.logzio.batchsize", config.AsIntPtr(LOGZIO_BATCH_SIZE))
	shipmentStackSize := conf.GetAsInt("log.logzio.shipmentstacksize", config.AsIntPtr(SHIPMENT_STACK_SIZE))
	messageStackSize := conf.GetAsInt("log.logzio.messagestacksize", config.AsIntPtr(MESSAGE_STACK_SIZE))
	shipmentTimeout := conf.GetAsDuration("log.logzio.shipmenttimeout", config.AsDurationPtr(SHIPMENT_WAIT_TIMEOUT))
	messageReadTimeout := conf.GetAsDuration("log.logzio.messagereadtimeout", config.AsDurationPtr(MESSAGE_READ_TIMEOUT))
	httpTimeout := conf.GetAsDuration("log.logzio.timeout", config.AsDurationPtr(HTTP_CLIENT_TIMEOUT))

	shipper := &Shipper{
		logzioUrl:             *logzioUrl,
		batchSize:             *batchSize,
		shipmentStack:         make(chan bool, *shipmentStackSize),
		messageStack:          make(chan string, *messageStackSize),
		obtainShipmentTimeout: *shipmentTimeout,
		messageReadTimeout:    *messageReadTimeout,
		httpClient:            &http.Client{Timeout: *httpTimeout},
		secretsManager:        secretsManager,
	}
	shipper.initShipmentStack()
	return shipper
}

// Send will add passed log message to an internal queue and starts shipment if
// number of buffered messages exceeds defined batch size.
func (shipper *Shipper) Send(message string) {

	shipper.messageStack <- message

	if len(shipper.messageStack) <= shipper.batchSize {
		return
	}

	if !shipper.obtainShipment() {
		return
	}

	go func() {

		wg := &sync.WaitGroup{}
		wg.Add(1)
		shipper.shipBatch(wg)

		wg.Wait()
		shipper.releaseShipment()
	}()
}

// initShipmentStack fills the shipment stack with all slots.
func (shipper *Shipper) initShipmentStack() {
	for len(shipper.shipmentStack) < cap(shipper.shipmentStack) {
		shipper.shipmentStack <- true
	}
}

// Flush will deliver all messages from internal channel to Logz.io.
func (shipper *Shipper) Flush() {

	wg := &sync.WaitGroup{}
	for len(shipper.messageStack) > 0 {
		wg.Add(1)
		shipper.shipBatch(wg)
		wg.Wait()
	}
}

// obtainShipment will try to get a slot for shipment from shipment stack.
// It will return with false if obtainShipmentTimeout exceeds.
func (shipper *Shipper) obtainShipment() bool {

	timeout := time.NewTimer(shipper.obtainShipmentTimeout)
	defer timeout.Stop()
	select {
	case <-shipper.shipmentStack:
		return true
	case <-timeout.C:
		return false
	}
}

// releaseShipment will return a used slot to the shipment stack.
func (shipper *Shipper) releaseShipment() {

	if len(shipper.shipmentStack) < cap(shipper.shipmentStack) {
		shipper.shipmentStack <- true
	}
}

// shipBatch will read number of messages defined by batch size from internal channel
// and start shipment for all of them.
func (shipper *Shipper) shipBatch(wg *sync.WaitGroup) {

	messages := shipper.readMessages()
	shipper.shipMessages(wg, messages)
}

// readMessages will try to read number of messages defined by batch size from internal buffer.
// If it exceeds messages read timeout it will return messages it reads up to this point in time.
func (shipper *Shipper) readMessages() []string {

	var messages []string
	timeout := time.NewTimer(shipper.messageReadTimeout)
	defer timeout.Stop()
	for len(messages) < shipper.batchSize {
		select {
		case message := <-shipper.messageStack:
			messages = append(messages, message)
		case <-timeout.C:
			return messages
		}
	}
	return messages
}

// shipMessages will send passed log messages to defines Logz.io endpoint.
func (shipper *Shipper) shipMessages(wg *sync.WaitGroup, messages []string) {

	defer wg.Done()

	messageBatch := strings.Join(messages, "\n")
	req, err := http.NewRequest(http.MethodPost, shipper.logzIoUrl(), strings.NewReader(messageBatch))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	shipper.sendRequest(req)
}

// sendRequest will execute passed request and validate it's response.
func (shipper *Shipper) sendRequest(request *http.Request) {

	resp, err := shipper.httpClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	if resp == nil {
		log.Println("logz.io: received nil response")
		return
	}
	if resp.StatusCode >= 400 {
		var responseBody string
		if resp.Body != nil {
			defer resp.Body.Close()
			if bodyBytes, err := io.ReadAll(resp.Body); err == nil {
				responseBody = string(bodyBytes)
			}
		}
		log.Printf("Logz.io response, %d: %s\n", resp.StatusCode, responseBody)
	}
}

// logError writes given error to STDERR.
func (shipper *Shipper) logError(err error) {
	log.Println(err)
}

// logzIoUrl generates the Logz.io endpoint for importing logs.
func (shipper *Shipper) logzIoUrl() string {
	token, err := shipper.secretsManager.Obtain(LOGZIO_TOKEN_KEY)
	if err != nil {
		shipper.logError(err)
		return fmt.Sprintf("%s?token=%s&type=go-logs", shipper.logzioUrl, "<LogzioTokenNotFound>")
	}
	return fmt.Sprintf("%s?token=%s&type=go-logs", shipper.logzioUrl, *token)
}
