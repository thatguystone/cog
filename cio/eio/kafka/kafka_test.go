package kafka

import (
	"net"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
)

const kafkaAddr = "localhost:9092"

var skipTests = true

func init() {
	c, err := net.Dial("tcp", kafkaAddr)
	if err == nil {
		c.Close()
		skipTests = false
	}
}

func newTest(t *testing.T) *check.C {
	c := check.New(t)

	if skipTests {
		c.Skip("local kafka not available")
	}

	return c
}

func TestBasic(t *testing.T) {
	c := newTest(t)

	pr, err := eio.NewProducer("kafka", eio.Args{
		"brokers": []string{kafkaAddr},
		"topic":   check.GetTestName(),
	})
	c.MustNotError(err)

	co, err := eio.NewConsumer("kafka", eio.Args{
		"brokers": []string{kafkaAddr},
		"topic":   check.GetTestName(),
	})
	c.MustNotError(err)

	pr.Rotate()

	const smsg = "fancy message"
	pr.Produce([]byte(smsg))

	msg, err := co.Next()
	c.MustNotError(err)
	c.Equal(string(msg), smsg)

	es := pr.Close()
	c.NotError(es.Error())

	es = co.Close()
	c.NotError(es.Error())
}
