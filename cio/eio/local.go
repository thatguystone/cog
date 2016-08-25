package eio

import (
	"fmt"
	"sync"

	"github.com/iheartradio/cog"
)

// LocalProducer is a channel-backed producer that works locally
type LocalProducer struct {
	*local

	Args struct {
		// Name of the local topic to push to
		Topic string
	}
}

// LocalConsumer works on the other side of LocalProducer, consuming local messages
type LocalConsumer struct {
	*local

	Args struct {
		// Name of the local topic to pull from
		Topic string
	}
}

type local struct {
	*localTopic
	once sync.Once
	exit chan struct{}
}

type localTopic struct {
	refs  int
	topic string
	ch    chan []byte
}

const localChSize = 128

var (
	localMtx    sync.Mutex
	localTopics = map[string]*localTopic{}
)

func init() {
	RegisterProducer("local",
		func(args Args) (Producer, error) {
			p := &LocalProducer{}

			err := args.ApplyTo(&p.Args)
			if err != nil {
				return nil, err
			}

			p.local = newLocal(p.Args.Topic)

			return p, nil
		})
	RegisterConsumer("local",
		func(args Args) (Consumer, error) {
			c := &LocalConsumer{}

			err := args.ApplyTo(&c.Args)
			if err != nil {
				return nil, err
			}

			c.local = newLocal(c.Args.Topic)

			return c, nil
		})
}

func newLocal(topic string) *local {
	l := &local{
		exit: make(chan struct{}),
	}

	localMtx.Lock()
	defer localMtx.Unlock()

	tpc, ok := localTopics[topic]
	if !ok {
		tpc = &localTopic{
			topic: topic,
			ch:    make(chan []byte, localChSize),
		}
		localTopics[topic] = tpc
	}

	tpc.refs++
	l.localTopic = tpc

	return l
}

func (l *local) close() (es cog.Errors) {
	l.once.Do(func() {
		localMtx.Lock()
		defer localMtx.Unlock()

		l.refs--
		if l.refs == 0 {
			delete(localTopics, l.topic)

			if len(l.ch) > 0 {
				es.Add(fmt.Errorf("all producers and consumers closed, "+
					"but %d messages left on local topic %s",
					len(l.ch), l.topic))
			}
		}

		close(l.exit)
	})

	return
}

// Produce implements Producer.Produce
func (p *LocalProducer) Produce(b []byte) {
	select {
	case p.ch <- b:
	case <-p.exit:
	}
}

// ProduceTo implements TopicProducer.ProduceTo
func (p *LocalProducer) ProduceTo(topic string, b []byte) {
	localMtx.Lock()
	lt := localTopics[topic]
	localMtx.Unlock()

	if lt != nil {
		select {
		case lt.ch <- b:
		case <-p.exit:
		}
	}
}

// Errs implements Producer.Errs
func (p *LocalProducer) Errs() <-chan error { return ClosedErrCh }

// Rotate implements Producer.Rotate
func (p *LocalProducer) Rotate() error { return nil }

// Close implements Producer.Close
func (p *LocalProducer) Close() cog.Errors {
	return p.local.close()
}

// Next implements Consumer.Next
func (c *LocalConsumer) Next() ([]byte, error) {
	select {
	case b := <-c.ch:
		return b, nil

	case <-c.exit:
		return nil, nil
	}
}

// Close implements Consumer.Close
func (c *LocalConsumer) Close() (es cog.Errors) {
	es = c.local.close()
	return
}
