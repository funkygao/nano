package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type tcpTran struct {
	localAddr net.Addr
}

func (t *tcpTran) Scheme() string {
	return "tcp"
}

func (t *tcpTran) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
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

func (t *tcpTran) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
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

// NewTransport allocates a new TCP transport.
func NewTransport() nano.Transport {
	return &tcpTran{}
}
