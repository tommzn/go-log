package log

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambdacontext"
	config "github.com/tommzn/go-config"
	utils "github.com/tommzn/go-utils"
)

// testShipper is a mock for testing with an internal message stack. Guarded by
// a mutex so it can be shared across goroutines in concurrency tests, same as
// the real shippers.
type testShipper struct {
	mu       sync.Mutex
	messages []string
}

func newTestShipper() LogShipper {
	return &testShipper{messages: []string{}}
}

func (shipper *testShipper) Send(message string) {
	shipper.mu.Lock()
	defer shipper.mu.Unlock()
	shipper.messages = append(shipper.messages, message)
}

func (shipper *testShipper) Flush() {
	fmt.Println("Test Shipper flush!")
}

func loadConfigFromFile(fileName string) config.Config {
	configSource := config.NewFileConfigSource(&fileName)
	config, _ := configSource.Load()
	return config
}

func lambdaContextForTest(ctx context.Context) context.Context {
	return lambdacontext.NewContext(ctx, &lambdacontext.LambdaContext{AwsRequestID: utils.NewId()})
}
