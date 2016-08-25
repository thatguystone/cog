package eio

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/iheartradio/cog"
)

// MakeProducer creates a new Producer
type MakeProducer func(args Args) (Producer, error)

// A Producer writes messages
type Producer interface {
	// Send everything here
	Produce(b []byte)

	// Select on this, otherwise a producer might block
	Errs() <-chan error

	// Some producers rotate (files). Call this to get them to reopen any
	// underlying files.
	//
	// The error is returned here so that logrotate (and friends) can quickly
	// determine if rotation succeeded.
	Rotate() error

	// Tear the producer down and flush any pending messages.
	//
	// If you use NewProducer(), this is automatically called when the object
	// is GCd.
	Close() cog.Errors
}

type producer struct {
	Producer
}

// TopicProducer is just like Producer, except it allows you to send events to
// specific topics.
type TopicProducer interface {
	Producer

	// Send to the given topic
	ProduceTo(topic string, b []byte)
}

type topicProducer struct {
	TopicProducer
}

// MakeConsumer creates a new Consumer
type MakeConsumer func(args Args) (Consumer, error)

// A Consumer reads messages
type Consumer interface {
	// Get messages from here.
	Next() ([]byte, error)

	// Tear the consumer down and wait for it to exit.
	//
	// If you use NewConsumer(), this is automatically called when the object
	// is GCd.
	Close() cog.Errors
}

type consumer struct {
	Consumer
}

var (
	// ClosedErrCh should be returned from Errs() when no errors can be
	// produced. This ensures that any receiver immediately returns and doesn't
	// block forever.
	ClosedErrCh chan error

	regdPs = map[string]MakeProducer{}
	regdCs = map[string]MakeConsumer{}
)

func init() {
	ClosedErrCh = make(chan error)
	close(ClosedErrCh)
}

// RegisterProducer registers a Producer for use. Names are case insensitive.
//
// If your producer implements TopicProducer, it will automatically be made
// available.
func RegisterProducer(name string, np MakeProducer) {
	lname := strings.ToLower(name)

	if _, ok := regdPs[lname]; ok {
		panic(fmt.Errorf("producer `%s` already exists", name))
	}

	regdPs[lname] = np
}

func newProducer(name string, args Args) (Producer, error) {
	lname := strings.ToLower(name)

	np, ok := regdPs[lname]
	if !ok {
		return nil, fmt.Errorf("producer `%s` does not exist", name)
	}

	p, err := np(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer `%s`: %v", name, err)
	}

	return p, nil
}

// NewProducer creates a new producer with the given arguments
func NewProducer(name string, args Args) (Producer, error) {
	p, err := newProducer(name, args)
	if err != nil {
		return nil, err
	}

	pp := &producer{Producer: p}
	runtime.SetFinalizer(pp, finalizeProducer)

	return pp, nil
}

func (p *producer) Close() (es cog.Errors) {
	runtime.SetFinalizer(p, nil)
	return p.Producer.Close()
}

func finalizeProducer(p *producer) {
	go p.Close()
}

// NewTopicProducer creates a new producer, provided that the producer
// implements TopicProducer.
func NewTopicProducer(name string, args Args) (TopicProducer, error) {
	p, err := newProducer(name, args)
	if err != nil {
		return nil, err
	}

	tp, ok := p.(TopicProducer)
	if !ok {
		p.Close()
		return nil, fmt.Errorf("`%s` does not implement TopicProducer", name)
	}

	tpp := &topicProducer{TopicProducer: tp}
	runtime.SetFinalizer(tpp, finalizeTopicProducer)

	return tpp, nil
}

func (tp *topicProducer) Close() (es cog.Errors) {
	runtime.SetFinalizer(tp, nil)
	return tp.TopicProducer.Close()
}

func finalizeTopicProducer(tp *topicProducer) {
	go tp.Close()
}

// RegisterConsumer registers a Consumer for use. Names are case insensitive.
func RegisterConsumer(name string, nc MakeConsumer) {
	lname := strings.ToLower(name)

	if _, ok := regdCs[lname]; ok {
		panic(fmt.Errorf("consumer %s already exists", name))
	}

	regdCs[lname] = nc
}

// NewConsumer creates a new consumer with the given arguments
func NewConsumer(name string, args Args) (Consumer, error) {
	lname := strings.ToLower(name)

	nc, ok := regdCs[lname]
	if !ok {
		return nil, fmt.Errorf("consumer `%s` does not exist", name)
	}

	c, err := nc(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer `%s`: %v", name, err)
	}

	cp := &consumer{Consumer: c}
	runtime.SetFinalizer(cp, finalizeConsumer)

	return cp, nil
}

func (c *consumer) Close() (es cog.Errors) {
	runtime.SetFinalizer(c, nil)
	return c.Consumer.Close()
}

func finalizeConsumer(c *consumer) {
	go c.Close()
}
