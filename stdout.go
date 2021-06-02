package log

import "fmt"

func newStdoutShipper() LogShipper {
	return &StdoutShipper{}
}

// Send print given log message in stdout.
func (shipper *StdoutShipper) send(message string) {
	fmt.Println(message)
}

// Flush is not necessary for StdoutShipper, because it
// prints all log messages directly.
func (shipper *StdoutShipper) flush() {
	fmt.Sprintln("Stdout shipper flush!")
}
