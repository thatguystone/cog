package eio

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/config"
)

// MakeProducer creates a new Producer
type MakeProducer func(args config.Args) (Producer, error)

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

// MakeConsumer creates a new Consumer
type MakeConsumer func(args config.Args) (Consumer, error)

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

var regdPs = map[string]MakeProducer{}
var regdCs = map[string]MakeConsumer{}

// RegisterProducer registers a Producer for use. Names are case insensitive.
func RegisterProducer(name string, np MakeProducer) {
	lname := strings.ToLower(name)

	if _, ok := regdPs[lname]; ok {
		panic(fmt.Errorf("producer %s already exists", name))
	}

	regdPs[lname] = np
}

// NewProducer creates a new producer with the given arguments
func NewProducer(name string, args config.Args) (Producer, error) {
	lname := strings.ToLower(name)

	np, ok := regdPs[lname]
	if !ok {
		return nil, fmt.Errorf("producer %s does not exist", name)
	}

	p, err := np(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer %s: %v", name, err)
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
	p.Close()
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
func NewConsumer(name string, args config.Args) (Consumer, error) {
	lname := strings.ToLower(name)

	nc, ok := regdCs[lname]
	if !ok {
		return nil, fmt.Errorf("consumer %s does not exist", name)
	}

	c, err := nc(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer %s: %v", name, err)
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
	c.Close()
}
