package log

import (
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
	logger = appendContextValues(logger, context)

	logHandler, ok := logger.(*LogHandler)
	suite.True(ok)
	suite.Equal("test:val1", logHandler.context.String())
}
