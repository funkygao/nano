package mq

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type pubEp struct {
	ep      nano.Endpoint
	msgChan chan *nano.Message
}

func (this *pubEp) sendToBroker() {
	var msg *nano.Message
	for msg = range this.msgChan {
		if msg == nil {
			break
		}

		if this.ep.SendMsg(msg) != nil || this.ep.Flush() != nil {
			msg.Free()
			break
		}

		msg.Free()
	}

}

func (this *pubEp) Close() {
	close(this.msgChan)
}

type pub struct {
	sock    nano.ProtocolSocket
	brokers map[nano.EndpointId]*pubEp
	once    sync.Once
	waiter  nano.Waiter
	sync.Mutex
}

func (this *pub) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.brokers = make(map[nano.EndpointId]*pubEp)
	this.waiter.Init()
}

func (this *pub) AddEndpoint(ep nano.Endpoint) {
	this.once.Do(func() {
		this.waiter.Add()
		go this.sender()
	})

	nano.Debugf("connected with %s", ep.RemoteAddr())
	if err := handshake(ep); err != nil {
		nano.Debugf(err.Error())
		return
	}

	qlen := 16
	if i, err := this.sock.GetOption(nano.OptionWriteQLen); err == nil {
		qlen = i.(int)
	}
	nano.Debugf("qlen: %d", qlen)
	this.Lock()
	b := &pubEp{
		ep:      ep,
		msgChan: make(chan *nano.Message, qlen),
	}
	this.brokers[ep.Id()] = b
	this.Unlock()

	this.waiter.Add()
	go b.sendToBroker()
	go nano.NullRecv(ep)
}

// top sender
func (this *pub) sender() {
	defer this.waiter.Done()

	sendChan := this.sock.SendChannel()
	closeChan := this.sock.CloseChannel()
	var msg *nano.Message
	for {
		select {
		case <-closeChan:
			return

		case msg = <-sendChan:
			this.Lock()
			for _, b := range this.brokers {
				m := msg.Dup()
				select {
				case b.msgChan <- m:

				default:
					// queue full, drop silently
					m.Free()
				}
			}
			this.Unlock()
			msg.Free()
		}
	}
}

func (this *pub) RemoveEndpoint(ep nano.Endpoint) {
	this.Lock()
	b := this.brokers[ep.Id()]
	close(b.msgChan)
	delete(this.brokers, ep.Id())
	this.Unlock()
}

func (this *pub) Shutdown(expire time.Time) {
	this.waiter.WaitAbsTimeout(expire)

	this.Lock()
	brokers := make([]*pubEp, 0)
	for _, b := range this.brokers {
		brokers = append(brokers, b)
	}
	this.Unlock()

	for _, b := range brokers {
		b.Close()
	}

}

func (this *pub) SetOption(name string, val interface{}) error {
	return nano.ErrBadOption
}

func (this *pub) GetOption(name string) (interface{}, error) {
	return nil, nano.ErrBadOption
}

func (*pub) Number() uint16 {
	return nano.ProtoMq
}

func (*pub) PeerNumber() uint16 {
	return nano.ProtoMq
}

func NewPubSocket() nano.Socket {
	return nano.MakeSocket(&pub{})
}

func Pub(sock nano.Socket, topic string) {
	sock.Send(nil)

}
