package clog

import "fmt"

// BlackholeOutput drops all messages
type BlackholeOutput struct{}

func init() {
	RegisterOutputter("blackhole",
		FormatterConfig{Name: "Human"},
		func(ConfigArgs, Formatter) (Outputter, error) {
			return BlackholeOutput{}, nil
		})
}

// FormatEntry does nothing
func (BlackholeOutput) FormatEntry(e Entry) ([]byte, error) { return nil, nil }

func (BlackholeOutput) Write(b []byte) error { return nil }

// Rotate implements Outputter.Rotate
func (BlackholeOutput) Rotate() error { return nil }

// Exit implements Outputter.Exit
func (BlackholeOutput) Exit() {}

func (BlackholeOutput) String() string {
	return fmt.Sprintf("Blackhole")
}
