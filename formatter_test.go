package log

import (
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
	suite.Equal(expextedLogMessage, formatter.Format(logLevel, context, message))
}

func (suite *FomatterTestSuite) contextForTest() LogContext {
	logContext := make(map[string]string)
	logContext[LogCtxTimestamp] = "2021-05-30T12:08:47+02:00"
	logContext[LogCtxNamespace] = "FomatterTestSuite"
	return LogContext{values: logContext}
}
