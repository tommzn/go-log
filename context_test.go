package log

import (
	"context"
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
	logContext, ok := ctx.Value(logContextKey).(LogContext)
	suite.True(ok)
	suite.Len(logContext.values, 2)

	additionalContextValues := make(map[string]string)
	additionalContextValues["Key3"] = "Value3"
	ctx2 := LogContextWithValues(ctx, additionalContextValues)
	logContext2, ok2 := ctx2.Value(logContextKey).(LogContext)
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

	merged := logContext.AppendValues(additionalContextValues)

	// AppendValues returns a new LogContext rather than mutating the receiver.
	suite.Len(logContext.values, 2)
	suite.Len(merged.values, 3)
}

func (suite *ContextTestSuite) TestAppendValuesDoesNotMutatePassedMap() {

	logContext := newLogContext(suite.contextValuesForTest())

	additionalContextValues := make(map[string]string)
	additionalContextValues["Key3"] = "Value3"

	_ = logContext.AppendValues(additionalContextValues)

	suite.Len(additionalContextValues, 1)
	_, ok := additionalContextValues["Key1"]
	suite.False(ok)
}

func (suite *ContextTestSuite) TestLogContextWithValuesDoesNotMutateCallerMap() {

	// Regression test for GL-3: LogContextWithValues used to write the parent
	// context's values directly into the caller-supplied map.
	parentCtx := LogContextWithValues(context.Background(), map[string]string{"parent_key": "parent_val"})

	myValues := map[string]string{"foo": "bar"}
	_ = LogContextWithValues(parentCtx, myValues)

	suite.Equal(map[string]string{"foo": "bar"}, myValues)
}

func (suite *ContextTestSuite) TestDefaultContextForNodes() {

	logContext := DefaultContextForNodes()
	_, ok := logContext.values[LogCtxHostname]
	suite.True(ok)
	// IPv4 address may not be resolvable in all environments (e.g. IPv6-only CI)
	suite.True(len(logContext.values) >= 1)
}

func (suite *ContextTestSuite) TestDefaultContextForK8s() {

	suite.T().Setenv("K8S_NODE_NAME", "Node1")
	suite.T().Setenv("K8S_POD_NAME", "Pod1")

	logContext := DefaultContextForK8s()
	suite.Len(logContext.values, 2)
	_, ok1 := logContext.values[LogCtxK8sNode]
	suite.True(ok1)
	_, ok2 := logContext.values[LogCtxK8sPod]
	suite.True(ok2)
}

func (suite *ContextTestSuite) TestNewLogContextClonesInput() {

	original := map[string]string{"foo": "bar"}
	logContext := NewLogContext(original)
	suite.Equal("bar", logContext.values["foo"])

	// Mutating the caller-supplied map must not change the stored context.
	original["foo"] = "changed"
	suite.Equal("bar", logContext.values["foo"])

	// nil input is safe to pass and produces an empty (nil-cloned) context.
	empty := NewLogContext(nil)
	suite.Nil(empty.values)
}

func (suite *ContextTestSuite) TestLogContextValuesReturnsClone() {

	logContext := NewLogContext(map[string]string{"foo": "bar"})

	valuesCopy := logContext.Values()
	suite.Equal("bar", valuesCopy["foo"])

	// Mutating the returned map must not affect the stored context.
	valuesCopy["foo"] = "changed"
	valuesCopy["extra"] = "value"
	suite.Equal("bar", logContext.values["foo"])
	_, ok := logContext.values["extra"]
	suite.False(ok)
}

func (suite *ContextTestSuite) TestAppendValuesOnZeroContext() {

	// A zero LogContext has a nil values map; AppendValues must still work
	// (this exercises the merged == nil branch that was previously uncovered).
	var zero LogContext
	merged := zero.AppendValues(map[string]string{"a": "b"})
	suite.Equal("b", merged.values["a"])
}

func (suite *ContextTestSuite) contextValuesForTest() map[string]string {

	values := make(map[string]string)
	values["Key1"] = "Value1"
	values["Key2"] = "Value2"
	return values
}
