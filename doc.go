/*
Create a new publisher-subscriber (PubSub) that allows multiple producers
publishing to multiple consumers. The consumers can subscribe to messages by
specifying a topic pattern (supporting wildcards) that it wants to receive
messages from, and producers can only write to a specific topic (i.e one to many).

For example, two consumers where one subscribes to “cosmosA-events-eventA” and
another subscribes to “cosmosA-events-*” will both receive the message that was
published by producers writing to “cosmosA-events-eventA”.

A consumer can start subscribing to events before there is any publisher, and
will start to receive data once a publisher is created and publishes to that
topic.
*/
package pubsub
