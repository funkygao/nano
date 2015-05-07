// Package pubsub implements a basic simple (X)PUB/(X)SUB protocol.
// (X)PUB listens for subscriptions and publishes messages
// to subscribers (SUB peers).
// (X)SUB will filter incoming messages from the publisher based on their
// subscription(see nano.OptionSubscribe).
package pubsub
