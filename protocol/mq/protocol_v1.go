package mq

import (
	"bytes"

	"github.com/funkygao/nano"
)

type protocolV1 struct {
	ctx *context
	ep  nano.Endpoint
}

func (this *protocolV1) IOLoop(ep nano.Endpoint) {
	this.ep = ep

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

		msg.Free()
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

func (this *protocolV1) checkAuth() {

}

func (this *protocolV1) PUB(args [][]byte, m *nano.Message) {
	this.checkAuth()

	topicName := string(args[0])
	t := this.ctx.mq.getTopic(topicName)
	body := m.ReadFull()
	msg := nano.NewMessage(len(body))
	msg.Body = body
	t.PutMessage(msg)
}

func (this *protocolV1) FIN(args [][]byte, m *nano.Message) {

}

func (this *protocolV1) SUB(args [][]byte, m *nano.Message) {
	topicName := string(args[0])
	channelName := string(args[1])
	t := this.ctx.mq.getTopic(topicName)
	c := t.GetChannel(channelName)
	c.AddEndpoint(this.ep)
}

func (this *protocolV1) BYE(args [][]byte, m *nano.Message) {

}

func (this *protocolV1) AUTH(args [][]byte, m *nano.Message) {

}
