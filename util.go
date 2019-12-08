package pubsub

import (
	"errors"
	"regexp"
)

var separatorsRegex = regexp.MustCompile(`/|\.|-`)

// Topic errors
var (
	ErrEmptyTopic   = errors.New("topic cannot be empty")
	ErrInvalidTopic = errors.New("invalid topic")
)

// ValidTopic returns an error if a topic is invalid and nil otherwise. A valid
// topic consists of arbitrary-length alphanumeric characters separated by a valid
// deliminator: /,.,-
func ValidTopic(topic string) error {
	if len(topic) == 0 {
		return ErrEmptyTopic
	}

	for _, c := range topic {
		if (c > 47 && c < 58) ||
			(c > 64 && c < 91) ||
			(c > 96 && c < 123) ||
			c == 45 ||
			c == 46 ||
			c == 47 {
			continue
		} else {
			return ErrInvalidTopic
		}
	}

	return nil
}

// MatchTopic returns true if a given pattern matches the provided topic and
// returns false otherwise. The pattern may contain arbitrary wildcards '*'. It
// is assumed the topic is valid.
//
// Example:
// ok := MatchTopic("a.b.c", "a.b.c") // ok: true
// ok := MatchTopic("a.b.c", "a.*.c") // ok: true
// ok := MatchTopic("a.b.c", "a.*.*") // ok: true
// ok := MatchTopic("a.b.c", "c.b.a") // ok: false
// ok := MatchTopic("a.b.c", "a.*") // ok: false
func MatchTopic(topic, pattern string) bool {
	splitTopic := separatorsRegex.Split(topic, -1)
	splitPattern := separatorsRegex.Split(pattern, -1)

	if len(splitPattern) != len(splitTopic) {
		return false
	}

	for i, s := range splitPattern {
		if s != "*" && s != splitTopic[i] {
			return false
		}
	}

	return true
}
