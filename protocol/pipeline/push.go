package pipeline

import (
	"time"

	nano "github.com/funkygao/nano"
)

type push struct {
	sock nano.ProtocolSocket
	raw  bool
	w    nano.Waiter
}

func (this *push) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.w.Init()

	// send only
	this.sock.SetRecvError(nano.ErrProtoOp)
}

func (this *push) AddEndpoint(ep nano.Endpoint) {
	this.w.Add()
	go this.sender(ep)
	go nano.NullRecv(ep)
}

func (*push) RemoveEndpoint(ep nano.Endpoint) {}

func (this *push) sender(ep nano.Endpoint) {
	defer this.w.Done()
	sendChan := this.sock.SendChannel()
	closeChan := this.sock.CloseChannel()
	for {
		select {
		case <-closeChan:
			return

		case msg := <-sendChan:
			if err := ep.SendMsg(msg); err != nil {
				nano.Debugf("%v", err)
				msg.Free()
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

func (this *push) Shutdown(expire time.Time) {
	this.w.WaitAbsTimeout(expire)
}

func (this *push) SetOption(name string, v interface{}) error {
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

func (this *push) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return this.raw, nil

	default:
		return nil, nano.ErrBadOption
	}
}

// NewPushSocket allocates a new Socket using the PUSH protocol.
func NewPushSocket() nano.Socket {
	return nano.MakeSocket(&push{})
}
