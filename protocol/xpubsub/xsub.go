package xpubsub

import (
	"time"

	"github.com/funkygao/nano"
)

type xsub struct {
	sock nano.ProtocolSocket
}

func (this *xsub) Init(sock nano.ProtocolSocket) {
	this.sock = sock
}

func (this *xsub) AddEndpoint(ep nano.Endpoint) {

}

func (this *xsub) RemoveEndpoint(ep nano.Endpoint) {

}

func (this *xsub) Shutdown(expire time.Time) {

}

func (this *xsub) SetOption(name string, val interface{}) error {
	return nil
}

func (this *xsub) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*xsub) Name() string {
	return "sub"
}

func (*xsub) PeerName() string {
	return "pub"
}

func (*xsub) Handshake() bool {
	return true
}

func (*xsub) Number() uint16 {
	return nano.ProtoSub
}

func (*xsub) PeerNumber() uint16 {
	return nano.ProtoPub
}

func NewXSubSocket() nano.Socket {
	return nano.MakeSocket(&xsub{})
}
