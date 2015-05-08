// Package pubsub implements a basic simple PUB/SUB protocol.
// PUB listens for subscriptions and publishes messages to subscribers (SUB peers).
// SUB will filter incoming messages from the publisher based on their
// subscription(see nano.OptionSubscribe).
package pubsub
