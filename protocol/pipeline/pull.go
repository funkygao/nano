package pipeline

import (
	"time"

	"github.com/funkygao/nano"
)

type pull struct {
	sock nano.ProtocolSocket
	raw  bool
}

func (x *pull) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.sock.SetSendError(nano.ErrProtoOp)
}

func (x *pull) Shutdown(time.Time) {} // No sender to drain

func (x *pull) receiver(ep nano.Endpoint) {
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

func (*pull) Number() uint16 {
	return nano.ProtoPull
}

func (*pull) PeerNumber() uint16 {
	return nano.ProtoPush
}

func (*pull) Name() string {
	return "pull"
}

func (*pull) PeerName() string {
	return "push"
}

func (x *pull) AddEndpoint(ep nano.Endpoint) {
	go x.receiver(ep)
}

func (x *pull) RemoveEndpoint(ep nano.Endpoint) {}

func (*pull) SendHook(msg *nano.Message) bool {
	return false
}

func (x *pull) SetOption(name string, v interface{}) error {
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

func (x *pull) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewPullSocket allocates a new Socket using the PULL protocol.
func NewPullSocket() nano.Socket {
	return nano.MakeSocket(&pull{})
}
