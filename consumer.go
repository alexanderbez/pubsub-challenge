package pubsub

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// BaseConsumer defines a consumer of messages from a Producer. A BasePubSub
// manages a one-to-one relationship between a BaseProducer and a BaseConsumer.
// Each BaseConsumer internally manages an arbitrary set of topic pattern
// subscribers.
type BaseConsumer struct {
	mu sync.RWMutex

	inCh          <-chan Message
	subscriptions []chan<- Message
	logger        zerolog.Logger
}

func NewBaseConsumer(inCh <-chan Message) *BaseConsumer {
	return &BaseConsumer{
		inCh:          inCh,
		subscriptions: make([]chan<- Message, 0),
		logger:        log.With().Str("module", "consumer").Logger(),
	}
}

// TotalSubscriptions returns the total number of subscriptions the consumer
// currently has.
func (bc *BaseConsumer) TotalSubscriptions() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.subscriptions)
}

// AddSubscription adds a subscription (read-only Message channel) to the
// consumer.
func (bc *BaseConsumer) AddSubscription(s chan<- Message) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.subscriptions = append(bc.subscriptions, s)
}

// listen starts a blocking listener for incoming Message objects from a
// BaseProducer's channel. For each Message, the consumer will send it to each
// subscriber's (write-only) Message channel.
func (bc *BaseConsumer) listen() {
	for msg := range bc.inCh {
		bc.mu.Lock()
		bc.logger.Debug().Str("message", msg.String()).Str("action", "received message from producer").Msg("")

		for _, s := range bc.subscriptions {
			go func(s chan<- Message, msg Message) {
				s <- msg
				bc.logger.Debug().Str("message", msg.String()).Str("action", "sent message to subscription").Msg("")
			}(s, msg)
		}

		bc.mu.Unlock()
	}
}
