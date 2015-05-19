package mq

import (
	"bytes"

	"github.com/funkygao/nano"
)

type protocolV1 struct {
	broker *broker
	ep     nano.Endpoint
}

func (this *protocolV1) IOLoop(ep nano.Endpoint) {
	nano.Debugf("endpoint: %d", ep.Id())

	this.ep = ep

	go this.sender(ep)

	//recvChan := this.ctx.mq.sock.RecvChannel()
	//closeChan := this.ctx.mq.sock.CloseChannel()

	var msg *nano.Message
	for {
		msg = ep.RecvMsg()
		if msg == nil {
			return
		}

		line := msg.ReadSlice('\n')
		// trim the '\n'
		line = line[:len(line)-1]
		params := bytes.Split(line, []byte{' '})
		this.execute(params, msg)

		// will not feed upper reader
		msg.Free()

		/*
			select {
			//case recvChan <- msg:
			case <-closeChan:
				return
			}*/
	}
}

func (this *protocolV1) execute(params [][]byte, m *nano.Message) {
	switch {
	case bytes.Equal(params[0], []byte("FIN")):
		this.FIN(params[1:], m)

	case bytes.Equal(params[0], []byte("SUB")):
		this.SUB(params[1:], m)

	case bytes.Equal(params[0], []byte("PUB")):
		this.PUB(params[1:], m)

	case bytes.Equal(params[0], []byte("BYE")):
		this.BYE(params[1:], m)

	case bytes.Equal(params[0], []byte("AUTH")):
		this.AUTH(params[1:], m)

	}

}

func (this *protocolV1) sender(ep nano.Endpoint) {
	sendChan := this.broker.sock.SendChannel()
	closeChan := this.broker.sock.CloseChannel()

	for {
		select {
		case <-closeChan:
			return

		case msg := <-sendChan:
			if err := ep.SendMsg(msg); err != nil {
				return
			}
			ep.Flush()
		}
	}
}

func (this *protocolV1) checkAuth() {

}

func (this *protocolV1) PUB(args [][]byte, m *nano.Message) {
	this.checkAuth()

	topicName := string(args[0])

	body := m.ReadFull()
	nano.Debugf("topic:%s body:%s", topicName, string(body))
	msg := nano.NewMessage(len(body))
	msg.Body = body
	t := this.broker.getTopic(topicName)
	t.PutMessage(msg)

	// send OK ack
}

func (this *protocolV1) FIN(args [][]byte, m *nano.Message) {

}

func (this *protocolV1) SUB(args [][]byte, m *nano.Message) {
	topicName := string(args[0])
	channelName := string(args[1])
	t := this.broker.getTopic(topicName)
	c := t.GetChannel(channelName)
	c.AddEndpoint(this.ep)
}

func (this *protocolV1) BYE(args [][]byte, m *nano.Message) {

}

func (this *protocolV1) AUTH(args [][]byte, m *nano.Message) {

}
