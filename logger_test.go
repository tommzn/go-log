package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	secrets "github.com/tommzn/go-secrets"
)

type LoggerTestSuite struct {
	suite.Suite
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (suite *LoggerTestSuite) TestCreateLoggerWithDefaults() {

	logger := NewLogger(Debug, nil, nil)
	suite.IsType(&DefaultFormatter{}, logger.(*LogHandler).formatter)
	suite.IsType(&StdoutShipper{}, logger.(*LogHandler).shipper)
}

func (suite *LoggerTestSuite) TestCreateLoggerFromConfig() {

	conf1 := loadConfigFromFile("config/logzio.yml")
	logger1 := NewLoggerFromConfig(conf1, secrets.NewSecretsManager())
	suite.IsType(&LogzioShipper{}, logger1.(*LogHandler).shipper)
	suite.Equal(Debug, logger1.(*LogHandler).logLevel)
	logger1.Error("Test Log")

	conf2 := loadConfigFromFile("config/stdout.yml")
	logger2 := NewLoggerFromConfig(conf2, nil)
	suite.IsType(&StdoutShipper{}, logger2.(*LogHandler).shipper)
	suite.Equal(Info, logger2.(*LogHandler).logLevel)
	logger2.Error("Test Log")

	conf3 := loadConfigFromFile("config/empty.yml")
	logger3 := NewLoggerFromConfig(conf3, nil)
	suite.IsType(&StdoutShipper{}, logger3.(*LogHandler).shipper)
	suite.Equal(Error, logger3.(*LogHandler).logLevel)
	logger3.Error("Test Log")
}

func (suite *LoggerTestSuite) TestLogging() {

	shipper := newTestShipper().(*testShipper)
	logger := NewLogger(Debug, nil, shipper)

	expectedNumberOfLogMessages := 1
	logger.Status("This ", "is ", "a ", "test.")
	suite.assertLogMessage(expectedNumberOfLogMessages, "Status: This is a test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Error("This ", "is ", "a ", "test.")
	suite.assertLogMessage(expectedNumberOfLogMessages, "Error: This is a test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Info("This ", "is ", "a ", "test.")
	suite.assertLogMessage(expectedNumberOfLogMessages, "Info: This is a test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Debug("This ", "is ", "a ", "test.")
	suite.assertLogMessage(expectedNumberOfLogMessages, "Debug: This is a test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Statusf("This is the %dst test.", 1)
	suite.assertLogMessage(expectedNumberOfLogMessages, "Status: This is the 1st test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Errorf("This is the %dst test.", 1)
	suite.assertLogMessage(expectedNumberOfLogMessages, "Error: This is the 1st test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Infof("This is the %dst test.", 1)
	suite.assertLogMessage(expectedNumberOfLogMessages, "Info: This is the 1st test., Context: ", shipper)

	expectedNumberOfLogMessages++
	logger.Debugf("This is the %dst test.", 1)
	suite.assertLogMessage(expectedNumberOfLogMessages, "Debug: This is the 1st test., Context: ", shipper)
}

func (suite *LoggerTestSuite) TestLoggingWithContext() {

	shipper := newTestShipper().(*testShipper)
	contextValues := make(map[string]string)
	contextValues["Key1"] = "Value1"
	contextValues["Key2"] = "Value2"
	ctx := LogContextWithValues(context.Background(), contextValues)
	logger := NewLogger(Debug, nil, shipper)
	logger.WithContext(ctx)

	expectedNumberOfLogMessages := 1
	logger.Error("This ", "is ", "a ", "test.")
	suite.assertLogMessage(expectedNumberOfLogMessages, "Error: This is a test., Context: Key1:Value1,Key2:Value2", shipper)

	// FLuah will have no effect, but should not throw any errors.
	logger.Flush()
}

func (suite *LoggerTestSuite) assertLogMessage(expectedNumberOfLogMessages int, expectedMessage string, in *testShipper) {
	suite.Len(in.messages, expectedNumberOfLogMessages)
	suite.Equal(expectedMessage, in.messages[expectedNumberOfLogMessages-1])
}

/**
func (suite *LoggerTestSuite) SetupTest() {
	suite.Nil(config.UseConfigFile("logger_testconfig"))
}

func (suite *LoggerTestSuite) skipGitLabCI() {
	if _, ok := os.LookupEnv("GITLAB_CI"); ok {
		suite.T().Skip("Skipping testing in GitLab CI environment")
	}
}

func (suite *LoggerTestSuite) TestLogging() {

	logger := NewLogger(Error, "test")
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)
}

func (suite *LoggerTestSuite) TestLogWithContext() {

	logger := NewLogger(Error, "test")
	logger.AppendContext(LogCtx_MessageId, "6484fd4e-7429-4f71-b1a4-346e52846092")
	logger.AppendContext(LogCtx_FactoryId, "1aa9f2e0-db9a-427d-bad7-88ea41684fdb")
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)

	logger.ClearContext()
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)
}

func (suite *LoggerTestSuite) TestS3Logger() {

	suite.skipGitLabCI()

	config := getLoggerConfig()
	logger := NewS3Logger(Error, "test", config)
	suite.Equal(5, logger.(*S3Logger).batchSize)
	logger.AppendContext(LogCtx_MessageId, "6484fd4e-7429-4f71-b1a4-346e52846092")
	logger.AppendContext(LogCtx_FactoryId, "1aa9f2e0-db9a-427d-bad7-88ea41684fdb")
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)

	logger.ClearContext()
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)

	logger.Flush()
	time.Sleep(1 * time.Second)

}

func (suite *LoggerTestSuite) TestLogzIoLogger() {

	config := getLoggerConfig()
	defaultContext := make(map[string]string)
	logger := NewLogzIoLogger(config, secretsManagerForTest())
	logger.ApplyDefaultContext(defaultContext)
	suite.Equal(5, logger.(*LogzIoLogger).batchSize)
	logger.AppendContext(LogCtx_MessageId, "6484fd4e-7429-4f71-b1a4-346e52846092")
	logger.AppendContext(LogCtx_FactoryId, "1aa9f2e0-db9a-427d-bad7-88ea41684fdb")
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)

	logger.ClearContext()
	logger.Errorf("D: %d, S: %s, V: %s, D; %d", 1, "xxx", "ref", 44)

	logger.FlushWithWait()
}

func (suite *LoggerTestSuite) TestMockedLogger() {

	logger := NewMockedLogger()

	logger.Status("Hello World")
	logger.Statusf("Hello %s", "World")

	logger.Error("Hello World")
	logger.Errorf("Hello %s", "World")

	logger.Info("Hello World")
	logger.Infof("Hello %s", "World")

	logger.Debug("Hello World")
	logger.Debugf("Hello %s", "World")

	logs := logger.GetLogs()
	suite.assertLogCount(logs, Status, 2)
	suite.assertLogCount(logs, Error, 2)
	suite.assertLogCount(logs, Info, 2)
	suite.assertLogCount(logs, Debug, 2)
}

func (suite *LoggerTestSuite) TestMockedLoggerWithChannel() {

	logger := NewMockedLoggerWithChannel(2)

	logger.Error("Hello World")
	logger.Info("Hello World")
	logger.Debug("Hello World")

	logs := logger.GetLogs()
	suite.assertLogCount(logs, Error, 1)
	suite.assertLogCount(logs, Info, 1)

	// Log channel is to small for last message
	suite.assertLogCount(logs, Debug, 0)
}

func (suite *LoggerTestSuite) TestMockedLoggerWithAsyncChannel() {

	logger := NewMockedLoggerWithChannel(2)

	go produceLogs(logger)
	time.Sleep(1 * time.Second)

	logs := logger.GetLogs()
	suite.assertLogCount(logs, Error, 1)
	suite.assertLogCount(logs, Info, 1)
}

func (suite *LoggerTestSuite) assertLogCount(logs []LogMessage, loglevel LogLevel, expectedCount int) {

	logCount := 0
	for _, log := range logs {
		if log.Level == loglevel {
			logCount++
		}
	}
	suite.Equal(expectedCount, logCount)
}

func (suite *LoggerTestSuite) TestDefaultContextForNodes() {

	context := DefaultContextForNodes()
	suite.Len(context, 2)
}

func (suite *LoggerTestSuite) TestDefaultContextForK8s() {

	os.Setenv("TSL_K8S_NODE_NAME", "xxx")
	os.Setenv("TSL_K8S_POD_NAME", "yyy")
	context := DefaultContextForK8s()
	suite.Len(context, 2)
}

func getLoggerConfig() config.Config {
	configLoader := config.NewConfigLoader()
	return configLoader.Load()
}

func secretsManagerForTest() secrets.SecretsManager {
	return secrets.NewSecretsManager()
}

func produceLogs(logger LoggerMock) {
	logger.Error("Error 1")
	logger.Info("Info 1")
	logger.Info("Info 2")
}
*/
