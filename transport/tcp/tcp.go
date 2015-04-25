package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type tcpTransport struct {
	localAddr net.Addr
}

func (t *tcpTransport) Scheme() string {
	return "tcp"
}

func (t *tcpTransport) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
	var err error
	d := &dialer{proto: proto, opts: newOptions()}

	if addr, err = nano.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if d.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}
	return d, nil
}

func (t *tcpTransport) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
	var err error
	l := &listener{proto: proto, opts: newOptions()}

	if addr, err = nano.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if l.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}

	return l, nil
}

// NewtcpTransport allocates a new TCP tcpTransport.
func NewTransport() nano.Transport {
	return &tcpTransport{}
}
