package mq

import (
	"github.com/funkygao/nano"
)

func handshake(ep nano.Endpoint) (err error) {
	msg := nano.NewMessage(2)
	msg.Body = msg.Body[:2]
	msg.Body[0] = 0
	msg.Body[1] = 1
	if err = ep.SendMsg(msg); err != nil {
		msg.Free()
		return
	}
	if err = ep.Flush(); err != nil {
		return
	}

	msg.Free()
	return
}
