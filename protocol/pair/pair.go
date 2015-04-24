// Package pair implements the PAIR protocol.  This protocol is a 1:1
// peering protocol.
package pair

import (
	"sync"
	"time"

	mangos "github.com/funkygao/nano"
)

type pair struct {
	sock mangos.ProtocolSocket
	peer mangos.Endpoint
	raw  bool
	w    mangos.Waiter
	sync.Mutex
}

func (x *pair) Init(sock mangos.ProtocolSocket) {
	x.sock = sock
	x.w.Init()
}

func (x *pair) Shutdown(expire time.Time) {
	x.w.WaitAbsTimeout(expire)
}

func (x *pair) sender(ep mangos.Endpoint) {

	defer x.w.Done()
	sq := x.sock.SendChannel()
	cq := x.sock.CloseChannel()

	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, we'll just drop the message on the floor.
	for {
		select {
		case m := <-sq:
			if ep.SendMsg(m) != nil {
				m.Free()
				return
			}
		case <-cq:
			return
		}
	}
}

func (x *pair) receiver(ep mangos.Endpoint) {

	rq := x.sock.RecvChannel()
	cq := x.sock.CloseChannel()

	for {
		m := ep.RecvMsg()
		if m == nil {
			return
		}

		select {
		case rq <- m:
		case <-cq:
			return
		}
	}
}

func (x *pair) AddEndpoint(ep mangos.Endpoint) {
	x.Lock()
	if x.peer != nil {
		x.Unlock()
		ep.Close()
		return
	}
	x.peer = ep
	x.Unlock()

	x.w.Add()
	go x.receiver(ep)
	go x.sender(ep)
}

func (x *pair) RemoveEndpoint(ep mangos.Endpoint) {
	x.Lock()
	if x.peer == ep {
		x.peer = nil
	}
	x.Unlock()
}

func (*pair) Number() uint16 {
	return mangos.ProtoPair
}

func (*pair) Name() string {
	return "pair"
}

func (*pair) PeerNumber() uint16 {
	return mangos.ProtoPair
}

func (*pair) PeerName() string {
	return "pair"
}

func (x *pair) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case mangos.OptionRaw:
		if x.raw, ok = v.(bool); !ok {
			return mangos.ErrBadValue
		}
		return nil
	default:
		return mangos.ErrBadOption
	}
}

func (x *pair) GetOption(name string) (interface{}, error) {
	switch name {
	case mangos.OptionRaw:
		return x.raw, nil
	default:
		return nil, mangos.ErrBadOption
	}
}

// NewProtocol returns a new PAIR protocol object.
func NewProtocol() mangos.Protocol {
	return &pair{}
}

// NewSocket allocates a new Socket using the PAIR protocol.
func NewSocket() (mangos.Socket, error) {
	return mangos.MakeSocket(&pair{}), nil
}
