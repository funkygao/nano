package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type dialer struct {
	addr  *net.TCPAddr
	proto nano.Protocol
	opts  options
}

func (d *dialer) Dial() (nano.Pipe, error) {
	conn, err := net.DialTCP("tcp", nil, d.addr)
	if err != nil {
		return nil, err
	}
	if err = d.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return nano.NewConnPipe(conn, d.proto)
}

func (d *dialer) SetOption(n string, v interface{}) error {
	return d.opts.set(n, v)
}

func (d *dialer) GetOption(n string) (interface{}, error) {
	return d.opts.get(n)
}
