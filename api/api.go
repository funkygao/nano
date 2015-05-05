package api

import (
	"github.com/funkygao/nano"
)

type PubSocket struct {
	*Socket
}

func NewPubSocket() (*PubSocket, error) {
	s, err := NewSocket(AF_SP, PUB)
	return &PubSocket{Socket: s}, err
}

type SubSocket struct {
	*Socket
}

func NewSubSocket() (*SubSocket, error) {
	s, err := NewSocket(AF_SP, SUB)
	return &SubSocket{Socket: s}, err
}

func (this *SubSocket) Subscribe(topic string) error {
	return this.sock.SetOption(nano.OptionSubscribe, topic)
}

func (this *SubSocket) Unsubscribe(topic string) error {
	return this.sock.SetOption(nano.OptionUnsubscribe, topic)
}
