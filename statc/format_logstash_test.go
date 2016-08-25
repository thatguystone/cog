package statc

import (
	"encoding/json"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/clog"
)

func TestLogstashFormatBasic(t *testing.T) {
	c := check.New(t)

	fmttr, err := newFormatter("logstash", nil)
	c.MustNotError(err)

	c.Contains(fmttr.MimeType(), "json")

	b, err := fmttr.FormatSnap(Snapshot{
		Stat{
			Name: "int",
			Val:  int64(123),
		},
	})
	c.MustNotError(err)

	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	c.MustNotError(err)

	c.NotEqual(m["@timestamp"], "")
	c.Equal(m["@host"], clog.Hostname())
	c.Equal(m["@version"], float64(1))
	c.Equal(m["int"], float64(123))
}

func TestLogstashFormatNoOverride(t *testing.T) {
	c := check.New(t)

	fmttr, err := newFormatter("logstash", nil)
	c.MustNotError(err)

	snap := make(Snapshot, 1, 8)
	snap[0] = Stat{
		Name: "int",
		Val:  int64(123),
	}

	_, err = fmttr.FormatSnap(snap)
	c.MustNotError(err)

	c.Equal(snap[0].Name, "int")
}
