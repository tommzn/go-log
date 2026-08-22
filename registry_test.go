package log

import (
	"testing"

	"github.com/stretchr/testify/suite"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

type RegistryTestSuite struct {
	suite.Suite
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (suite *RegistryTestSuite) TestRegisterShipperAndResolve() {

	defer delete(shipperFactories, "custom-test-shipper")

	registeredFormatter := newDefaultFormatter()
	registeredShipper := newStdoutShipper()
	RegisterShipper("Custom-Test-Shipper", func(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper) {
		return registeredFormatter, registeredShipper
	})

	conf := config.NewStaticConfigSource("log:\n  shipper: custom-test-shipper\n")
	loaded, err := conf.Load()
	suite.NoError(err)

	formatter, shipper := resolveShipperFromConfig(loaded, nil)
	suite.Same(registeredFormatter, formatter)
	suite.Same(registeredShipper, shipper)
}

func (suite *RegistryTestSuite) TestResolveFallsBackToStdoutWhenUnregistered() {

	conf := config.NewStaticConfigSource("log:\n  shipper: some-unregistered-shipper\n")
	loaded, err := conf.Load()
	suite.NoError(err)

	formatter, shipper := resolveShipperFromConfig(loaded, nil)
	suite.IsType(&DefaultFormatter{}, formatter)
	suite.IsType(&StdoutShipper{}, shipper)
}

func (suite *RegistryTestSuite) TestResolveFallsBackToStdoutWhenUnset() {

	conf := config.NewStaticConfigSource("log:\n  loglevel: debug\n")
	loaded, err := conf.Load()
	suite.NoError(err)

	formatter, shipper := resolveShipperFromConfig(loaded, nil)
	suite.IsType(&DefaultFormatter{}, formatter)
	suite.IsType(&StdoutShipper{}, shipper)
}
