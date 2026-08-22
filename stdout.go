package log

import "fmt"

func newStdoutShipper() LogShipper {
	return &StdoutShipper{}
}

// Send print given log message in stdout.
func (shipper *StdoutShipper) Send(message string) {
	fmt.Println(message)
}

// Flush is not necessary for StdoutShipper, because it
// prints all log messages directly.
func (shipper *StdoutShipper) Flush() {
	fmt.Println("Stdout shipper flush!")
}
