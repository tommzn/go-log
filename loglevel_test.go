package log

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LogLevelTestSuite struct {
	suite.Suite
}

func TestLogLevelTestSuite(t *testing.T) {
	suite.Run(t, new(LogLevelTestSuite))
}

func (suite *LogLevelTestSuite) TestLogLevelNames() {

	suite.Equal("Error", fmt.Sprintf("%s", Error))
	suite.Equal("Info", fmt.Sprintf("%s", Info))
	suite.Equal("Debug", fmt.Sprintf("%s", Debug))
	suite.Equal("None", fmt.Sprintf("%s", None))
}

func (suite *LogLevelTestSuite) TestGetLogLevelByName() {

	suite.Equal(Error, LogLevelByName("Error"))
	suite.Equal(Info, LogLevelByName("info"))
	suite.Equal(Debug, LogLevelByName("DeBuG"))
	suite.Equal(None, LogLevelByName("XXX"))
}

func (suite *LogLevelTestSuite) TestLogLevelFromEnv() {

	os.Setenv(ENV_LOGLEVEL, "info")
	suite.Equal(Info, LogLevelFromEnv())

	os.Unsetenv(ENV_LOGLEVEL)
}
