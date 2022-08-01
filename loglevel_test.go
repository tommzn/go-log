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
	suite.Equal(Status, LogLevelByName("Status"))
	suite.Equal(None, LogLevelByName("XXX"))
}

func (suite *LogLevelTestSuite) TestLogLevelFromEnv() {

	conf := loadConfigFromFile("config/testconfig.yml")
	suite.Equal(Debug, LogLevelFromConfig(conf))

	conf = loadConfigFromFile("config/empty.yml")
	suite.Equal(Error, LogLevelFromConfig(conf))
}

func (suite *LogLevelTestSuite) TestLogLevelFromConfig() {

	os.Setenv(ENV_LOGLEVEL, "info")
	suite.Equal(Info, LogLevelFromEnv())

	os.Unsetenv(ENV_LOGLEVEL)
}

func (suite *LogLevelTestSuite) TestSyslogLevel() {

	suite.Equal(0, None.SyslogLevel())
	suite.Equal(0, Status.SyslogLevel())
	suite.Equal(3, Error.SyslogLevel())
	suite.Equal(6, Info.SyslogLevel())
	suite.Equal(7, Debug.SyslogLevel())
}
