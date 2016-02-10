package kafka

import "github.com/thatguystone/cog"

// A MockProducer implements Producer.
type MockProducer struct {
	Pending chan MockMsg
	exit    *cog.GExit
}

// MockMsg is what you get from Pending
type MockMsg struct {
	Topic string
	Msg   []byte
}

func newMockProducer(opts ProducerOpts) *MockProducer {
	return &MockProducer{
		Pending: make(chan MockMsg, 32),
		exit:    opts.Exit,
	}
}

// Bytes implements Producer.Bytes()
func (mp *MockProducer) Bytes(topic string, b []byte) {
	mm := MockMsg{
		Topic: topic,
		Msg:   b,
	}

	select {
	case mp.Pending <- mm:
	case <-mp.exit.C:
	}
}
