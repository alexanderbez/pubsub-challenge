package pubsub_test

import (
	"testing"

	pubsub "github.com/alexanderbez/pubsub-challenge"
	"github.com/stretchr/testify/require"
)

func TestBaseConsumer_AddSubscription(t *testing.T) {
	c := pubsub.NewBaseConsumer(nil)
	require.Equal(t, c.TotalSubscriptions(), 0)
	c.AddSubscription(nil)
	require.Equal(t, c.TotalSubscriptions(), 1)
}
