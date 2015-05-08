package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

// dialer implements the nano.PipeDialer interface.
type dialer struct {
	t     *tcpTransport
	addr  *net.TCPAddr
	proto nano.Protocol
	opts  options
}

func (this *dialer) Dial() (nano.Pipe, error) {
	conn, err := net.DialTCP("tcp", nil, this.addr)
	if err != nil {
		return nil, err
	}

	if err = this.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}

	nano.Debugf("dial tcp:%v done, NewConnPipe...", *this.addr)

	props := make([]interface{}, 0)
	for n, v := range this.t.opts {
		props = append(props, n, v)
	}
	return nano.NewConnPipe(conn, this.proto, props...)
}

func (this *dialer) SetOption(name string, val interface{}) error {
	return this.opts.set(name, val)
}

func (this *dialer) GetOption(name string) (interface{}, error) {
	return this.opts.get(name)
}
