// Package star implements a new, experimental protocol called "STAR".
// This is like the BUS protocol, except that each member of the network
// automatically forwards any message it receives to any other peers.
// In a star network, this means that all members should receive all messages,
// assuming that there is a central server.  Its important to ensure that
// the topology is free from cycles, as there is no protection against
// that, and cycles can lead to infinite message storms.  (TODO: Add a TTL,
// and basic message ID / anti-replay protection.)
package star

import (
	"sync"
	"time"

	nano "github.com/funkygao/nano"
)

type starEp struct {
	ep nano.Endpoint
	q  chan *nano.Message
	x  *star
}

type star struct {
	sock nano.ProtocolSocket
	eps  map[uint32]*starEp
	raw  bool
	w    nano.Waiter
	init sync.Once

	sync.Mutex
}

func (x *star) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.eps = make(map[uint32]*starEp)
	x.w.Init()
}

func (x *star) Shutdown(expire time.Time) {

	x.w.WaitAbsTimeout(expire)

	x.Lock()
	peers := x.eps
	x.eps = make(map[uint32]*starEp)
	x.Unlock()

	for id, peer := range peers {
		delete(peers, id)
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
	}
}

// Bottom sender.
func (pe *starEp) sender() {
	for {
		m := <-pe.q
		if m == nil {
			break
		}

		if pe.ep.SendMsg(m) != nil {
			m.Free()
			break
		}
	}
}

func (x *star) broadcast(m *nano.Message, sender *starEp) {

	x.Lock()
	if sender == nil || !x.raw {
		for _, pe := range x.eps {
			if sender == pe {
				continue
			}
			m = m.Dup()
			select {
			case pe.q <- m:
			default:
				// No room on outbound queue, drop it.
				if m != nil {
					m.Free()
				}
			}
		}
	}
	x.Unlock()

	// Grab a local copy and send it up if we aren't originator
	if sender != nil {
		select {
		case x.sock.RecvChannel() <- m:
		case <-x.sock.CloseChannel():
			m.Free()
			return
		default:
			// No room, so we just drop it.
			m.Free()
		}
	} else {
		// Not sending it up, so we need to release it.
		m.Free()
	}
}

func (x *star) sender() {
	defer x.w.Done()
	sq := x.sock.SendChannel()
	cq := x.sock.CloseChannel()

	for {
		select {
		case <-cq:
			return
		case m := <-sq:
			x.broadcast(m, nil)
		}
	}
}

func (pe *starEp) receiver() {
	for {
		msg := pe.ep.RecvMsg()
		if msg == nil {
			return
		}

		// if we're in raw mode, this does only a sendup, otherwise
		// it does both a retransmit + sendup
		pe.x.broadcast(msg, pe)
	}
}

func (x *star) AddEndpoint(ep nano.Endpoint) {
	x.init.Do(func() {
		x.w.Add()
		go x.sender()
	})
	depth := 16
	if i, err := x.sock.GetOption(nano.OptionWriteQLen); err == nil {
		depth = i.(int)
	}
	pe := &starEp{ep: ep, x: x, q: make(chan *nano.Message, depth)}
	x.Lock()
	x.eps[ep.Id()] = pe
	x.Unlock()
	go pe.sender()
	go pe.receiver()
}

func (x *star) RemoveEndpoint(ep nano.Endpoint) {
	x.Lock()
	if peer := x.eps[ep.Id()]; peer != nil {
		delete(x.eps, ep.Id())
		close(peer.q)
	}
	x.Unlock()
}

func (*star) Number() uint16 {
	return nano.ProtoStar
}

func (*star) PeerNumber() uint16 {
	return nano.ProtoStar
}

func (x *star) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if x.raw = v.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (x *star) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewSocket allocates a new Socket using the STAR protocol.
func NewSocket() nano.Socket {
	return nano.MakeSocket(&star{})
}
