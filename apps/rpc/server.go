package rpc

import (
	"bytes"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/reqrep"
)

type server struct {
	sock nano.Socket
}

func (this *server) Register() {
}

func (this *server) handleCall(buf *bytes.Buffer) error {
	nano.BufferPoolPut(buf)
	return nil
}

func (this *server) Serv(addr string) error {
	var err error
	if err = this.sock.Listen(addr); err != nil {
		return err
	}

	for {
		buf := nano.BufferPoolGet()
		_, err = this.sock.XRecv(buf)
		if err != nil {
			return err
		}

		if err = this.handleCall(buf); err != nil {
			return err
		}

	}
}

func NewServer() *server {
	s := &server{
		sock: reqrep.NewRepSocket(),
	}
	return s
}
