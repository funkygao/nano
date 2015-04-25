// Package tcp implements the TCP transport for nano.
package tcp

import (
	"net"

	"github.com/funkygao/nano"
)

// options is used for shared GetOption/SetOption logic.
type options map[string]interface{}

// GetOption retrieves an option value.
func (o options) get(name string) (interface{}, error) {
	if v, ok := o[name]; !ok {
		return nil, nano.ErrBadOption
	} else {
		return v, nil
	}
}

// SetOption sets an option.
func (o options) set(name string, val interface{}) error {
	switch name {
	case nano.OptionNoDelay:
		fallthrough
	case nano.OptionKeepAlive:
		switch v := val.(type) {
		case bool:
			o[name] = v
			return nil
		default:
			return nano.ErrBadValue
		}
	}
	return nano.ErrBadOption
}

func newOptions() options {
	o := make(map[string]interface{})
	o[nano.OptionNoDelay] = true
	o[nano.OptionKeepAlive] = true
	return options(o)
}

func (o options) configTCP(conn *net.TCPConn) error {
	if v, ok := o[nano.OptionNoDelay]; ok {
		if err := conn.SetNoDelay(v.(bool)); err != nil {
			return err
		}
	}
	if v, ok := o[nano.OptionKeepAlive]; ok {
		if err := conn.SetKeepAlive(v.(bool)); err != nil {
			return err
		}
	}
	return nil
}

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
