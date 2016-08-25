// Package kafka implements an eio producer and consumer for kafka.
//
// You must import this package separately to make these available.
package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/cio/eio"
)

// Producer implements a kafka producer
type Producer struct {
	ap   sarama.AsyncProducer
	errs chan error

	Args struct {
		// List of brokers to connect to
		Brokers []string

		// Name of the topic to push to. This is only necessary when not
		// creating a TopicProducer.
		Topic string
	}
}

func init() {
	eio.RegisterProducer("kafka", newProducer)
}

func newProducer(args eio.Args) (eio.Producer, error) {
	p := &Producer{
		errs: make(chan error, 8),
	}

	err := args.ApplyTo(&p.Args)
	if err != nil {
		return nil, err
	}

	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForLocal

	p.ap, err = sarama.NewAsyncProducer(p.Args.Brokers, cfg)

	if err == nil {
		go p.drainErrors()
	} else {
		p = nil
	}

	return p, err
}

func (p *Producer) drainErrors() {
	defer close(p.errs)

	for err := range p.ap.Errors() {
		p.errs <- err.Err
	}
}

// Produce implements Producer.Produce
func (p *Producer) Produce(b []byte) {
	p.ProduceTo(p.Args.Topic, b)
}

// ProduceTo implements TopicProducer.ProduceTo
func (p *Producer) ProduceTo(topic string, b []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(b),
	}

	p.ap.Input() <- msg
}

// Errs implements Producer.Errs
func (p *Producer) Errs() <-chan error {
	return p.errs
}

// Rotate implements Producer.Rotate
func (p *Producer) Rotate() error { return nil }

// Close implements Producer.Close
func (p *Producer) Close() (es cog.Errors) {
	if p.ap != nil {
		p.ap.AsyncClose()
		es.Drain(p.errs)
	}

	return
}
