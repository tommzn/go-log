package log

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (suite *UtilsTestSuite) TestNameSpaceContext() {

	logger := NewLogger(Debug, nil, nil)
	logger = WithNameSpace(logger, "ns.test")
	logHandler, ok := logger.(*LogHandler)
	suite.True(ok)
	suite.Len(logHandler.context.values, 1)
}

func (suite *UtilsTestSuite) TestKubernetesLogContext() {

	logger := NewLogger(Debug, nil, nil)
	logger = WithK8sContext(logger)
	logHandler, ok := logger.(*LogHandler)
	suite.True(ok)
	suite.Len(logHandler.context.values, 0)

	os.Setenv("K8S_NODE_NAME", "Node-Name")
	os.Setenv("K8S_POD_NAME", "Pod-Name")

	logger2 := NewLogger(Debug, nil, nil)
	logger2 = WithK8sContext(logger2)
	logHandler2, ok2 := logger2.(*LogHandler)
	suite.True(ok2)
	suite.Len(logHandler2.context.values, 2)
}

func (suite *UtilsTestSuite) TestAppendContextValues() {

	logger := NewLogger(Debug, nil, nil)
	context := make(map[string]string)
	context["test"] = "val1"
	logger = AppendContextValues(logger, context)

	logHandler, ok := logger.(*LogHandler)
	suite.True(ok)
	suite.Len(logHandler.context.values, 1)
	suite.Equal("test:val1", logHandler.context.String())

	context2 := make(map[string]string)
	context2["test-3"] = "val2"
	logger = AppendContextValues(logger, context2)
	suite.Len(logHandler.context.values, 2)
}

func (suite *UtilsTestSuite) TestAppendFromLambdaContext() {

	logger := NewLogger(Debug, nil, nil)
	lambdaContext := lambdaContextForTest(context.Background())
	logger = AppendFromLambdaContext(logger, lambdaContext)

	logHandler, ok := logger.(*LogHandler)
	suite.True(ok)
	suite.Len(logHandler.context.values, 1)
	_, ok1 := logHandler.context.values[LogCtxRequestId]
	suite.True(ok1)

	logger2 := NewLogger(Debug, nil, nil)
	logger2 = AppendFromLambdaContext(logger2, context.Background())

	logHandler2, ok2 := logger2.(*LogHandler)
	suite.True(ok2)
	suite.Len(logHandler2.context.values, 0)
}
