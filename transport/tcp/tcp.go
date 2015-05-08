package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

type tcpTransport struct {
	opts map[string]interface{}
}

func (this *tcpTransport) Scheme() string {
	return "tcp"
}

func (this *tcpTransport) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
	var err error
	if addr, err = nano.StripScheme(this, addr); err != nil {
		return nil, err
	}

	d := &dialer{
		t:     this,
		proto: proto,
		opts:  newOptions(),
	}
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

	l := &listener{
		t:     this,
		proto: proto,
		opts:  newOptions(),
	}
	if l.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}

	nano.Debugf("listener:%#v", *l)
	return l, nil
}

var validOpts = map[string]bool{
	nano.OptionDisableHandshake: true,
	nano.OptionDeflate:          true,
	nano.OptionSnappy:           true,
}

// NewTransport allocates a new TCP Transport.
func NewTransport(opts ...interface{}) nano.Transport {
	t := &tcpTransport{opts: make(map[string]interface{})}
	if len(opts)%2 != 0 {
		return nil
	}
	for i := 0; i+1 < len(opts); i += 2 {
		name := opts[i].(string)
		if _, present := validOpts[name]; !present {
			// invalid option
			return nil
		}

		t.opts[name] = opts[i+1]
	}

	return t
}
