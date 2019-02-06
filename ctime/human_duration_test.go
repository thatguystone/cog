package ctime

import (
	"encoding/json"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestHumanDuration(t *testing.T) {
	c := check.New(t)

	v := struct {
		A HumanDuration
	}{}

	b := []byte(`{"A":"10s"}`)

	err := json.Unmarshal(b, &v)
	c.Nil(err)
	c.Equal(v.A, Second*10)

	res, err := json.Marshal(v)
	c.Nil(err)
	c.Equal(string(res), string(b))
}

func TestHumanDurationFallback(t *testing.T) {
	c := check.New(t)

	v := struct {
		A HumanDuration
	}{}

	err := json.Unmarshal([]byte(`{"A":10000000000}`), &v)
	c.Nil(err)
	c.Equal(v.A, Second*10)

	res, err := json.Marshal(v)
	c.Nil(err)
	c.Equal(string(res), `{"A":"10s"}`)
}
