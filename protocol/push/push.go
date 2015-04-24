// Package push implements the PUSH protocol, which is the write side of
// the pipeline pattern.  (PULL is the reader.)
package push

import (
	"time"

	mangos "github.com/funkygao/nano"
)

type push struct {
	sock mangos.ProtocolSocket
	raw  bool
	w    mangos.Waiter
}

func (x *push) Init(sock mangos.ProtocolSocket) {
	x.sock = sock
	x.w.Init()
	x.sock.SetRecvError(mangos.ErrProtoOp)
}

func (x *push) Shutdown(expire time.Time) {
	x.w.WaitAbsTimeout(expire)
}

func (x *push) sender(ep mangos.Endpoint) {
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
	return mangos.ProtoPush
}

func (*push) PeerNumber() uint16 {
	return mangos.ProtoPull
}

func (*push) Name() string {
	return "push"
}

func (*push) PeerName() string {
	return "pull"
}

func (x *push) AddEndpoint(ep mangos.Endpoint) {
	x.w.Add()
	go x.sender(ep)
	go mangos.NullRecv(ep)
}

func (x *push) RemoveEndpoint(ep mangos.Endpoint) {}

func (x *push) SetOption(name string, v interface{}) error {
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

func (x *push) GetOption(name string) (interface{}, error) {
	switch name {
	case mangos.OptionRaw:
		return x.raw, nil
	default:
		return nil, mangos.ErrBadOption
	}
}

// NewProtocol returns a new PUSH protocol object.
func NewProtocol() mangos.Protocol {
	return &push{}
}

// NewSocket allocates a new Socket using the PUSH protocol.
func NewSocket() (mangos.Socket, error) {
	return mangos.MakeSocket(&push{}), nil
}
