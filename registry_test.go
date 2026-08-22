package log

import (
	"sync"
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

	defer func() {
		shipperRegistry.mu.Lock()
		delete(shipperRegistry.factories, "custom-test-shipper")
		shipperRegistry.mu.Unlock()
	}()

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

// TestConcurrentRegisterAndResolve is a regression test for a data race
// flagged in review: RegisterShipper (write) and resolveShipperFromConfig
// (read) on shipperRegistry.factories with no synchronization. Run with
// -race; nothing here is meaningful without it.
func (suite *RegistryTestSuite) TestConcurrentRegisterAndResolve() {

	conf := config.NewStaticConfigSource("log:\n  shipper: concurrent-test-shipper\n")
	loaded, err := conf.Load()
	suite.NoError(err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			RegisterShipper("concurrent-test-shipper", func(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper) {
				return newDefaultFormatter(), newStdoutShipper()
			})
		}()
		go func() {
			defer wg.Done()
			resolveShipperFromConfig(loaded, nil)
		}()
	}
	wg.Wait()

	shipperRegistry.mu.Lock()
	delete(shipperRegistry.factories, "concurrent-test-shipper")
	shipperRegistry.mu.Unlock()
}
