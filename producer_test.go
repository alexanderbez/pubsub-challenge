package pubsub_test

import (
	"testing"

	pubsub "github.com/alexanderbez/pubsub-challenge"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestBaseProducer_Publish(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	capacity := 10
	p := pubsub.NewBaseProducer("test", uint(capacity))

	for i := 0; i < capacity; i++ {
		require.NoError(t, p.Publish(testMessage{"a.b.c", "test"}))
	}

	require.Error(t, p.Publish(testMessage{"a.b.c", "test"}))
}
