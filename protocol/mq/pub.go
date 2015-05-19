package mq

import (
	"time"

	"github.com/funkygao/nano"
)

type pub struct {
	sock nano.ProtocolSocket
}

func (this *pub) Init(sock nano.ProtocolSocket) {
	this.sock = sock
}

func (this *pub) AddEndpoint(ep nano.Endpoint) {

}

func (this *pub) RemoveEndpoint(ep nano.Endpoint) {

}

func (this *pub) Shutdown(expire time.Time) {

}

func (this *pub) SetOption(name string, val interface{}) error {
	return nil
}

func (this *pub) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*pub) Number() uint16 {
	return nano.ProtoMq
}

func (*pub) PeerNumber() uint16 {
	return nano.ProtoMq
}

func NewPubSocket() nano.Socket {
	return nano.MakeSocket(&pub{})
}
