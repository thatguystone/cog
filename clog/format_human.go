package clog

import (
	"bytes"
	"fmt"
	"time"
)

// HumanFormat formats log entries so that a human can quickly decipher them
type HumanFormat struct {
	Args struct {
		ShortTime bool
	}
}

var startTime = time.Now()

// FormatEntry implements Formatter
func (f HumanFormat) FormatEntry(e Entry) ([]byte, error) {
	b := bytes.Buffer{}

	timeS := ""
	if f.Args.ShortTime {
		timeS = fmt.Sprintf("%0.6d", e.Time.Sub(startTime)/time.Second)
	} else {
		timeS = e.Time.Format(time.StampMicro)
	}

	fmt.Fprintf(&b, "[%s] %s : %c-%s : %-44s",
		timeS,
		e.Src,
		e.Level.Rune(),
		e.Module,
		e.Msg)

	for k, v := range e.Data {
		fmt.Fprintf(&b, " data.%s=%#v", k, v)
	}

	return b.Bytes(), nil
}
