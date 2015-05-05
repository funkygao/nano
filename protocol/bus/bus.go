// Package bus implements the BUS protocol.  In this protocol, participants
// send a message to each of their peers.
package bus

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type busEp struct {
	ep nano.Endpoint
	q  chan *nano.Message
	x  *bus
}

type bus struct {
	sock  nano.ProtocolSocket
	peers map[uint32]*busEp
	raw   bool
	w     nano.Waiter
	init  sync.Once

	sync.Mutex
}

// Init implements the Protocol Init method.
func (x *bus) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.peers = make(map[uint32]*busEp)
	x.w.Init()
}

func (x *bus) Shutdown(expire time.Time) {

	x.w.WaitAbsTimeout(expire)

	x.Lock()
	peers := x.peers
	x.peers = make(map[uint32]*busEp)
	x.Unlock()

	for id, peer := range peers {
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
		delete(peers, id)
	}
}

// Bottom sender.
func (pe *busEp) peerSender() {
	for {
		m := <-pe.q
		if m == nil {
			return
		}
		if pe.ep.SendMsg(m) != nil {
			m.Free()
			return
		}
	}
}

func (x *bus) broadcast(m *nano.Message, sender uint32) {

	x.Lock()
	for id, pe := range x.peers {
		if sender == id {
			continue
		}
		m = m.Dup()

		select {
		case pe.q <- m:
		default:
			// No room on outbound queue, drop it.
			// Note that if we are passing on a linger/shutdown
			// notification and we can't deliver due to queue
			// full, it means we will wind up waiting the full
			// linger time in the lower sender.  Its correct, if
			// suboptimal, behavior.
			m.Free()
		}
	}
	x.Unlock()
}

func (x *bus) sender() {
	sq := x.sock.SendChannel()
	cq := x.sock.CloseChannel()
	defer x.w.Done()
	for {
		var id uint32
		select {
		case <-cq:
			return
		case m := <-sq:
			// If a header was present, it means this message is
			// being rebroadcast.  It should be a pipe ID.
			if len(m.Header) >= 4 {
				id = binary.BigEndian.Uint32(m.Header)
				m.Header = m.Header[4:]
			}
			x.broadcast(m, id)
			m.Free()
		}
	}
}

func (pe *busEp) receiver() {

	rq := pe.x.sock.RecvChannel()
	cq := pe.x.sock.CloseChannel()

	for {
		m := pe.ep.RecvMsg()
		if m == nil {
			return
		}
		v := pe.ep.Id()
		m.Header = append(m.Header,
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))

		select {
		case rq <- m:
		case <-cq:
			m.Free()
			return
		default:
			// No room, so we just drop it.
			m.Free()
		}
	}
}

func (x *bus) AddEndpoint(ep nano.Endpoint) {
	x.init.Do(func() {
		x.w.Add()
		go x.sender()
	})
	// Set our broadcast depth to match upper depth -- this should
	// help avoid dropping when bursting, if we burst before we
	// context switch.
	depth := 16
	if i, err := x.sock.GetOption(nano.OptionWriteQLen); err == nil {
		depth = i.(int)
	}
	pe := &busEp{ep: ep, x: x, q: make(chan *nano.Message, depth)}
	x.Lock()
	x.peers[ep.Id()] = pe
	x.Unlock()
	go pe.peerSender()
	go pe.receiver()
}

func (x *bus) RemoveEndpoint(ep nano.Endpoint) {
	x.Lock()
	if peer := x.peers[ep.Id()]; peer != nil {
		close(peer.q)
		delete(x.peers, ep.Id())
	}
	x.Unlock()
}

func (*bus) Number() uint16 {
	return nano.ProtoBus
}

func (*bus) Name() string {
	return "bus"
}

func (*bus) PeerNumber() uint16 {
	return nano.ProtoBus
}

func (*bus) PeerName() string {
	return "bus"
}

func (x *bus) RecvHook(m *nano.Message) bool {
	if !x.raw && len(m.Header) >= 4 {
		m.Header = m.Header[4:]
	}
	return true
}

func (x *bus) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if x.raw, ok = v.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (x *bus) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewSocket allocates a new Socket using the BUS protocol.
func NewSocket() nano.Socket {
	return nano.MakeSocket(&bus{})
}
