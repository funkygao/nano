package xpubsub

import (
	"time"

	"github.com/funkygao/nano"
)

type xpub struct {
	sock nano.ProtocolSocket
}

func (this *xpub) Init(sock nano.ProtocolSocket) {
	this.sock = sock
}

func (this *xpub) AddEndpoint(ep nano.Endpoint) {

}

func (this *xpub) RemoveEndpoint(ep nano.Endpoint) {

}

func (this *xpub) Shutdown(expire time.Time) {

}

func (this *xpub) SetOption(name string, val interface{}) error {
	return nil
}

func (this *xpub) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*xpub) Number() uint16 {
	return nano.ProtoXPub
}

func (*xpub) PeerNumber() uint16 {
	return nano.ProtoXSub
}

func MakeXPubSocket() nano.Socket {
	return nano.MakeSocket(&xpub{})
}
