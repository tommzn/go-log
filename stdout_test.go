package log

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StdoutShipperTestSuite struct {
	suite.Suite
}

func TestStdoutShipperTestSuite(t *testing.T) {
	suite.Run(t, new(StdoutShipperTestSuite))
}

func (suite *StdoutShipperTestSuite) TestShipper() {

	orig_out := os.Stdout

	logMessage := "Debug: Test Message, Context: key1,val1"

	shipper := newStdoutShipper()
	suite.IsType(&StdoutShipper{}, shipper)

	// capture from standard output
	r, w, _ := os.Pipe()
	os.Stdout = w

	shipper.send(logMessage)

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()
	out := <-outC
	suite.Equal(logMessage+"\n", out)

	// FLuah will have no effect, but should not throw any errors.
	shipper.flush()
	os.Stdout = orig_out
}
