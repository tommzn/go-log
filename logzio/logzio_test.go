package logzio

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	golog "github.com/tommzn/go-log/v2"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

type LogzioShipperTestSuite struct {
	suite.Suite
}

func TestLogzioShipperTestSuite(t *testing.T) {
	suite.Run(t, new(LogzioShipperTestSuite))
}

func (suite *LogzioShipperTestSuite) TestCreateShipperFromConfig() {

	conf := loadConfigFromFile("config/logzio.yml")

	shipper := NewShipper(conf, suite.secretsManagerForTest())
	suite.IsType(&Shipper{}, shipper)

	logzioShipper, _ := shipper.(*Shipper)
	suite.Equal("https://example.com/", logzioShipper.logzioUrl)
	suite.Equal(12, logzioShipper.batchSize)
	suite.True(cap(logzioShipper.shipmentStack) == 3)
	suite.True(cap(logzioShipper.messageStack) == 123)
	suite.Equal(7*time.Second, logzioShipper.obtainShipmentTimeout)
	suite.Equal(14*time.Second, logzioShipper.messageReadTimeout)
}

func (suite *LogzioShipperTestSuite) TestLogWithoutShipment() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"

	shipper.Send(logMessage)
	suite.Len(shipper.messageStack, 1)
}

func (suite *LogzioShipperTestSuite) TestLogWithShipment() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).setResponse(&http.Response{StatusCode: 200}, nil)
	shipper.Send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Equal(1, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestObtainShipmentTimeout() {

	shipper := suite.shipperForTest()
	shipper.shipmentStack = make(chan bool, 1)
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).setResponse(&http.Response{StatusCode: 200}, nil)
	shipper.Send(logMessage)

	time.Sleep(2 * time.Second)
	suite.Len(shipper.messageStack, 4)
	suite.Equal(0, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestReadMessageTimeout() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	// Increase timeout to obtain shipment, to have enough time to manipulate message channel
	shipper.obtainShipmentTimeout = 5 * time.Second
	<-shipper.shipmentStack
	shipper.httpClient.(*testClient).setResponse(&http.Response{StatusCode: 200}, nil)
	go shipper.Send(logMessage)
	time.Sleep(1 * time.Second)

	// read two messages from internal channel to have less messages than batch size
	<-shipper.messageStack
	<-shipper.messageStack

	// Add shipment slot to continue message shipment
	shipper.shipmentStack <- true

	time.Sleep(2 * time.Second)
	suite.Len(shipper.messageStack, 0)
	suite.Equal(1, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestShipmentWithFailedRequest() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).setResponse(&http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("Shipment Error!"))}, nil)
	shipper.Send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Equal(1, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestShipmentWithRequestError() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).setResponse(nil, errors.New("Shipment Error!"))
	shipper.Send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Equal(1, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestFlusgMessages() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.Send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).setResponse(&http.Response{StatusCode: 200}, nil)
	shipper.Flush()

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 0)
	suite.Equal(1, shipper.httpClient.(*testClient).requestCount())
}

func (suite *LogzioShipperTestSuite) TestGetLogzIoUrl() {

	shipper := suite.shipperForTest()
	suite.Equal("https://localhost:8071/?token=<LogzioToken>&type=go-logs", shipper.logzIoUrl())

	shipper.secretsManager = secrets.NewStaticSecretsManager(make(map[string]string))
	suite.Equal("https://localhost:8071/?token=<LogzioTokenNotFound>&type=go-logs", shipper.logzIoUrl())
}

func (suite *LogzioShipperTestSuite) TestLogError() {

	shipper := suite.shipperForTest()
	// logError should not panic; it writes to stderr via log.Println
	suite.NotPanics(func() {
		shipper.logError(errors.New("test error"))
	})
}

func (suite *LogzioShipperTestSuite) TestLogzioIntegration() {

	if _, ok := os.LookupEnv("LOGZIO_TOKEN"); !ok {
		suite.T().Skip("Skip Logz.io integration test without token.")
	}
	conf := loadConfigFromFile("config/testconfig.yml")
	shipper := NewShipper(conf, secrets.NewSecretsManager())

	formatter := NewFormatter()
	logMessage := formatter.Format(golog.Debug, golog.NewLogContext(make(map[string]string)), "go-log test")
	for i := 1; i <= shipper.(*Shipper).batchSize; i++ {
		shipper.Send(logMessage)
	}
	shipper.Flush()
}

// TestRegisteredWithGoLog confirms this package's init() actually registers it
// with go-log's config-driven dispatch, end to end - the actual scenario
// RegisterShipper exists for.
func (suite *LogzioShipperTestSuite) TestRegisteredWithGoLog() {

	conf := loadConfigFromFile("config/logzio.yml")
	logger := golog.NewLoggerFromConfig(conf, suite.secretsManagerForTest())
	suite.NotNil(logger)
	// No exported way to inspect the concrete shipper type from outside
	// go-log's own package - constructing without panicking/erroring and
	// logging without blocking is the externally-observable contract here.
	suite.NotPanics(func() {
		logger.Debug("registered shipper smoke test")
	})
}

func (suite *LogzioShipperTestSuite) shipperForTest() *Shipper {
	shipper := &Shipper{
		logzioUrl:             "https://localhost:8071/",
		batchSize:             3,
		shipmentStack:         make(chan bool, 1),
		messageStack:          make(chan string, 10),
		obtainShipmentTimeout: 500 * time.Millisecond,
		messageReadTimeout:    500 * time.Millisecond,
		httpClient:            newHttpTestClient(nil, nil),
		secretsManager:        suite.secretsManagerForTest(),
	}
	shipper.initShipmentStack()
	return shipper
}

func (suite *LogzioShipperTestSuite) secretsManagerForTest() secrets.SecretsManager {
	secretsMap := make(map[string]string)
	secretsMap[LOGZIO_TOKEN_KEY] = "<LogzioToken>"
	return secrets.NewStaticSecretsManager(secretsMap)
}

// testClient is a HTTP client mock for testing. Guarded by a mutex since Do
// runs on a shipper-internal goroutine while tests read/write requests,
// response and err directly from the test goroutine.
type testClient struct {
	mu       sync.Mutex
	requests []*http.Request
	response *http.Response
	err      error
}

func newHttpTestClient(response *http.Response, err error) httpClient {
	return &testClient{response: response, err: err, requests: []*http.Request{}}
}

func (client *testClient) Do(req *http.Request) (*http.Response, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.requests = append(client.requests, req)
	return client.response, client.err
}

// setResponse sets the response/error Do will return for subsequent calls.
func (client *testClient) setResponse(response *http.Response, err error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.response = response
	client.err = err
}

// requestCount returns the number of requests Do has received so far.
func (client *testClient) requestCount() int {
	client.mu.Lock()
	defer client.mu.Unlock()
	return len(client.requests)
}

func loadConfigFromFile(fileName string) config.Config {
	configSource := config.NewFileConfigSource(&fileName)
	conf, _ := configSource.Load()
	return conf
}
