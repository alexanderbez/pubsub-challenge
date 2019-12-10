package pubsub

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

type (
	// Message defines a type alias for an interface and represents arbitrary
	// messages that a Publisher may publish and a Consumer may receive.
	Message interface {
		fmt.Stringer
	}

	// PubSub defines a minimal interface for a publisher-subscriber (PubSub) model.
	// A PubSub must be able to registry arbitrary Producers which each Producer
	// publishes messages on a unique topic.
	//
	// Clients must be able to subscribe to topics using a topic pattern where the
	// pattern may contain wildcards. Subscriptions must return a channel from which
	// clients can read from.
	PubSub interface {
		RegisterProducer(topic string, producer Producer) error

		// TODO: consider returning a concrete type for richer functionality
		Subscribe(topicPattern string) <-chan Message
	}

	// BasePubSub implements a simple PubSub model where internally it maintains
	// a set of BaseProducers per topic where each topic is unique and a set of
	// BaseConsumers such that the relationship between BaseProducer and
	// BaseConsumer is one-to-one. Each internal BaseConsumer maintains an arbitrary
	// set of subscriptions for a given topic pattern.
	BasePubSub struct {
		mu sync.Mutex

		producers map[string]*BaseProducer
	}
)

func NewBasePubSub() PubSub {
	return &BasePubSub{
		producers: make(map[string]*BaseProducer),
	}
}

// RegisterProducer attempts to register a Producer for a given topic. A
// BaseConsumer is created for each Producer and starts to listen for new
// Messages. The Producer must be of type BaseProducer. An error will be returned
// if the topic is invalid or if a Producer already registered for that topic.
func (bps *BasePubSub) RegisterProducer(topic string, producer Producer) error {
	if err := ValidTopic(topic); err != nil {
		return err
	}

	bps.mu.Lock()
	defer bps.mu.Unlock()

	if _, ok := bps.producers[topic]; ok {
		return fmt.Errorf("producer for topic '%s' already registered", topic)
	}

	baseProducer, ok := producer.(*BaseProducer)
	if !ok {
		return fmt.Errorf("unexpected producer type: %T", producer)
	}

	go func() { baseProducer.loopBroadcast() }()

	bps.producers[topic] = baseProducer

	return nil
}

// Subscribe will create and return a new subscription (read-only Message channel)
// for a topic pattern. It will find all matching topics (if any exist) where
// the subscription will receive Messages from each associated Producer. If no
// topics exist for that pattern, the subscription will never receive any
// Messages even if a matching topic is created later.
//
// Note, there is no guarantee when the subscription will be added to any given
// producer so the client must not rely on timing. This allows Subscribe to not
// take abnormal amount of time when adding subscriptions.
//
// TODO:
// 1. Allow subscriptions to be added when a matching producer is registered at
// a later time.
// 2. Allow newly registered producers to match against existing subscriptions
func (bps *BasePubSub) Subscribe(topicPattern string) <-chan Message {
	bps.mu.Lock()
	defer bps.mu.Unlock()

	subscription := make(chan Message)

	for topic, producer := range bps.producers {
		// TODO: Use of a modified radix trie would provide significant improvement
		// if the number of producers is extremely large.
		if MatchTopic(topic, topicPattern) {
			// Add the subscription to the producer in a goroutine as to not have the
			// Subscribe call hang for longer than necessary.
			go func(producer *BaseProducer) {
				producer.AddSubscription(subscription)
			}(producer)
		}
	}

	return subscription
}
