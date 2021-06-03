package log

import (
	"fmt"
	"io/ioutil"
	"log"
	syslog "log"
	"net/http"
	"strings"
	"sync"
	"time"

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

func newLogzioShipper(conf config.Config, secretsManager secrets.SecretsManager) LogShipper {

	logzioUrl := conf.Get("log.logzio.url", config.AsStringPtr(LOGZIO_URL))
	batchSize := conf.GetAsInt("log.logzio.batchsize", config.AsIntPtr(LOGZIO_BATCH_SIZE))
	shipmentStackSize := conf.GetAsInt("log.logzio.shipmentstacksize", config.AsIntPtr(SHIPMENT_STACK_SIZE))
	messageStackSize := conf.GetAsInt("log.logzio.messagestacksize", config.AsIntPtr(MESSAGE_STACK_SIZE))
	shipmentTimeout := conf.GetAsDuration("log.logzio.shipmenttimeout", config.AsDurationPtr(SHIPMENT_WAIT_TIMEOUT))
	messageReadTimeout := conf.GetAsDuration("log.logzio.messagereadtimeout", config.AsDurationPtr(MESSAGE_READ_TIMEOUT))

	shipper := &LogzioShipper{
		logzioUrl:             *logzioUrl,
		batchSize:             *batchSize,
		shipmentStack:         make(chan bool, *shipmentStackSize),
		messageStack:          make(chan string, *messageStackSize),
		obtainShipmentTimeout: *shipmentTimeout,
		messageReadTimeout:    *messageReadTimeout,
		httpClient:            &http.Client{},
		secretsManager:        secretsManager,
	}
	shipper.initShipmentStack()
	return shipper
}

// Send will add passed log message to an internal queue and starts shipment if
// number of buffered messages exceeds defined batch size.
func (shipper *LogzioShipper) send(message string) {

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
func (shipper *LogzioShipper) initShipmentStack() {
	for len(shipper.shipmentStack) < cap(shipper.shipmentStack) {
		shipper.shipmentStack <- true
	}
}

// Flush will deliver all messages from internal channel to Logz.io.
func (shipper *LogzioShipper) flush() {

	wg := &sync.WaitGroup{}
	for len(shipper.messageStack) > 0 {
		wg.Add(1)
		shipper.shipBatch(wg)
		wg.Wait()
	}
}

// ObtainShipment will try to get a slot for shipment from shipment stack.
// It will return with false if obtainShipmentTimeout exceeds.
func (shipper *LogzioShipper) obtainShipment() bool {

	timeout := time.NewTimer(shipper.obtainShipmentTimeout)
	select {
	case <-shipper.shipmentStack:
		return true
	case <-timeout.C:
		return false
	}
}

// ReleaseShipment will return a used slot to the shipment stack.
func (shipper *LogzioShipper) releaseShipment() {

	if len(shipper.shipmentStack) < cap(shipper.shipmentStack) {
		shipper.shipmentStack <- true
	}
}

// ShipBatch will read number of messages defined by batch size from internal channel
// and start shipment for all of them.
func (shipper *LogzioShipper) shipBatch(wg *sync.WaitGroup) {

	messages := shipper.readMessages()
	shipper.shipMessages(wg, messages)
}

// ReadMessages will try to read number of messages defined by batch size from internal buffer.
// If it exceeds messages read timeout it will return messages it reads up to this point in time.
func (shipper *LogzioShipper) readMessages() []string {

	var messages []string
	timeout := time.NewTimer(shipper.messageReadTimeout)
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

// ShipMessages will send passed log messages to defines Logz.io endpoint.
func (shipper *LogzioShipper) shipMessages(wg *sync.WaitGroup, messages []string) {

	defer wg.Done()

	messageBatch := strings.Join(messages, "\n")
	req, _ := http.NewRequest("POST", shipper.logzIoUrl(), strings.NewReader(messageBatch))
	req.Header.Set("Content-Type", "application/json")
	shipper.sendRequest(req)
}

// SendRequest will execute passed request and validate it's response.
func (shipper *LogzioShipper) sendRequest(request *http.Request) {

	resp, err := shipper.httpClient.Do(request)
	if err != nil {
		log.Println(err)
	} else if resp.StatusCode >= 400 {
		var responseBody string
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			if bodyBytes, err := ioutil.ReadAll(resp.Body); err == nil {
				responseBody = string(bodyBytes)
			}
		}
		log.Println(fmt.Errorf("Logz.io response, %d: %s", resp.StatusCode, responseBody))
	}
}

// logError writes given error to STDERR.
func (shipper *LogzioShipper) logError(err error) {
	syslog.Println(err)
}

// logzIoUrl generates the Logz.io endpoint for importing logs.
func (shipper *LogzioShipper) logzIoUrl() string {
	token, err := shipper.secretsManager.Obtain(LOGZIO_TOKEN_KEY)
	if err != nil {
		shipper.logError(err)
		return fmt.Sprintf("%s?token=%s&type=go-logs", shipper.logzioUrl, "<LogzioTokenNotFound>")
	}
	return fmt.Sprintf("%s?token=%s&type=go-logs", shipper.logzioUrl, *token)
}
