package survey

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type resp struct {
	sock      nano.ProtocolSocket
	peers     map[nano.EndpointId]*respPeer
	raw       bool
	ttl       int
	backbuf   []byte
	backtrace []byte
	w         nano.Waiter
	init      sync.Once
	sync.Mutex
}

type respPeer struct {
	q  chan *nano.Message
	ep nano.Endpoint
	x  *resp
}

func (x *resp) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.ttl = 8
	x.peers = make(map[nano.EndpointId]*respPeer)
	x.w.Init()
	x.backbuf = make([]byte, 0, 64)
	x.sock.SetSendError(nano.ErrProtoState)
}

func (x *resp) Shutdown(expire time.Time) {
	peers := make(map[nano.EndpointId]*respPeer)
	x.w.WaitAbsTimeout(expire)
	x.Lock()
	for id, peer := range x.peers {
		delete(x.peers, id)
		peers[id] = peer
	}
	x.Unlock()

	for id, peer := range peers {
		delete(peers, id)
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
	}
}

func (x *resp) sender() {
	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, we'll just drop the message on the floor.

	defer x.w.Done()
	cq := x.sock.CloseChannel()
	sq := x.sock.SendChannel()
	for {
		var m *nano.Message
		select {
		case m = <-sq:
		case <-cq:
			return
		}

		// Lop off the 32-bit peer/pipe ID.  If absent, drop.
		if len(m.Header) < 4 {
			m.Free()
			continue
		}

		id := nano.EndpointId(binary.BigEndian.Uint32(m.Header))
		m.Header = m.Header[4:]

		x.Lock()
		peer := x.peers[id]
		x.Unlock()

		if peer == nil {
			m.Free()
			continue
		}

		// Put it on the outbound queue
		select {
		case peer.q <- m:
		default:
			// Backpressure, drop it.
			m.Free()
		}
	}
}

// When sending, we should have the survey ID in the header.
func (peer *respPeer) sender() {
	for {
		m := <-peer.q
		if m == nil {
			break
		}
		if peer.ep.SendMsg(m) != nil {
			m.Free()
			break
		}
	}
}

func (x *resp) receiver(ep nano.Endpoint) {

	rq := x.sock.RecvChannel()
	cq := x.sock.CloseChannel()

	for {
		m := ep.RecvMsg()
		if m == nil {
			return
		}

		v := ep.Id()
		m.Header = append(m.Header,
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
		hops := 0

		for {
			if hops >= x.ttl {
				m.Free() // ErrTooManyHops
				return
			}
			hops++
			if len(m.Body) < 4 {
				m.Free()
				continue
			}
			m.Header = append(m.Header, m.Body[:4]...)
			m.Body = m.Body[4:]
			if m.Header[len(m.Header)-4]&0x80 != 0 {
				break
			}
		}

		select {
		case rq <- m:
		case <-cq:
			m.Free()
			return
		}
	}
}

func (x *resp) RecvHook(m *nano.Message) bool {
	if x.raw {
		// Raw mode receivers get the message unadulterated.
		return true
	}

	if len(m.Header) < 4 {
		return false
	}

	x.Lock()
	x.backbuf = x.backbuf[0:0] // avoid allocations
	x.backtrace = append(x.backbuf, m.Header...)
	x.Unlock()
	x.sock.SetSendError(nil)
	return true
}

func (x *resp) SendHook(m *nano.Message) bool {
	if x.raw {
		// Raw mode senders expected to have prepared header already.
		return true
	}
	x.sock.SetSendError(nano.ErrProtoState)
	x.Lock()
	m.Header = append(m.Header[0:0], x.backtrace...)
	x.backtrace = nil
	x.Unlock()
	if len(m.Header) == 0 {
		return false
	}
	return true
}

func (x *resp) AddEndpoint(ep nano.Endpoint) {
	x.init.Do(func() {
		x.w.Add()
		go x.sender()
	})
	peer := &respPeer{ep: ep, x: x, q: make(chan *nano.Message, 1)}

	x.Lock()
	x.peers[ep.Id()] = peer
	x.Unlock()

	go x.receiver(ep)
	go peer.sender()
}

func (x *resp) RemoveEndpoint(ep nano.Endpoint) {
	x.Lock()
	delete(x.peers, ep.Id())
	x.Unlock()
}

func (*resp) Number() uint16 {
	return nano.ProtoRespondent
}

func (*resp) PeerNumber() uint16 {
	return nano.ProtoSurveyor
}

func (x *resp) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if x.raw, ok = v.(bool); !ok {
			return nano.ErrBadValue
		}
		if x.raw {
			x.sock.SetSendError(nil)
		} else {
			x.sock.SetSendError(nano.ErrProtoState)
		}
		return nil
	case nano.OptionTtl:
		if ttl, ok := v.(int); !ok {
			return nano.ErrBadValue
		} else if ttl < 1 || ttl > 255 {
			return nano.ErrBadValue
		} else {
			x.ttl = ttl
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (x *resp) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewRespondentSocket allocates a new Socket using the RESPONDENT protocol.
func NewRespondentSocket() nano.Socket {
	return nano.MakeSocket(&resp{})
}
