// Package req implements the REQ protocol, which is the request side of
// the request/response pattern.  (REP is the reponse.)
package req

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

// req is an implementation of the req protocol.
type req struct {
	sync.Mutex
	sock          nano.ProtocolSocket
	eps           map[uint32]nano.Endpoint
	resendMsgChan chan *nano.Message
	raw           bool
	retry         time.Duration
	nextid        uint32
	waker         *time.Timer
	waiter        nano.Waiter
	once          sync.Once

	// fields describing the outstanding request
	reqmsg *nano.Message
	reqid  uint32
}

func (r *req) Init(socket nano.ProtocolSocket) {
	r.sock = socket
	r.eps = make(map[uint32]nano.Endpoint)
	r.resendMsgChan = make(chan *nano.Message) // TODO buffer
	r.waiter.Init()

	r.nextid = uint32(time.Now().UnixNano()) // quasi-random
	r.retry = time.Minute * 1                // retry after a minute
	r.waker = time.NewTimer(r.retry)
	r.waker.Stop()
	r.sock.SetRecvError(nano.ErrProtoState)

	nano.Debugf("got initial nextid:%d, recv state:%v, this:%+v",
		r.nextid, nano.ErrProtoState, *r)
}

func (r *req) Shutdown(expire time.Time) {
	nano.Debugf("expire:%v", expire)
	r.waiter.WaitAbsTimeout(expire)
}

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
		m := r.reqmsg
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

		nano.Debugf("sending: %+v", *m)
		if ep.SendMsg(m) != nil {
			r.resendMsgChan <- m
			break
		}
	}
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

func (r *req) AddEndpoint(ep nano.Endpoint) {
	r.once.Do(func() {
		r.waiter.Add()
		go r.resendMsgChaner()
	})

	r.Lock()
	r.eps[ep.Id()] = ep
	r.Unlock()
	go r.receiver(ep)
	r.waiter.Add()
	go r.sender(ep)

	nano.Debugf("invoke go: receiver,sender,resendMsgChaner eps:%v, ep:%#v",
		r.eps, ep)
}

func (*req) RemoveEndpoint(nano.Endpoint) {}

func (r *req) SendHook(m *nano.Message) bool {
	if r.raw {
		// Raw mode has no automatic retry, and must include the
		// request id in the header coming down.
		return true
	}
	r.Lock()
	defer r.Unlock()

	// We need to generate a new request id, and append it to the header.
	r.reqid = r.nextID()
	v := r.reqid
	m.Header = append(m.Header,
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))

	r.reqmsg = m.Dup()

	// Schedule a retry, in case we don't get a reply.
	if r.retry > 0 {
		r.waker.Reset(r.retry)
	} else {
		r.waker.Stop()
	}

	r.sock.SetRecvError(nil)

	nano.Debugf("send state normal, msg header:%v, reqid:%v", m.Header, v)

	return true
}

func (r *req) RecvHook(m *nano.Message) bool {
	if r.raw {
		// Raw mode just passes up messages unmolested.
		return true
	}
	r.Lock()
	defer r.Unlock()
	if len(m.Header) < 4 {
		return false
	}
	if r.reqmsg == nil {
		return false
	}
	if binary.BigEndian.Uint32(m.Header) != r.reqid {
		return false
	}
	r.waker.Stop()
	r.reqmsg.Free()
	r.reqmsg = nil
	r.sock.SetRecvError(nano.ErrProtoState)

	nano.Debugf("recv state:%v, msg header:%v, reqmsg:%v",
		nano.ErrProtoState, m.Header, *r.reqmsg)
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

// NewReq returns a new REQ protocol object.
func NewProtocol() nano.Protocol {
	return &req{}
}

// NewSocket allocates a new Socket using the REQ protocol.
func NewSocket() (nano.Socket, error) {
	return nano.MakeSocket(&req{}), nil
}
