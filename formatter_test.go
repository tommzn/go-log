package log

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FomatterTestSuite struct {
	suite.Suite
}

func TestFomatterTestSuite(t *testing.T) {
	suite.Run(t, new(FomatterTestSuite))
}

func (suite *FomatterTestSuite) TestDefaultFormatter() {

	formatter := newDefaultFormatter()
	logLevel := Error
	context := suite.contextForTest()
	message := "Test Message"
	expextedLogMessage := "Error: Test Message, Context: namespace:FomatterTestSuite,timestamp:2021-05-30T12:08:47+02:00"

	suite.Equal(expextedLogMessage, formatter.format(logLevel, context, message))
}

func (suite *FomatterTestSuite) TestLogzioJsonFormatter() {

	formatter := newLogzioJsonFormatter()
	logLevel := Error
	context := suite.contextForTest()
	message := "Test Message"

	logMessage := formatter.format(logLevel, context, message)
	suite.True(strings.Contains(logMessage, "Error"))
	suite.True(strings.Contains(logMessage, "@timestamp"))
	suite.True(strings.Contains(logMessage, message))
}

func (suite *FomatterTestSuite) contextForTest() LogContext {
	logContext := make(map[string]string)
	logContext[LogCtxTimestamp] = "2021-05-30T12:08:47+02:00"
	logContext[LogCtxNamespace] = "FomatterTestSuite"
	return LogContext{values: logContext}
}
