package reqrep

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

// req is an implementation of the req protocol.
type req struct {
	sock          nano.ProtocolSocket
	resendMsgChan chan *nano.Message
	raw           bool
	retry         time.Duration
	nextid        uint32
	reqid         uint32
	waker         *time.Timer
	waiter        nano.Waiter

	outstandingReq *nano.Message

	sync.Mutex
}

func (r *req) Init(socket nano.ProtocolSocket) {
	r.sock = socket
	r.resendMsgChan = make(chan *nano.Message) // TODO buffer

	r.nextid = uint32(time.Now().UnixNano()) // quasi-random
	r.retry = time.Minute * 1                // retry after a minute
	r.waker = time.NewTimer(r.retry)
	r.waker.Stop()

	r.sock.SetRecvError(nano.ErrProtoState)

	r.waiter.Init()
	r.waiter.Add()
	go r.resendMsgChaner()

	nano.Debugf("got initial nextid:%d, recv state:%v, this:%+v",
		r.nextid, nano.ErrProtoState, *r)
}

func (r *req) AddEndpoint(ep nano.Endpoint) {
	go r.receiver(ep)

	r.waiter.Add()
	go r.sender(ep)
}

func (*req) RemoveEndpoint(nano.Endpoint) {}

// nextID returns the next request ID.
func (r *req) nextID() uint32 {
	// The high order bit is "special", and must always be set.  (This is
	// how the peer will detect the end of the backtrace.)
	v := r.nextid | 0x80000000
	r.nextid++
	nano.Debugf("nextid: %d: %d", v, r.nextid)
	return v
}

// resendMsgChan sends the request message again, after a timer has expired.
func (r *req) resendMsgChaner() {
	defer r.waiter.Done()
	closeChan := r.sock.CloseChannel()

	for {
		select {
		case <-r.waker.C:
		case <-closeChan:
			return
		}

		r.Lock()
		m := r.outstandingReq
		if m == nil {
			r.Unlock()
			continue
		}
		m = m.Dup()
		r.Unlock()

		r.resendMsgChan <- m
		r.Lock()
		if r.retry > 0 {
			r.waker.Reset(r.retry)
		} else {
			r.waker.Stop()
		}
		r.Unlock()
	}
}

func (r *req) receiver(ep nano.Endpoint) {
	recvChan := r.sock.RecvChannel()
	closeChan := r.sock.CloseChannel()
	var m *nano.Message
	for {
		m = ep.RecvMsg()
		if m == nil {
			break
		}

		if len(m.Body) < 4 {
			m.Free()
			continue
		}
		m.Header = append(m.Header, m.Body[:4]...)
		m.Body = m.Body[4:]

		select {
		case recvChan <- m:
		case <-closeChan:
			m.Free()
			break
		}
	}
}

func (r *req) sender(ep nano.Endpoint) {
	// NB: Because this function is only called when an endpoint is
	// added, we can reasonably safely cache the channels -- they won't
	// be changing after this point.

	defer r.waiter.Done()
	sendChan := r.sock.SendChannel()
	closeChan := r.sock.CloseChannel()
	resendChan := r.resendMsgChan
	var m *nano.Message
	for {
		select {
		case m = <-resendChan:
		case m = <-sendChan:
		case <-closeChan:
			return
		}

		nano.Debugf("sending: %+v ep:%T", *m, ep)
		if ep.SendMsg(m) != nil || ep.Flush() != nil {
			r.resendMsgChan <- m
			break
		}

	}
}

func (r *req) Shutdown(expire time.Time) {
	nano.Debugf("expire:%v", expire)
	r.waiter.WaitAbsTimeout(expire)
}

func (*req) Number() uint16 {
	return nano.ProtoReq
}

func (*req) PeerNumber() uint16 {
	return nano.ProtoRep
}

func (*req) Name() string {
	return "req"
}

func (*req) PeerName() string {
	return "rep"
}

func (*req) Handshake() bool {
	return true
}

func (r *req) SendHook(m *nano.Message) bool {
	if r.raw {
		// Raw mode has no automatic retry, and must include the
		// request id in the header coming down.
		return true
	}

	r.Lock()

	// We need to generate a new request id, and append it to the header.
	r.reqid = r.nextID()
	m.Header = append(m.Header,
		byte(r.reqid>>24), byte(r.reqid>>16), byte(r.reqid>>8), byte(r.reqid))
	nano.Debugf("reqid:%d %v", r.reqid, *m)

	r.outstandingReq = m.Dup()

	// Schedule a retry, in case we don't get a reply.
	if r.retry > 0 {
		r.waker.Reset(r.retry)
	} else {
		r.waker.Stop()
	}

	r.sock.SetRecvError(nil)

	nano.Debugf("send state normal, msg header:%v, reqid:%v", m.Header,
		r.reqid)

	r.Unlock()
	return true
}

func (r *req) RecvHook(m *nano.Message) bool {
	if r.raw {
		// Raw mode just passes up messages unmolested.
		return true
	}

	nano.Debugf("%+v", *m)

	r.Lock()
	if len(m.Header) < 4 {
		return false
	}
	if r.outstandingReq == nil {
		return false
	}
	if binary.BigEndian.Uint32(m.Header) != r.reqid {
		r.Unlock()
		return false
	}
	r.waker.Stop()
	r.outstandingReq.Free()
	r.outstandingReq = nil

	r.sock.SetRecvError(nano.ErrProtoState)

	r.Unlock()
	return true
}

func (r *req) SetOption(option string, value interface{}) error {
	var ok bool
	switch option {
	case nano.OptionRaw:
		if r.raw, ok = value.(bool); !ok {
			return nano.ErrBadValue
		}
		if r.raw {
			r.sock.SetRecvError(nil)
		} else {
			nano.Debugf("set recv state: %v", nano.ErrProtoState)
			r.sock.SetRecvError(nano.ErrProtoState)
		}
		return nil
	case nano.OptionRetryTime:
		r.Lock()
		r.retry, ok = value.(time.Duration)
		r.Unlock()
		if !ok {
			return nano.ErrBadValue
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (r *req) GetOption(option string) (interface{}, error) {
	switch option {
	case nano.OptionRaw:
		return r.raw, nil
	case nano.OptionRetryTime:
		r.Lock()
		v := r.retry
		r.Unlock()
		return v, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewReqSocket allocates a new Socket using the REQ protocol.
func NewReqSocket() nano.Socket {
	return nano.MakeSocket(&req{})
}
