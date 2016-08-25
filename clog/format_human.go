package clog

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iheartradio/cog/cio/eio"
)

// HumanFormat formats log entries so that a human can quickly decipher them
type HumanFormat struct {
	Args struct {
		ShortTime bool // Format timestamp as "seconds since start"
	}
}

var startTime = time.Now()

func init() {
	RegisterFormatter("Human",
		func(args eio.Args) (Formatter, error) {
			f := HumanFormat{}
			err := args.ApplyTo(&f.Args)
			if err != nil {
				return nil, err
			}

			return f, nil
		})
}

// FormatEntry implements Formatter.FormatEntry
func (f HumanFormat) FormatEntry(e Entry) ([]byte, error) {
	b := bytes.Buffer{}

	timeS := ""
	if f.Args.ShortTime {
		timeS = fmt.Sprintf("%0.6d", e.Time.Sub(startTime)/time.Second)
	} else {
		timeS = e.Time.Format(time.StampMicro)
	}

	msg := ""
	if len(e.Msg) > 0 {
		msg = fmt.Sprintf(" %-44s", e.Msg)
	}

	fmt.Fprintf(&b, "[%s] %c-%s : %s :%s",
		timeS,
		e.Level.Rune(),
		e.Module,
		e.Src,
		msg)

	for k, v := range e.Data {
		fmt.Fprintf(&b, " data.%s=%#v", k, v)
	}

	return bytes.TrimSpace(b.Bytes()), nil
}

// MimeType implements Formatter.MimeType
func (HumanFormat) MimeType() string {
	return "text/plain"
}
