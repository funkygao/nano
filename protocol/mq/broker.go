package mq

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type broker struct {
	sock     nano.ProtocolSocket
	topicMap map[string]*Topic
	sync.RWMutex
}

func (this *broker) Init(sock nano.ProtocolSocket) {
	this.sock = sock
	this.initializeTopics()
}

func (this *broker) initializeTopics() {
	// TODO read from mysql and build topics
	this.topicMap = make(map[string]*Topic)
}

func (this *broker) AddEndpoint(ep nano.Endpoint) {
	go this.ioLoop(ep)
}

func (this *broker) ioLoop(ep nano.Endpoint) {
	// handshake for mq protocol version magic
	protocolMagic, err := this.handshake(ep)
	if err != nil {
		return
	}

	var prot Protocol
	switch protocolMagic {
	case 1:
		prot = &protocolV1{broker: this, ep: ep}

	default:
		nano.Debugf("invalid protocol")
		ep.Close()
		return
	}

	prot.IOLoop(ep)
}

func (this *broker) handshake(ep nano.Endpoint) (protocolMagic int, err error) {
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

func (this *broker) RemoveEndpoint(subscriber nano.Endpoint) {

}

func (this *broker) Shutdown(expire time.Time) {
	this.Lock()
	topics := make([]*Topic, 0)
	for _, t := range this.topicMap {
		topics = append(topics, t)
	}
	this.Unlock()

	for _, t := range topics {
		t.Close()
	}

}

func (this *broker) getTopic(topicName string) *Topic {
	this.Lock()
	if t, present := this.topicMap[topicName]; present {
		this.Unlock()
		return t
	}

	t := NewTopic(topicName, this)
	this.topicMap[topicName] = t
	this.Unlock()
	return t
}

func (this *broker) SetOption(name string, val interface{}) error {
	return nil
}

func (this *broker) GetOption(name string) (interface{}, error) {
	return nil, nil
}

func (*broker) Number() uint16 {
	return nano.ProtoMq
}

func (*broker) PeerNumber() uint16 {
	return nano.ProtoMq
}

func NewBrokerSocket() nano.Socket {
	return nano.MakeSocket(&broker{})
}
