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

		producers         map[string]*BaseProducer
		idleSubscriptions map[string]subscription
	}
)

func NewBasePubSub() PubSub {
	return &BasePubSub{
		producers:         make(map[string]*BaseProducer),
		idleSubscriptions: make(map[string]subscription),
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

	bps.matchNewProducer(topic, baseProducer)
	bps.producers[topic] = baseProducer
	bps.subscribeIdle()

	return nil
}

// matchNewProducer will check to see if any existing subscriptions match against
// a new producer's topic, where if they match, the subscription will be added
// to the new producer.
func (bps *BasePubSub) matchNewProducer(topic string, newProducer *BaseProducer) {
	for _, producer := range bps.producers {
		producer.matchSubscription(topic, newProducer)
	}
}

// subscribeIdle checks to see if any (idle) subscriptions can be matched and
// subscribed against existing producers.
func (bps *BasePubSub) subscribeIdle() {
	for pattern, idleSub := range bps.idleSubscriptions {
		var matched bool

		for topic, producer := range bps.producers {
			if MatchTopic(topic, pattern) {
				matched = true

				go func(producer *BaseProducer, idleSub subscription) {
					producer.addSubscription(idleSub)
				}(producer, idleSub)
			}
		}

		if matched {
			delete(bps.idleSubscriptions, pattern)
		}
	}
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
func (bps *BasePubSub) Subscribe(pattern string) <-chan Message {
	bps.mu.Lock()
	defer bps.mu.Unlock()

	subscription := subscription{
		ch:      make(chan Message),
		pattern: pattern,
	}

	var matched bool

	for topic, producer := range bps.producers {
		// TODO: Use of a modified radix trie would provide significant improvement
		// if the number of producers is extremely large.
		if MatchTopic(topic, pattern) {
			matched = true

			// Add the subscription to the producer in a goroutine as to not have the
			// Subscribe call hang for longer than necessary.
			go func(producer *BaseProducer) {
				producer.addSubscription(subscription)
			}(producer)
		}
	}

	if !matched {
		bps.idleSubscriptions[pattern] = subscription
	}

	return subscription.ch
}
