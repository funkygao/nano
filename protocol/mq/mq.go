package mq

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type mq struct {
	sock nano.ProtocolSocket

	topicMap map[string]*Topic

	sync.RWMutex
}

func (this *mq) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.topicMap = make(map[string]*Topic)
}

func (this *mq) AddEndpoint(ep nano.Endpoint) {
	go this.ioLoop(ep)
	//go nano.NullRecv(ep)
}

func (this *mq) ioLoop(ep nano.Endpoint) {
	// handshake for mq protocol version magic
	protocolMagic, err := this.handshake(ep)
	if err != nil {
		return
	}

	nano.Debugf("handshake done")

	var prot Protocol
	switch protocolMagic {
	case 1:
		prot = &protocolV1{ctx: &context{this}, ep: ep}

	default:
		nano.Debugf("invalid protocol")
		ep.Close()
		return
	}

	prot.IOLoop(ep)
}

func (this *mq) handshake(ep nano.Endpoint) (protocolMagic int, err error) {
	msg := nano.NewMessage(2)
	msg.Body = msg.Body[:2]
	msg.Body[0] = 0
	msg.Body[1] = 1
	if err = ep.SendMsg(msg); err != nil {
		return
	}
	if err = ep.Flush(); err != nil {
		return
	}

	msg = ep.RecvMsg()
	if msg == nil {
		return -1, nano.ErrClosed
	}

	protocolMagic = int(msg.Body[1])
	msg.Free()
	return
}

func (this *mq) RemoveEndpoint(subscriber nano.Endpoint) {

}

func (this *mq) Shutdown(expire time.Time) {

}

func (this *mq) getTopic(topicName string) *Topic {
	this.Lock()
	if t, present := this.topicMap[topicName]; present {
		this.Unlock()
		return t
	}

	t := NewTopic(topicName, &context{mq: this})
	this.topicMap[topicName] = t
	this.Unlock()
	return t
}

func (this *mq) SetOption(name string, val interface{}) error {
	return nil
}

func (this *mq) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*mq) Number() uint16 {
	return nano.ProtoMq
}

func (*mq) PeerNumber() uint16 {
	return nano.ProtoMq
}

func NewSocket() nano.Socket {
	return nano.MakeSocket(&mq{})
}
