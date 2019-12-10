package pubsub_test

import (
	"fmt"
	"sync"
	"testing"

	pubsub "github.com/alexanderbez/pubsub-challenge"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type testMessage struct {
	topic, data string
}

func (tm testMessage) String() string {
	return tm.data
}

func TestPubSub(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var (
		wg1 sync.WaitGroup
		wg2 sync.WaitGroup
	)

	ps := pubsub.NewBasePubSub()

	producers := []struct {
		topic       string
		producer    pubsub.Producer
		subscribers int
		numMsgs     int
	}{
		{"a.b.c", pubsub.NewBaseProducer("a.b.c", 1000), 1, 100},
		{"x.y.c", pubsub.NewBaseProducer("x.y.c", 1000), 2, 50},
		{"foo/bar", pubsub.NewBaseProducer("foo/bar", 1000), 1, 25},
	}

	subscriptions := []struct {
		pattern     string
		expectedMsg int
	}{
		{"*.*.c", 150},
		{"x.y.*", 50},
		{"foo/*", 125},
	}

	// register all producers
	for _, p := range producers {
		require.NoError(t, ps.RegisterProducer(p.topic, p.producer))
	}

	// Range over subscriptions and verify we receive the total number of expected
	// messages for each pattern.
	for _, sub := range subscriptions {
		s := ps.Subscribe(sub.pattern)
		require.NotNil(t, s)

		wg1.Add(1)
		wg2.Add(1)

		go func(wg1, wg2 *sync.WaitGroup, s <-chan pubsub.Message, pattern string, expectedMsg int) {
			wg2.Done()

			i := 0
			for range s {
				i++

				if i == expectedMsg {
					wg1.Done()
					return
				}
			}
		}(&wg1, &wg2, s, sub.pattern, sub.expectedMsg)
	}

	wg2.Wait()

	// publish messages for each producer
	for _, p := range producers {
		for p.producer.TotalSubscriptions() != p.subscribers {
			// Wait till goroutine scheduler added expected subscriptions. This is only
			// needed for deterministic testing.
		}

		for i := 0; i < p.numMsgs; i++ {
			msg := testMessage{p.topic, fmt.Sprintf("%s-%d", p.topic, i)}
			require.NoError(t, p.producer.Publish(msg))
		}
	}

	// Create a new producer and publish messages which should match against an
	// existing subscription (subscriptions[2]).
	newTopic := "foo/baz"
	newProducer := pubsub.NewBaseProducer(newTopic, 1000)
	require.NoError(t, ps.RegisterProducer(newTopic, newProducer))

	for newProducer.TotalSubscriptions() != 1 {
		// Will till the goroutine scheduler schedules the subscription addition to
		// the new producer.
	}

	for i := 0; i < 100; i++ {
		msg := testMessage{newTopic, fmt.Sprintf("%s-%d", newTopic, i)}
		require.NoError(t, newProducer.Publish(msg))
	}

	// All groups should finish assuming they received total number of expected
	// messages.
	wg1.Wait()
}
