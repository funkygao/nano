package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

// listener implements the nano.PipeListener interface.
type listener struct {
	addr     *net.TCPAddr
	proto    nano.Protocol
	listener *net.TCPListener
	opts     options
}

func (this *listener) Accept() (nano.Pipe, error) {
	if this.listener == nil {
		return nil, nano.ErrClosed
	}

	nano.Debugf("accepting...")
	conn, err := this.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	nano.Debugf("accepted tcp:%#s", conn.RemoteAddr().String())
	if err = this.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return nano.NewConnPipe(conn, this.proto)
}

func (this *listener) Listen() (err error) {
	nano.Debugf("%s", this.addr)
	this.listener, err = net.ListenTCP("tcp", this.addr)
	return
}

func (this *listener) Close() error {
	this.listener.Close()
	return nil
}

func (this *listener) SetOption(name string, val interface{}) error {
	return this.opts.set(name, val)
}

func (this *listener) GetOption(name string) (interface{}, error) {
	return this.opts.get(name)
}
