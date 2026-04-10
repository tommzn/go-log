[![Go Reference](https://pkg.go.dev/badge/github.com/tommzn/go-utils.svg)](https://pkg.go.dev/github.com/tommzn/go-log)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tommzn/go-log)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/tommzn/go-log)
[![Go Report Card](https://goreportcard.com/badge/github.com/tommzn/go-log)](https://goreportcard.com/report/github.com/tommzn/go-log)

# go-log

A flexible, pluggable logging library for Go applications. Supports multiple log levels, swappable formatters and shippers, and first-class integration with AWS Lambda, Kubernetes, and [Logz.io](https://logz.io).

## Features

- Five log levels: `None`, `Status`, `Error`, `Info`, `Debug`
- Pluggable formatters: plain text or Logz.io-compatible JSON
- Pluggable shippers: stdout or async batched delivery to Logz.io
- Context-aware logging with key/value metadata
- AWS Lambda request ID injection
- Kubernetes node/pod metadata injection
- Configuration-driven setup via YAML
- Thread-safe for concurrent use

## Installation

```sh
go get github.com/tommzn/go-log
```

## Usage

### Basic logger

```go
import log "github.com/tommzn/go-log"

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

Attach key/value metadata to every log message via a standard `context.Context`:

```go
ctx := log.LogContextWithValues(context.Background(), map[string]string{
    "requestid": "abc-123",
    "env":       "production",
})
logger.WithContext(ctx)
logger.Info("context is now attached to all subsequent messages")
```

### Utility helpers

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
```

The Logz.io authentication token is read at shipment time from a secrets manager under the key `LOGZIO_TOKEN`. See [go-secrets](https://github.com/tommzn/go-secrets) for configuration options.

## Shippers

| Shipper | Description |
|---------|-------------|
| `StdoutShipper` | Prints each message immediately to stdout (default) |
| `LogzioShipper` | Buffers messages and ships them in async batches to Logz.io |

## Formatters

| Formatter | Output |
|-----------|--------|
| `DefaultFormatter` | `<Level>: <message>, Context: <key:value,...>` |
| `LogzioJsonFormatter` | JSON with `@timestamp`, `loglevel`, `message`, and all context fields |

## Requirements

- Go 1.25.9+
