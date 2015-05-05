// Package pair implements the PAIR protocol.  This protocol is a two-way 1:1
// peering protocol.
package pair

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type pair struct {
	sock   nano.ProtocolSocket
	peer   nano.Endpoint
	raw    bool
	waiter nano.Waiter
	sync.Mutex
}

func (this *pair) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.waiter.Init()
}

func (this *pair) Shutdown(expire time.Time) {
	this.waiter.WaitAbsTimeout(expire)
}

func (this *pair) sender(endpoint nano.Endpoint) {
	defer this.waiter.Done()

	sendChan := this.sock.SendChannel()
	closeChan := this.sock.CloseChannel()

	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, we'll just drop the message on the floor.
	for {
		select {
		case msg := <-sendChan:
			if endpoint.SendMsg(msg) != nil {
				msg.Free()
				return
			}

		case <-closeChan:
			return
		}
	}
}

func (x *pair) receiver(endpoint nano.Endpoint) {
	recvChan := x.sock.RecvChannel()
	closeChan := x.sock.CloseChannel()

	for {
		msg := endpoint.RecvMsg()
		if msg == nil {
			return
		}

		select {
		case recvChan <- msg:
		case <-closeChan:
			return
		}
	}
}

func (this *pair) AddEndpoint(endpoint nano.Endpoint) {
	this.Lock()
	if this.peer != nil {
		// TODO not good design
		this.Unlock()
		endpoint.Close()
		return
	}

	this.peer = endpoint
	this.Unlock()

	this.waiter.Add()
	go this.receiver(endpoint)
	go this.sender(endpoint)
}

func (this *pair) RemoveEndpoint(endpoint nano.Endpoint) {
	this.Lock()
	if this.peer == endpoint {
		this.peer = nil
	}
	this.Unlock()
}

func (*pair) Number() uint16 {
	return nano.ProtoPair
}

func (*pair) Name() string {
	return "pair"
}

func (*pair) PeerNumber() uint16 {
	return nano.ProtoPair
}

func (*pair) PeerName() string {
	return "pair"
}

func (this *pair) SetOption(name string, val interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if this.raw, ok = val.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (this *pair) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return this.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewSocket allocates a new Socket using the PAIR protocol.
func NewSocket() nano.Socket {
	return nano.MakeSocket(&pair{})
}
