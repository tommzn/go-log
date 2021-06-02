package log

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (suite *ContextTestSuite) TestCreateLogContext() {

	context := newLogContext(suite.contextValuesForTest())
	suite.IsType(LogContext{}, context)
	suite.Len(context.values, 2)
}

func (suite *ContextTestSuite) TestLogContextWithValues() {

	ctx := LogContextWithValues(context.Background(), suite.contextValuesForTest())
	suite.Implements((*context.Context)(nil), ctx)
	logContext, ok := ctx.Value(LOG_CONTEXT_KEY).(LogContext)
	suite.True(ok)
	suite.Len(logContext.values, 2)

	additionalContextValues := make(map[string]string)
	additionalContextValues["Key3"] = "Value3"
	ctx2 := LogContextWithValues(ctx, additionalContextValues)
	logContext2, ok2 := ctx2.Value(LOG_CONTEXT_KEY).(LogContext)
	suite.True(ok2)
	suite.Len(logContext2.values, 3)
}

func (suite *ContextTestSuite) TestGetLogContext() {

	ctx := LogContextWithValues(context.Background(), suite.contextValuesForTest())
	logContext := getLogContext(ctx)
	suite.Len(logContext.values, 2)

	logContext2 := getLogContext(context.Background())
	suite.Len(logContext2.values, 0)
}

func (suite *ContextTestSuite) TestAppendValues() {

	additionalContextValues := make(map[string]string)
	additionalContextValues["Key3"] = "Value3"

	logContext := newLogContext(suite.contextValuesForTest())
	suite.Len(logContext.values, 2)
	logContext.AppendValues(additionalContextValues)
	suite.Len(logContext.values, 3)
}

func (suite *ContextTestSuite) TestDefaultContextForNodes() {

	logContext := DefaultContextForNodes()
	suite.Len(logContext.values, 2)
	_, ok1 := logContext.values[LogCtx_Hostname]
	suite.True(ok1)
	_, ok2 := logContext.values[LogCtx_Ip]
	suite.True(ok2)
}

func (suite *ContextTestSuite) TestDefaultContextForK8s() {

	os.Setenv("K8S_NODE_NAME", "Node1")
	os.Setenv("K8S_POD_NAME", "Pod1")

	logContext := DefaultContextForK8s()
	suite.Len(logContext.values, 2)
	_, ok1 := logContext.values[LogCtx_K8s_Node]
	suite.True(ok1)
	_, ok2 := logContext.values[LogCtx_K8s_Pod]
	suite.True(ok2)

	os.Unsetenv("K8S_NODE_NAME")
	os.Unsetenv("K8S_POD_NAME")
}

func (suite *ContextTestSuite) contextValuesForTest() map[string]string {

	values := make(map[string]string)
	values["Key1"] = "Value1"
	values["Key2"] = "Value2"
	return values
}
