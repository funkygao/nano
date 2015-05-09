// Package bus implements the GOSSIP protocol.  In this protocol, participants
// send a message to each of their peers.
package gossip

import (
	"time"

	"github.com/funkygao/nano"
)

type peerEp struct{}

type gossip struct {
	sock nano.ProtocolSocket

	maxPeers     int           // maximum number of connected gossip peers
	testInterval time.Duration //

	peers map[nano.EndpointId]peerEp
}

func (this *gossip) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.maxPeers = 10
	this.testInterval = time.Millisecond * 10
}

func (this *gossip) AddEndpoint(ep nano.Endpoint) {

}

func (this *gossip) RemoveEndpoint(ep nano.Endpoint) {

}

func (this *gossip) Shutdown(expire time.Time) {

}

func (this *gossip) SetOption(name string, val interface{}) error {
	return nil
}

func (this *gossip) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*gossip) Number() uint16 {
	return 0
}

func (*gossip) PeerNumber() uint16 {
	return 0
}

func NewSocket() nano.Socket {
	return nano.MakeSocket(&gossip{})
}
