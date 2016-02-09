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

func init() {
	RegisterFormatter("Human",
		func(args ConfigArgs) (Formatter, error) {
			f := HumanFormat{}
			err := args.ApplyTo(&f.Args)
			if err != nil {
				return nil, err
			}

			return f, nil
		})
}

// FormatEntry implements Formatter
func (f HumanFormat) FormatEntry(e Entry) ([]byte, error) {
	b := bytes.Buffer{}

	timeS := ""
	if f.Args.ShortTime {
		timeS = fmt.Sprintf("%0.6d", e.Time.Sub(startTime)/time.Second)
	} else {
		timeS = e.Time.Format(time.StampMicro)
	}

	fmt.Fprintf(&b, "[%s] %c-%s : %s : %-44s",
		timeS,
		e.Level.Rune(),
		e.Module,
		e.Src,
		e.Msg)

	for k, v := range e.Data {
		fmt.Fprintf(&b, " data.%s=%#v", k, v)
	}

	return bytes.TrimSpace(b.Bytes()), nil
}
