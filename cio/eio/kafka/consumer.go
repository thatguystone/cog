package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/cio/eio"
)

// Consumer implements a kafka consumer
type Consumer struct {
	c    sarama.Consumer
	msgs chan []byte
	errs chan error
	exit *cog.Exit

	Args struct {
		// List of brokers to connect to
		Brokers []string

		// Name of the topic to read from.
		Topic string
	}
}

func init() {
	eio.RegisterConsumer("kafka", newConsumer)
}

func newConsumer(args eio.Args) (eio.Consumer, error) {
	c := &Consumer{
		msgs: make(chan []byte, 32),
		errs: make(chan error, 8),
		exit: cog.NewExit(),
	}

	err := args.ApplyTo(&c.Args)
	if err != nil {
		return nil, err
	}

	c.c, err = sarama.NewConsumer(c.Args.Brokers, nil)
	if err != nil {
		return nil, err
	}

	parts, err := c.c.Partitions(c.Args.Topic)
	if err == nil {
		c.exit.Add(len(parts))
		for _, part := range parts {
			go c.consume(part)
		}

		go c.waitForExit()
	} else {
		c.waitForExit()
		c.Close()
		c = nil
	}

	return c, err
}

func (c *Consumer) waitForExit() {
	c.exit.Wait()
	c.c.Close()
	close(c.errs)
}

func (c *Consumer) consume(partID int32) {
	defer c.exit.Done()

	pc, err := c.c.ConsumePartition(
		c.Args.Topic,
		partID,
		sarama.OffsetNewest)
	if err != nil {
		c.errs <- err
		return
	}

	defer pc.Close()

	msgs := pc.Messages()
	errs := pc.Errors()

	for {
		select {
		case msg := <-msgs:
			select {
			case c.msgs <- msg.Value:
			case <-c.exit.C:
				return
			}

		case err := <-errs:
			c.errs <- err.Err

		case <-c.exit.C:
			return
		}
	}
}

// Next implemnets Consumer.Next
func (c *Consumer) Next() ([]byte, error) {
	select {
	case msg := <-c.msgs:
		return msg, nil

	case err := <-c.errs:
		return nil, err
	}
}

// Close implemnets Consumer.Close
func (c *Consumer) Close() (es cog.Errors) {
	c.exit.Signal()
	es.Drain(c.errs)
	return
}
