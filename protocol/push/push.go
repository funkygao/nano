// Package push implements the PUSH protocol, which is the write side of
// the pipeline pattern.  (PULL is the reader.)
package push

import (
	"time"

	nano "github.com/funkygao/nano"
)

type push struct {
	sock nano.ProtocolSocket
	raw  bool
	w    nano.Waiter
}

func (x *push) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.w.Init()
	x.sock.SetRecvError(nano.ErrProtoOp)
}

func (x *push) Shutdown(expire time.Time) {
	x.w.WaitAbsTimeout(expire)
}

func (x *push) sender(ep nano.Endpoint) {
	defer x.w.Done()
	sq := x.sock.SendChannel()
	cq := x.sock.CloseChannel()

	for {
		select {
		case <-cq:
			return
		case m := <-sq:
			if ep.SendMsg(m) != nil {
				m.Free()
				return
			}
		}
	}

}

func (*push) Number() uint16 {
	return nano.ProtoPush
}

func (*push) PeerNumber() uint16 {
	return nano.ProtoPull
}

func (*push) Name() string {
	return "push"
}

func (*push) PeerName() string {
	return "pull"
}

func (x *push) AddEndpoint(ep nano.Endpoint) {
	x.w.Add()
	go x.sender(ep)
	go nano.NullRecv(ep)
}

func (x *push) RemoveEndpoint(ep nano.Endpoint) {}

func (x *push) SetOption(name string, v interface{}) error {
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

func (x *push) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewProtocol returns a new PUSH protocol object.
func NewProtocol() nano.Protocol {
	return &push{}
}

// NewSocket allocates a new Socket using the PUSH protocol.
func NewSocket() (nano.Socket, error) {
	return nano.MakeSocket(&push{}), nil
}
