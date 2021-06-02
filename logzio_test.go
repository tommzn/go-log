package log

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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

	shipper := newLogzioShipper(conf, suite.secretsManagerForTest())
	suite.IsType(&LogzioShipper{}, shipper)

	logzioShipper, _ := shipper.(*LogzioShipper)
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

	shipper.send(logMessage)
	suite.Len(shipper.messageStack, 1)
}

func (suite *LogzioShipperTestSuite) TestLogWithShipment() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).response = &http.Response{StatusCode: 200}
	shipper.send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Len(shipper.httpClient.(*testClient).requests, 1)
}

func (suite *LogzioShipperTestSuite) TestObtainShipmentTimeout() {

	shipper := suite.shipperForTest()
	shipper.shipmentStack = make(chan bool, 1)
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).response = &http.Response{StatusCode: 200}
	shipper.send(logMessage)

	time.Sleep(2 * time.Second)
	suite.Len(shipper.messageStack, 4)
	suite.Len(shipper.httpClient.(*testClient).requests, 0)
}

func (suite *LogzioShipperTestSuite) TestReadMessageTimeout() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	// Increase timeout to obtain shipment, to have enough time to manipulate message channel
	shipper.obtainShipmentTimeout = 5 * time.Second
	<-shipper.shipmentStack
	shipper.httpClient.(*testClient).response = &http.Response{StatusCode: 200}
	go shipper.send(logMessage)
	time.Sleep(1 * time.Second)

	// read two messages from internal channel to have less messages than batch size
	<-shipper.messageStack
	<-shipper.messageStack

	// Add shipment slot to continue message shipment
	shipper.shipmentStack <- true

	time.Sleep(2 * time.Second)
	suite.Len(shipper.messageStack, 0)
	suite.Len(shipper.httpClient.(*testClient).requests, 1)
}

func (suite *LogzioShipperTestSuite) TestShipmentWithFailedRequest() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).response = &http.Response{StatusCode: 400, Body: ioutil.NopCloser(strings.NewReader("Shipment Error!"))}
	shipper.send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Len(shipper.httpClient.(*testClient).requests, 1)
}

func (suite *LogzioShipperTestSuite) TestShipmentWithRequestError() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).err = errors.New("Shipment Error!")
	shipper.send(logMessage)

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 1)
	suite.Len(shipper.httpClient.(*testClient).requests, 1)
}

func (suite *LogzioShipperTestSuite) TestFlusgMessages() {

	shipper := suite.shipperForTest()
	logMessage := "Debug: Log Message"
	for i := 1; i <= shipper.batchSize; i++ {
		shipper.send(logMessage)
	}
	suite.Len(shipper.messageStack, shipper.batchSize)

	shipper.httpClient.(*testClient).response = &http.Response{StatusCode: 200}
	shipper.flush()

	time.Sleep(1 * time.Second)
	suite.Len(shipper.messageStack, 0)
	suite.Len(shipper.httpClient.(*testClient).requests, 1)
}

func (suite *LogzioShipperTestSuite) TestGetLogzIoUrl() {

	shipper := suite.shipperForTest()
	suite.Equal("https://localhost:8071/?token=<LogzioToken>&type=go-logs", shipper.logzIoUrl())

	shipper.secretsManager = secrets.NewStaticSecretsManager(make(map[string]string))
	suite.Equal("https://localhost:8071/?token=<LogzioTokenNotFound>&type=go-logs", shipper.logzIoUrl())
}

func (suite *LogzioShipperTestSuite) TestLogzioIntegration() {

	if _, ok := os.LookupEnv("LOGZIO_TOKEN"); !ok {
		suite.T().Skip("Skip Logz.io integration test without token.")
	}
	conf := loadConfigFromFile("config/testconfig.yml")
	shipper := newLogzioShipper(conf, secrets.NewSecretsManager())

	formatter := newLogzioJsonFormatter()
	logMessage := formatter.format(Debug, newLogContext(make(map[string]string)), "go-log test")
	for i := 1; i <= shipper.(*LogzioShipper).batchSize; i++ {
		shipper.send(logMessage)
	}
	shipper.flush()
}

func (suite *LogzioShipperTestSuite) shipperForTest() *LogzioShipper {
	shipper := &LogzioShipper{
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
