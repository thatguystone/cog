package kafka

import (
	"runtime/debug"

	"github.com/Shopify/sarama"
	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/clog"
)

// A Producer writes messages to topics
type Producer interface {
	Bytes(topic string, b []byte)
}

// ProducerOpts provides options for creating a new producer.
type ProducerOpts struct {
	Mock bool // Create a mock producer of type *MockProducer

	// Used only if not creating a mock producer
	Cl      sarama.Client // Client to use; you are responsible for closing this
	Brokers []string      // If not using a client, brokers to connect to
	Log     *clog.Logger
	Exit    *cog.GExit
}

type producer struct {
	cl   sarama.Client
	ap   sarama.AsyncProducer
	log  *clog.Logger
	exit *cog.GExit
}

// NewProducer creates a new Producer based on the given ProducerOpts
func NewProducer(opts ProducerOpts) (Producer, error) {
	if opts.Mock {
		return newMockProducer(opts), nil
	}

	var err error

	p := &producer{
		log:  opts.Log,
		exit: opts.Exit,
	}

	cl := opts.Cl
	if cl == nil {
		p.cl, err = sarama.NewClient(opts.Brokers, nil)
		cl = p.cl
	}

	if err == nil {
		err = p.init(cl)
	}

	if err != nil {
		p.close()
		p = nil
	}

	return p, err
}

func (p *producer) init(cl sarama.Client) (err error) {
	p.ap, err = sarama.NewAsyncProducerFromClient(cl)

	if err == nil {
		p.exit.Add(1)
		go p.run()
	}

	return
}

func (p *producer) close() {
	if p.ap != nil {
		p.ap.AsyncClose()
		for err := range p.ap.Errors() {
			p.logErr(err)
		}
	}

	if p.cl != nil {
		p.cl.Close()
	}
}

func (p *producer) run() {
	defer func() {
		p.close()
		p.exit.Done()
	}()

	for {
		select {
		case err := <-p.ap.Errors():
			p.logErr(err)

		case <-p.exit.C:
			return
		}
	}
}

func (p *producer) logErr(err *sarama.ProducerError) {
	p.log.Error("producer error: %v", err.Error())
}

func (p *producer) Bytes(topic string, b []byte) {
	// Can't use a select here: sarama closes the Input() channel, and any
	// sends on it panic. Hooray.
	defer func() {
		if err := recover(); err != nil {
			d := clog.Data{
				"stack": string(debug.Stack()),
			}
			p.log.Infod(d, "message sent after producer closed")
		}
	}()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(b),
	}

	p.ap.Input() <- msg
}
