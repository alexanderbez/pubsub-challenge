package pubsub_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	pubsub "github.com/alexanderbez/pubsub-challenge"
)

func TestValidTopic(t *testing.T) {
	testCases := []struct {
		topic     string
		expectErr bool
	}{
		{"a.b.c", false},
		{"foo", false},
		{"a.b-c/d", false},
		{"", true},
		{"a?b", true},
		{"©∂ç", true},
		{" foo", true},
	}

	for i, tc := range testCases {
		err := pubsub.ValidTopic(tc.topic)
		require.Equal(t, tc.expectErr, err != nil, "invalid result for tc #%d: %s", i, err)
	}
}

func TestMatchTopic(t *testing.T) {
	testCases := []struct {
		topic, pattern string
		result         bool
	}{
		{
			"a.b.c", "a.b.c", true,
		},
		{
			"a.b.c", "a.*.c", true,
		},
		{
			"a.b.c", "a.*.*", true,
		},
		{
			"a.b.c", "c.b.a", false,
		},
		{
			"a.b.c", "a.*", false,
		},
		{
			"a.b.c.d", "a.b.c", false,
		},
		{
			"a.b.c", "a.b.c.d", false,
		},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.result, pubsub.MatchTopic(tc.topic, tc.pattern), "invalid result for tc #%d", i)
	}
}
