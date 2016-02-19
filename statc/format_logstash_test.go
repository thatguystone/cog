package statc

import (
	"encoding/json"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestLogstashFormatBasic(t *testing.T) {
	c := check.New(t)

	fmttr, err := newFormatter("logstash", nil)
	c.MustNotError(err)

	b, err := fmttr.Format(Snapshot{
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
	c.Equal(m["@version"], float64(1))
	c.Equal(m["int"], float64(123))
}
