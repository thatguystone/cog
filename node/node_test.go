package node

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func TestResolveListenAddress(t *testing.T) {
	c := check.New(t)

	addr := ""
	err := ResolveListenAddress(&addr)
	c.MustError(err)

	addr = "nope"
	err = ResolveListenAddress(&addr)
	c.MustError(err)

	ec2md := testEC2MetadataServer()
	defer ec2md.Close()

	tests := []struct {
		in          string
		out         string
		err         bool
		ec2BaseAddr string
	}{
		{
			in:  "test:80",
			out: "test:80",
		},
		{
			in:  "<ec2-priv>:80",
			out: "EC2PRIV:80",
		},
		{
			in:  "<ec2-priv>",
			err: true,
		},
		{
			in:          "<ec2-priv>:80",
			ec2BaseAddr: "127.0.0.1:9999999",
			err:         true,
		},
		{
			in:  "",
			err: true,
		},
	}

	for i, test := range tests {
		ec2Base := ec2md.URL
		if test.ec2BaseAddr != "" {
			ec2Base = test.ec2BaseAddr
		}

		ec2 := EC2Metadata{
			base: ec2Base,
		}

		addr := test.in
		err := resolveListenAddress(&addr, &ec2)
		if test.err {
			c.Error(err, "failed at %d", i)
		} else {
			c.NotError(err, "failed at %d", i)
			c.Equal(test.out, addr, "failed at %d", i)
		}
	}
}
