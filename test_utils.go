package log

import (
	"fmt"
	"net/http"

	config "github.com/tommzn/go-config"
)

// testShipper is a mock for testing with an internal message stack.
type testShipper struct {
	messages []string
}

func newTestShipper() LogShipper {
	return &testShipper{messages: []string{}}
}

func (shipper *testShipper) send(message string) {
	shipper.messages = append(shipper.messages, message)
}

func (shipper *testShipper) flush() {
	fmt.Println("Test Shipper flush!")
}

// testClient is a HTTP client mock for testing.
type testClient struct {
	requests []*http.Request
	response *http.Response
	err      error
}

func newHttpTestClient(response *http.Response, err error) httpClient {
	return &testClient{response: response, err: err, requests: []*http.Request{}}
}

func (client *testClient) Do(req *http.Request) (*http.Response, error) {
	client.requests = append(client.requests, req)
	return client.response, client.err
}

func loadConfigFromFile(fileName string) config.Config {
	configSource := config.NewFileConfigSource(&fileName)
	config, _ := configSource.Load()
	return config
}
