package ctime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

func TestHumanDuration(t *testing.T) {
	gt := check.New(t)

	v := struct {
		A HumanDuration
	}{}

	b := []byte(`{"A":"10s"}`)

	err := json.Unmarshal(b, &v)
	gt.NotError(err)
	gt.Equal(v.A, time.Second*10)

	res, err := json.Marshal(v)
	gt.NotError(err)
	gt.Equal(string(res), string(b))
}

func TestHumanDurationFallback(t *testing.T) {
	gt := check.New(t)

	v := struct {
		A HumanDuration
	}{}

	err := json.Unmarshal([]byte(`{"A":10000000000}`), &v)
	gt.NotError(err)
	gt.Equal(v.A, time.Second*10)

	res, err := json.Marshal(v)
	gt.NotError(err)
	gt.Equal(string(res), `{"A":"10s"}`)
}
