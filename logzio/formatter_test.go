package logzio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	golog "github.com/tommzn/go-log/v2"
)

type FormatterTestSuite struct {
	suite.Suite
}

func TestFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(FormatterTestSuite))
}

func (suite *FormatterTestSuite) TestFormat() {

	formatter := NewFormatter()
	logLevel := golog.Error
	context := suite.contextForTest()
	message := "Test Message"

	logMessage := formatter.Format(logLevel, context, message)
	suite.True(strings.Contains(logMessage, "Error"))
	suite.True(strings.Contains(logMessage, "@timestamp"))
	suite.True(strings.Contains(logMessage, message))
}

func (suite *FormatterTestSuite) TestFormatDoesNotMutatePassedContext() {

	formatter := NewFormatter()
	context := golog.NewLogContext(map[string]string{"namespace": "test"})

	_ = formatter.Format(golog.Error, context, "message")

	// The context passed in must still only carry what was originally set -
	// Format must not have leaked loglevel/@timestamp/message back into it.
	suite.Equal(map[string]string{"namespace": "test"}, context.Values())
}

func (suite *FormatterTestSuite) contextForTest() golog.LogContext {
	return golog.NewLogContext(map[string]string{
		golog.LogCtxTimestamp: "2021-05-30T12:08:47+02:00",
		golog.LogCtxNamespace: "FormatterTestSuite",
	})
}
