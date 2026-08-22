package log

import (
	"strings"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

// ShipperFactory constructs a formatter/shipper pair for a given config and
// secrets manager. See RegisterShipper.
type ShipperFactory func(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper)

// shipperFactories holds shippers registered via RegisterShipper, keyed by
// lowercased name.
var shipperFactories = map[string]ShipperFactory{}

// RegisterShipper registers a shipper factory under the given name, so
// NewLoggerFromConfig can select it via the "log.shipper" config value.
//
// This exists so optional shippers with their own, potentially heavy
// dependencies (e.g. the Logz.io shipper needs net/http and everything that
// pulls in) can live in their own package instead of always being compiled
// into every consumer of this one. Activate a shipper by blank-importing its
// package, which registers it from an init() function:
//
//	import _ "github.com/tommzn/go-log/logzio"
//
// If "log.shipper" names a shipper that hasn't been registered this way -
// including built-in ones like "logzio" - NewLoggerFromConfig falls back to
// StdoutShipper/DefaultFormatter, same as if "log.shipper" were unset.
func RegisterShipper(name string, factory ShipperFactory) {
	shipperFactories[strings.ToLower(name)] = factory
}

// resolveShipperFromConfig picks a formatter/shipper pair based on the
// "log.shipper" config value, falling back to DefaultFormatter/StdoutShipper
// if it's unset or names a shipper that hasn't been registered.
func resolveShipperFromConfig(conf config.Config, secretsManager secrets.SecretsManager) (LogFormatter, LogShipper) {

	if shipperType := conf.Get("log.shipper", nil); shipperType != nil {
		if factory, ok := shipperFactories[strings.ToLower(*shipperType)]; ok {
			return factory(conf, secretsManager)
		}
	}
	return newDefaultFormatter(), newStdoutShipper()
}
