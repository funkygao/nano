package mq

import (
	"github.com/funkygao/nano"
)

func handshake(broker nano.Endpoint) (err error) {
	msg := nano.NewMessage(2)
	msg.Body = msg.Body[:2]
	msg.Body[0] = 0
	msg.Body[1] = 1
	if err = broker.SendMsg(msg); err != nil {
		msg.Free()
		return
	}
	if err = broker.Flush(); err != nil {
		return
	}

	msg.Free()
	return
}
