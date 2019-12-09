# PubSub Challenge

[![GolangCI](https://golangci.com/badges/github.com/alexanderbez/pubsub-challenge.svg)](https://golangci.com)
[![GoDoc](https://godoc.org/github.com/alexanderbez/pubsub-challenge?status.svg)](https://godoc.org/github.com/alexanderbez/pubsub-challenge)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexanderbez/pubsub-challenge)](https://goreportcard.com/report/github.com/alexanderbez/pubsub-challenge)

## Problem

Create a new publisher-subscriber (PubSub) that allows multiple producers
publishing to multiple consumers. The consumers can subscribe to messages by
specifying a topic pattern (supporting wildcards) that it wants to receive
messages from, and producers can only write to a specific topic (i.e one to many).

For example, two consumers where one subscribes to `cosmosA-events-eventA` and
another subscribes to `cosmosA-events-*` will both receive the message that was
published by producers writing to `cosmosA-events-eventA`.

A consumer can start subscribing to events before there is any publisher, and
will start to receive data once a publisher is created and publishes to that
topic.

## Solution

There exists a `PubSub` implementation, `BasePubSub`. The `BasePubSub` allows for
any number of producers to be registered. Each producer may publish messages to
a single unique topic. Internally, the `BasePubSub` maintains a list of subscribers.
Clients may then subscribe to messages using a topic pattern. For each matching
topic, the subscription will be added to the producer's list of subscriptions and
those messages will be sent out on each subscription channel (which is returned to the client).

Note, the `BaseProducer` type allows for buffered publishing. If the buffer/queue is
full, the producer will error on `Publish`.

e.g.

```ascii
+------------------+             +----------------------+
| producer (a.b.c) +------------>+ subscription (*.*.c) |
+------------------+             +----------------------+
                         +-------^
                         |
                         |
+------------------+     |       +----------------------+
| producer (x.y.c) +-----+------>+ subscription (x.y.*) |
+------------------+             +----------------------+


+------------------+             +----------------------+
|producer(foo/bar) +------------>+ subscription (foo/*) |
+------------------+             +----------------------+

```

### Potential Improvements

* Consider returning a richer concrete type for `Subscribe` (e.g the ability to close).
* Consider use of a modified radix trie which would provide significant improvement
  if the number of producers is extremely large.

## Assumptions

Valid topics consist of arbitrary-length alphanumeric characters separated by a
valid deliminator: `/`,`.`,`-`.
