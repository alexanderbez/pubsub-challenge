package pubsub

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	// Producer defines a contract which a producer in a pubsub model must implement.
	Producer interface {
		Publish(msg Message) error
		TotalSubscriptions() int
	}

	// subscription defines a type for managing a subscription. It contains the
	// subscription's topic pattern and the read-only Message channel.
	subscription struct {
		ch      chan Message
		pattern string
	}

	// BaseProducer implements the Producer interface by implementing basic publishing
	// capabilities on a single topic.
	BaseProducer struct {
		mu sync.RWMutex

		logger        zerolog.Logger
		ch            chan Message
		subscriptions []subscription
	}
)

func NewBaseProducer(topic string, capacity uint) Producer {
	return &BaseProducer{
		ch:            make(chan Message, capacity),
		subscriptions: make([]subscription, 0),
		logger:        log.With().Str("module", "producer").Str("topic", topic).Logger(),
	}
}

// TotalSubscriptions returns the total number of subscriptions the producer
// currently has.
func (bp *BaseProducer) TotalSubscriptions() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return len(bp.subscriptions)
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

// addSubscription adds a subscription (read-only Message channel) to the
// producer.
func (bp *BaseProducer) addSubscription(s subscription) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.subscriptions = append(bp.subscriptions, s)
}

// matchSubscription checks if a new producer's topic matches against any
// subscription of bp, where upon match, the subscription is added to the new
// producer.
func (bp *BaseProducer) matchSubscription(topic string, newProducer *BaseProducer) {
	for _, subscription := range bp.subscriptions {
		if MatchTopic(topic, subscription.pattern) {
			go func(newProducer *BaseProducer) {
				newProducer.addSubscription(subscription)
			}(newProducer)
		}
	}
}

// loopBroadcast starts a blocking broadcast loop for incoming Message objects
// from the producer and sends it to each subscriber's (write-only) Message
// channel.
func (bp *BaseProducer) loopBroadcast() {
	for msg := range bp.ch {
		bp.mu.RLock()
		bp.logger.Debug().Str("message", msg.String()).Str("action", "received message from producer").Msg("")

		for _, s := range bp.subscriptions {
			go func(s chan Message, msg Message) {
				s <- msg
				bp.logger.Debug().Str("message", msg.String()).Str("action", "sent message to subscription").Msg("")
			}(s.ch, msg)
		}

		bp.mu.RUnlock()
	}
}
