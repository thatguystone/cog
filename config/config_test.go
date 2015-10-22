package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/check"
)

type testCfger struct {
	One int
	Two int
}

func (t *testCfger) Validate(_ *Cfg, es *cog.Errors) {
	if t.One != 1 {
		es.Add(fmt.Errorf("`One` != 1 (==%d)", t.One))
	}

	if t.Two != 2 {
		es.Add(fmt.Errorf("`Two` != 2 (==%d)", t.Two))
	}
}

type testListener struct {
	Addr string
}

func (t *testListener) Validate(c *Cfg, es *cog.Errors) {
	c.ResolveListen(&t.Addr, es)
}

func TestMain(m *testing.M) {
	check.Main(m)
}

func testEC2MetadataServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest/meta-data/local-ipv4",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("EC2PRIV"))
		})

	return httptest.NewServer(mux)
}

func TestRegister(t *testing.T) {
	c := check.New(t)

	Register("TestRegister", func() Configer { return nil })
	c.Panic(func() {
		Register("TestRegister", nil)
	})
}

func TestNewCoverage(t *testing.T) {
	check.New(t)

	Register("TestNewCoverage", func() Configer { return nil })
	New()
}

func TestResolveListen(t *testing.T) {
	c := check.New(t)

	ec2md := testEC2MetadataServer()
	defer ec2md.Close()

	tests := []struct {
		in      string
		out     string
		err     bool
		ec2Addr string
	}{
		{
			in:  "test:80",
			out: "test:80",
		},
		{
			in:  "ec2:80",
			out: "EC2PRIV:80",
		},
		{
			in:  "ec2",
			err: true,
		},
		{
			in:      "ec2:80",
			ec2Addr: "127.0.0.1:9999999",
			err:     true,
		},
		{
			in:  "",
			err: true,
		},
	}

	for _, test := range tests {
		file := "config.json"
		c.FS.SWriteFile(file,
			fmt.Sprintf(`{"Listen": {"Addr": "%s"}}`,
				test.in))

		l := &testListener{}
		cfg := Cfg{
			Modules: Modules{
				"Listen": l,
			},
		}

		if test.ec2Addr != "" {
			cfg.ec2MetadataBase = test.ec2Addr
		} else {
			cfg.ec2MetadataBase = ec2md.URL
		}

		err := cfg.LoadAndValidate(c.FS.Path(file))
		if test.err {
			c.Error(err)
		} else {
			c.NotError(err)
			c.Equal(test.out, l.Addr)
		}
	}
}
