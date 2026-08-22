package log

import (
	"strings"
	"sync"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

// ShipperFactory constructs a formatter/shipper pair for a given config and
// secrets manager. See RegisterShipper.
type ShipperFactory func(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper)

// shipperRegistry holds shippers registered via RegisterShipper, keyed by
// lowercased name. Guarded by a mutex since RegisterShipper isn't guaranteed
// to only ever run from an init() function (single-threaded, before main) -
// nothing stops a consumer from calling it later, concurrently with
// NewLoggerFromConfig/resolveShipperFromConfig reading it.
var shipperRegistry = struct {
	mu        sync.RWMutex
	factories map[string]ShipperFactory
}{factories: map[string]ShipperFactory{}}

// RegisterShipper registers a shipper factory under the given name, so
// NewLoggerFromConfig can select it via the "log.shipper" config value.
//
// This exists so optional shippers with their own, potentially heavy
// dependencies (e.g. the Logz.io shipper needs net/http and everything that
// pulls in) can live in their own package instead of always being compiled
// into every consumer of this one. Activate a shipper by blank-importing its
// package, which registers it from an init() function:
//
//	import _ "github.com/tommzn/go-log/v2/logzio"
//
// If "log.shipper" names a shipper that hasn't been registered this way -
// e.g. "logzio" without that blank import - NewLoggerFromConfig falls back
// to StdoutShipper/DefaultFormatter, same as if "log.shipper" were unset.
func RegisterShipper(name string, factory ShipperFactory) {
	shipperRegistry.mu.Lock()
	defer shipperRegistry.mu.Unlock()
	shipperRegistry.factories[strings.ToLower(name)] = factory
}

// resolveShipperFromConfig picks a formatter/shipper pair based on the
// "log.shipper" config value, falling back to DefaultFormatter/StdoutShipper
// if it's unset or names a shipper that hasn't been registered.
func resolveShipperFromConfig(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper) {

	if shipperType := conf.Get("log.shipper", nil); shipperType != nil {
		shipperRegistry.mu.RLock()
		factory, ok := shipperRegistry.factories[strings.ToLower(*shipperType)]
		shipperRegistry.mu.RUnlock()
		if ok {
			return factory(conf, secretsManager)
		}
	}
	return newDefaultFormatter(), newStdoutShipper()
}
