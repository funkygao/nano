package mq

import (
	"time"

	"github.com/funkygao/nano"
)

type sub struct {
	sock nano.ProtocolSocket
}

func (this *sub) Init(sock nano.ProtocolSocket) {
	this.sock = sock
}

func (this *sub) AddEndpoint(ep nano.Endpoint) {
	if err := handshake(ep); err != nil {
		nano.Debugf(err.Error())
		return
	}
}

func (this *sub) RemoveEndpoint(ep nano.Endpoint) {

}

func (this *sub) Shutdown(expire time.Time) {

}

func (this *sub) SetOption(name string, val interface{}) error {
	return nil
}

func (this *sub) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*sub) Number() uint16 {
	return nano.ProtoMq
}

func (*sub) PeerNumber() uint16 {
	return nano.ProtoMq
}

func NewSubSocket() nano.Socket {
	return nano.MakeSocket(&sub{})
}

func Sub(topic string) {

}
