[![Go Reference](https://pkg.go.dev/badge/github.com/tommzn/go-log/v2.svg)](https://pkg.go.dev/github.com/tommzn/go-log/v2)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tommzn/go-log)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/tommzn/go-log)
[![Go Report Card](https://goreportcard.com/badge/github.com/tommzn/go-log/v2)](https://goreportcard.com/report/github.com/tommzn/go-log/v2)

# go-log

A flexible, pluggable logging library for Go applications. Supports multiple log levels, swappable formatters and shippers, and first-class integration with AWS Lambda, Kubernetes, and [Logz.io](https://logz.io).

The core package (stdout logging, context, levels) has no dependency on `net/http` or anything it pulls in - Logz.io support lives in a separate [`logzio`](./logzio) subpackage you only pay for if you import it. See [Package layout](#package-layout) below.

> **v2**: this is a new major version - update your import path from `github.com/tommzn/go-log` to `github.com/tommzn/go-log/v2`. If you used the Logz.io shipper, add `import _ "github.com/tommzn/go-log/v2/logzio"` (or construct it explicitly, see below) - `LogzioShipper`/`LogzioJsonFormatter` moved out of the core package into `logzio.Shipper`/`logzio.Formatter`. Everything else - `NewLogger`, `NewLoggerFromConfig`, `Logger`, log levels, context handling - is unchanged.

## Features

- Five log levels: `None`, `Status`, `Error`, `Info`, `Debug`
- Pluggable formatters: plain text or (via the `logzio` subpackage) Logz.io-compatible JSON
- Pluggable shippers: stdout or (via the `logzio` subpackage) async batched delivery to Logz.io
- Context-aware logging with key/value metadata
- AWS Lambda request ID injection
- Kubernetes node/pod metadata injection
- Configuration-driven setup via YAML
- Safe for concurrent use: a single logger can be shared across goroutines - `WithContext`/`WithFields` return a new logger rather than mutating the shared one

## Installation

```sh
# Core package - stdout logging, no net/http dependency
go get github.com/tommzn/go-log/v2

# Logz.io shipper (only when you need it)
go get github.com/tommzn/go-log/v2/logzio
```

## Usage

### Basic logger

```go
import log "github.com/tommzn/go-log/v2"

logger := log.NewLogger(log.Debug, nil, nil)

logger.Status("server started")
logger.Infof("listening on port %d", 8080)
logger.Errorf("unexpected error: %s", err)
logger.Debug("request payload", payload)
```

### Log levels

| Level | Description |
|-------|-------------|
| `None` | Disables all logging |
| `Status` | General status messages |
| `Error` | Errors |
| `Info` | Informational messages |
| `Debug` | Development/diagnostic output |

Only messages at or below the configured level are emitted. For example, a logger set to `Info` will emit `Status`, `Error`, and `Info` messages, but suppress `Debug`.

### Logger methods

Each level has two variants — a variadic form and a `fmt.Sprintf`-style form:

```go
logger.Status(v ...interface{})
logger.Statusf(message string, v ...interface{})

logger.Error(v ...interface{})
logger.Errorf(message string, v ...interface{})

logger.Info(v ...interface{})
logger.Infof(message string, v ...interface{})

logger.Debug(v ...interface{})
logger.Debugf(message string, v ...interface{})

// Generic form — pass the level explicitly
logger.Log(log.Info, "message")
logger.Logf(log.Error, "failed: %s", err)
```

### Log context

Attach key/value metadata to every log message via a standard `context.Context`.
`WithContext` and `WithFields` return a *new* Logger rather than modifying the
one they're called on, so a single base logger can safely be shared across
goroutines - derive a per-request logger from it instead of mutating it in place:

```go
logger := log.NewLogger(log.Debug, nil, nil) // one shared base logger

// per request/goroutine:
ctx := log.LogContextWithValues(context.Background(), map[string]string{
    "requestid": "abc-123",
    "env":       "production",
})
requestLogger := logger.WithContext(ctx)
requestLogger.Info("context is now attached to this logger only")

// or attach fields directly, without going through a context.Context:
requestLogger = logger.WithFields(map[string]string{"requestid": "abc-123"})
```

### Utility helpers

Like `WithContext`/`WithFields`, all of these return a new Logger and leave the one
passed in unchanged:

```go
// Attach a namespace label
logger = log.WithNameSpace(logger, "auth-service")

// Inject Kubernetes node/pod from K8S_NODE_NAME / K8S_POD_NAME env vars
logger = log.WithK8sContext(logger)

// Inject AWS Lambda request ID from a Lambda context
logger = log.AppendFromLambdaContext(logger, lambdaCtx)

// Attach arbitrary key/value pairs
logger = log.AppendContextValues(logger, map[string]string{"version": "1.2.3"})
```

### Pre-built context helpers

```go
// Returns a LogContext populated with hostname and IPv4 address
ctx := log.DefaultContextForNodes()

// Returns a LogContext with K8s node/pod read from environment variables
ctx := log.DefaultContextForK8s()
```

## Configuration

### From environment variable

Set the `LOGLEVEL` environment variable to one of `status`, `error`, `info`, or `debug`:

```sh
export LOGLEVEL=debug
```

```go
level := log.LogLevelFromEnv()
logger := log.NewLogger(level, nil, nil)
```

### From YAML config

Use [go-config](https://github.com/tommzn/go-config) to load configuration and pass it to the factory:

```go
logger := log.NewLoggerFromConfig(conf, secretsManager)
```

**Stdout (default):**
```yaml
log:
  loglevel: info
```

**Logz.io:**

```yaml
log:
  loglevel: debug
  shipper: logzio
  logzio:
    url: https://listener.logz.io:8071/  # default
    batchsize: 10                         # messages per batch, default: 10
    shipmentstacksize: 2                  # parallel workers, default: 2
    messagestacksize: 500                 # internal buffer size, default: 500
    shipmenttimeout: 1                    # worker acquire timeout in seconds, default: 1s
    messagereadtimeout: 50ms             # batch read timeout, default: 50ms
    timeout: 10                            # per-request HTTP client timeout in seconds, default: 10s
```

`timeout` bounds a single request to Logz.io (connection, TLS handshake, request write and response read combined). Without it, a hung request would block its shipment worker forever - with only `shipmentstacksize` workers (2 by default), that few hangs are enough to jam the shipper permanently, silently dropping every log from then on.

`log.shipper: logzio` only takes effect if the `logzio` subpackage has been activated with a blank import - `NewLoggerFromConfig` doesn't know about it otherwise, and falls back to stdout:

```go
import (
    log "github.com/tommzn/go-log/v2"
    _ "github.com/tommzn/go-log/v2/logzio" // activates "shipper: logzio"
)

logger := log.NewLoggerFromConfig(conf, secretsManager)
```

Or construct the Logz.io shipper/formatter explicitly, without going through config-driven dispatch at all:

```go
import (
    log "github.com/tommzn/go-log/v2"
    "github.com/tommzn/go-log/v2/logzio"
)

logger := log.NewLogger(log.Debug, logzio.NewFormatter(), logzio.NewShipper(conf, secretsManager))
```

The Logz.io authentication token is read at shipment time from a secrets manager under the key `LOGZIO_TOKEN`. See [go-secrets](https://github.com/tommzn/go-secrets) for configuration options.

## Package layout

- `github.com/tommzn/go-log/v2` - `Logger`, `LogHandler`, `DefaultFormatter`, `StdoutShipper`, levels, context. No `net/http` dependency.
- `github.com/tommzn/go-log/v2/logzio` - `Formatter` and `Shipper` for Logz.io. Pulls in `net/http` (and the TLS/certificate/IDNA machinery that comes with it) - only compiled into your binary if you actually import this package.

`LogFormatter` and `LogShipper` are plain exported interfaces, so you can plug in your own shipper the same way `logzio` does - implement `Send(string)`/`Flush()` or `Format(LogLevel, LogContext, string) string`, and optionally call `log.RegisterShipper(name, factory)` from an `init()` to hook it into `NewLoggerFromConfig`'s `log.shipper` dispatch.

## Shippers

| Shipper | Description |
|---------|-------------|
| `log.StdoutShipper` | Prints each message immediately to stdout (default) |
| `logzio.Shipper` | Buffers messages and ships them in async batches to Logz.io |

## Formatters

| Formatter | Output |
|-----------|--------|
| `log.DefaultFormatter` | `<Level>: <message>, Context: <key:value,...>` |
| `logzio.Formatter` | JSON with `@timestamp`, `loglevel`, `message`, and all context fields |

## Requirements

- Go 1.25.9+
