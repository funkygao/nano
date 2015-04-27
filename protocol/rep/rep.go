// Package rep implements the REP protocol, which is the response side of
// the request/response pattern.  (REQ is the request.)
// The REQ-REP socket pair is in lockstep. Doing any other
// sequence (e.g., sending two messages in a row) will result in error.
package rep

import (
	"encoding/binary"
	"sync"
	"time"

	nano "github.com/funkygao/nano"
)

type repEp struct {
	q    chan *nano.Message
	ep   nano.Endpoint
	sock nano.ProtocolSocket
	w    nano.Waiter
	r    *rep
}

func (pe *repEp) sender() {
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

type rep struct {
	sock         nano.ProtocolSocket
	eps          map[uint32]*repEp
	backtracebuf []byte
	backtrace    []byte
	backtraceL   sync.Mutex
	raw          bool
	ttl          int
	waiter       nano.Waiter
	once         sync.Once

	sync.Mutex
}

func (r *rep) Init(sock nano.ProtocolSocket) {
	r.sock = sock
	r.eps = make(map[uint32]*repEp)
	r.backtracebuf = make([]byte, 64)
	r.ttl = 8 // default specified in the RFC
	r.waiter.Init()
	nano.Debugf("set send state: %v", nano.ErrProtoState)
	r.sock.SetSendError(nano.ErrProtoState)
}

func (r *rep) Shutdown(expire time.Time) {
	r.waiter.WaitAbsTimeout(expire)

	r.Lock()
	peers := r.eps
	r.eps = make(map[uint32]*repEp)
	r.Unlock()

	for id, peer := range peers {
		delete(peers, id)
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
	}
}

func (r *rep) receiver(ep nano.Endpoint) {
	recvChan := r.sock.RecvChannel()
	closeChan := r.sock.CloseChannel()

	var m *nano.Message
	for {
		m = ep.RecvMsg()
		if m == nil {
			return
		}

		v := ep.Id()
		m.Header = append(m.Header,
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))

		hops := 0
		// Move backtrace from body to header.
		for {
			if hops >= r.ttl {
				m.Free() // ErrTooManyHops
				return
			}
			hops++
			if len(m.Body) < 4 {
				m.Free() // ErrGarbled
				return
			}
			m.Header = append(m.Header, m.Body[:4]...)
			m.Body = m.Body[4:]
			// Check for high order bit set (0x80000000, big endian)
			if m.Header[len(m.Header)-4]&0x80 != 0 {
				break
			}
		}

		select {
		case recvChan <- m:
		case <-closeChan:
			m.Free()
			return
		}
	}
}

func (r *rep) sender() {
	defer r.waiter.Done()
	sendChan := r.sock.SendChannel()
	closeChan := r.sock.CloseChannel()

	var m *nano.Message
	for {
		select {
		case m = <-sendChan:
		case <-closeChan:
			return
		}

		// Lop off the 32-bit peer/pipe ID.  If absent, drop.
		if len(m.Header) < 4 {
			m.Free()
			continue
		}
		id := binary.BigEndian.Uint32(m.Header)
		m.Header = m.Header[4:]
		r.Lock()
		pe := r.eps[id]
		r.Unlock()
		if pe == nil {
			m.Free()
			continue
		}

		select {
		case pe.q <- m:
		default:
			// If our queue is full, we have no choice but to
			// throw it on the floor.  This shoudn't happen,
			// since each partner should be running synchronously.
			// Devices are a different situation, and this could
			// lead to lossy behavior there.  Initiators will
			// resend if this happens.  Devices need to have deep
			// enough queues and be fast enough to avoid this.
			m.Free()
		}
	}
}

func (*rep) Number() uint16 {
	return nano.ProtoRep
}

func (*rep) PeerNumber() uint16 {
	return nano.ProtoReq
}

func (*rep) Name() string {
	return "rep"
}

func (*rep) PeerName() string {
	return "req"
}

func (r *rep) AddEndpoint(ep nano.Endpoint) {
	pe := &repEp{ep: ep, r: r, q: make(chan *nano.Message, 2)} // TODO
	pe.w.Init()
	r.Lock()
	r.once.Do(func() {
		r.waiter.Add()
		go r.sender()
	})
	r.eps[ep.Id()] = pe
	r.Unlock()

	nano.Debugf("%#v, go receiver, sender...", ep)

	go r.receiver(ep)
	go pe.sender()
}

func (r *rep) RemoveEndpoint(ep nano.Endpoint) {
	r.Lock()
	delete(r.eps, ep.Id())
	r.Unlock()
	nano.Debugf("%#v", ep)
}

// We save the backtrace from this message.  This means that if the app calls
// Recv before calling Send, the saved backtrace will be lost.  This is how
// the application discards / cancels a request to which it declines to reply.
// This is only done in cooked mode.
func (r *rep) RecvHook(m *nano.Message) bool {
	if r.raw {
		return true
	}

	nano.Debugf("recv state normal, msg: %+v", *m)

	r.sock.SetSendError(nil)
	r.backtraceL.Lock()
	r.backtrace = append(r.backtracebuf[0:0], m.Header...)
	r.backtraceL.Unlock()
	m.Header = nil
	return true
}

func (r *rep) SendHook(m *nano.Message) bool {
	// Store our saved backtrace.  Note that if none was previously stored,
	// there is no one to reply to, and we drop the message.  We only
	// do this in cooked mode.
	if r.raw {
		return true
	}
	r.sock.SetSendError(nano.ErrProtoState)
	r.backtraceL.Lock()
	m.Header = append(m.Header[0:0], r.backtrace...)
	r.backtrace = nil
	r.backtraceL.Unlock()
	if m.Header == nil {
		return false
	}
	return true
}

func (r *rep) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if r.raw, ok = v.(bool); !ok {
			return nano.ErrBadValue
		}
		if r.raw {
			r.sock.SetSendError(nil)
		} else {
			r.sock.SetSendError(nano.ErrProtoState)
		}
		return nil
	case nano.OptionTtl:
		if ttl, ok := v.(int); !ok {
			return nano.ErrBadValue
		} else if ttl < 1 || ttl > 255 {
			return nano.ErrBadValue
		} else {
			r.ttl = ttl
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (r *rep) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return r.raw, nil
	case nano.OptionTtl:
		return r.ttl, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewProtocol returns a new REP protocol object.
func NewProtocol() nano.Protocol {
	return &rep{}
}

// NewSocket allocates a new Socket using the REP protocol.
func NewSocket() (nano.Socket, error) {
	return nano.MakeSocket(&rep{}), nil
}
