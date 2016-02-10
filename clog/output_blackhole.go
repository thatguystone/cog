package clog

import (
	"fmt"

	"github.com/thatguystone/cog/config"
)

// BlackholeOutput drops all messages
type BlackholeOutput struct{}

func init() {
	RegisterOutputter("blackhole",
		FormatterConfig{Name: "Human"},
		func(config.Args, Formatter) (Outputter, error) {
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
