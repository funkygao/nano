package pipeline

import (
	"time"

	"github.com/funkygao/nano"
)

type pull struct {
	sock nano.ProtocolSocket
	raw  bool
}

func (this *pull) Init(sock nano.ProtocolSocket) {
	this.sock = sock

	// recv only
	this.sock.SetSendError(nano.ErrProtoOp)
}

func (this *pull) AddEndpoint(ep nano.Endpoint) {
	go this.receiver(ep)
}

func (*pull) RemoveEndpoint(ep nano.Endpoint) {}

func (this *pull) receiver(ep nano.Endpoint) {
	recvChan := this.sock.RecvChannel()
	closeChan := this.sock.CloseChannel()
	for {
		msg := ep.RecvMsg()
		if msg == nil {
			return
		}

		select {
		case recvChan <- msg:
			// will block if app consumes slow

		case <-closeChan:
			return
		}
	}
}

func (*pull) Shutdown(time.Time) {} // No sender to drain

func (*pull) Number() uint16 {
	return nano.ProtoPull
}

func (*pull) PeerNumber() uint16 {
	return nano.ProtoPush
}

func (*pull) SendHook(msg *nano.Message) bool {
	return false
}

func (this *pull) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if this.raw, ok = v.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil

	default:
		return nano.ErrBadOption
	}
}

func (this *pull) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return this.raw, nil

	default:
		return nil, nano.ErrBadOption
	}
}

// NewPullSocket allocates a new Socket using the PULL protocol.
func NewPullSocket() nano.Socket {
	return nano.MakeSocket(&pull{})
}
