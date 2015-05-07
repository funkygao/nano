package pubsub

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

func (this *xpub) Name() string {
	return "pub"
}

func (this *xpub) PeerName() string {
	return "sub"
}

func (this *xpub) Number() uint16 {
	return nano.ProtoPub
}

func (this *xpub) PeerNumber() uint16 {
	return nano.ProtoSub
}

func MakeXPubSocket() nano.Socket {
	return nano.MakeSocket(&xpub{})
}
