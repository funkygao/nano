// Package pull implements the PULL protocol, which is the read side of
// the pipeline pattern.  (PUSH is the reader.)
package pull

import (
	"time"

	mangos "github.com/funkygao/nano"
)

type pull struct {
	sock mangos.ProtocolSocket
	raw  bool
}

func (x *pull) Init(sock mangos.ProtocolSocket) {
	x.sock = sock
	x.sock.SetSendError(mangos.ErrProtoOp)
}

func (x *pull) Shutdown(time.Time) {} // No sender to drain

func (x *pull) receiver(ep mangos.Endpoint) {
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
	return mangos.ProtoPull
}

func (*pull) PeerNumber() uint16 {
	return mangos.ProtoPush
}

func (*pull) Name() string {
	return "pull"
}

func (*pull) PeerName() string {
	return "push"
}

func (x *pull) AddEndpoint(ep mangos.Endpoint) {
	go x.receiver(ep)
}

func (x *pull) RemoveEndpoint(ep mangos.Endpoint) {}

func (*pull) SendHook(msg *mangos.Message) bool {
	return false
}

func (x *pull) SetOption(name string, v interface{}) error {
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

func (x *pull) GetOption(name string) (interface{}, error) {
	switch name {
	case mangos.OptionRaw:
		return x.raw, nil
	default:
		return nil, mangos.ErrBadOption
	}
}

// NewProtocol() allocates a new PULL protocol object.
func NewProtocol() mangos.Protocol {
	return &pull{}
}

// NewSocket allocates a new Socket using the PULL protocol.
func NewSocket() (mangos.Socket, error) {
	return mangos.MakeSocket(&pull{}), nil
}
