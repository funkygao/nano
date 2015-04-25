package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type listener struct {
	addr     *net.TCPAddr
	proto    nano.Protocol
	listener *net.TCPListener
	opts     options
}

func (l *listener) Accept() (nano.Pipe, error) {

	if l.listener == nil {
		return nil, nano.ErrClosed
	}
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if err = l.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return nano.NewConnPipe(conn, l.proto)
}

func (l *listener) Listen() (err error) {
	l.listener, err = net.ListenTCP("tcp", l.addr)
	return
}

func (l *listener) Close() error {
	l.listener.Close()
	return nil
}

func (l *listener) SetOption(n string, v interface{}) error {
	return l.opts.set(n, v)
}

func (l *listener) GetOption(n string) (interface{}, error) {
	return l.opts.get(n)
}
