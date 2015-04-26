package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type tcpTransport struct {
	localAddr net.Addr
}

func (this *tcpTransport) Scheme() string {
	return "tcp"
}

func (this *tcpTransport) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
	var err error
	if addr, err = nano.StripScheme(this, addr); err != nil {
		return nil, err
	}
	d := &dialer{proto: proto, opts: newOptions()}
	if d.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}

	nano.Debugf("dialer:%v", *d)
	return d, nil
}

func (this *tcpTransport) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
	var err error
	if addr, err = nano.StripScheme(this, addr); err != nil {
		return nil, err
	}
	l := &listener{proto: proto, opts: newOptions()}
	if l.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}

	nano.Debugf("listener:%#v", *l)
	return l, nil
}

// NewTransport allocates a new TCP Transport.
func NewTransport() nano.Transport {
	return &tcpTransport{}
}
