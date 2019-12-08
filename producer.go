package pubsub

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	// Producer defines a contract which a producer in a pubsub model must implement.
	Producer interface {
		Publish(msg Message) error
	}

	// BaseProducer implements the Producer interface by implementing basic publishing
	// capabilities on a single topic.
	BaseProducer struct {
		ch     chan Message
		logger zerolog.Logger
	}
)

func NewBaseProducer(topic string, capacity uint) Producer {
	return &BaseProducer{
		ch:     make(chan Message, capacity),
		logger: log.With().Str("module", "producer").Str("topic", topic).Logger(),
	}
}

// Publish will attempt to publish the provided Message. It will return an error
// if the publisher's internal queue is full.
func (bp *BaseProducer) Publish(msg Message) error {
	select {
	case bp.ch <- msg:
		bp.logger.Debug().Str("message", msg.String()).Str("action", "publishing message").Msg("")
		return nil
	default:
		return fmt.Errorf("publisher queue is full; failed to publish message %X; please try again", msg)
	}
}
