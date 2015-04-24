// Package tcp implements the TCP transport for mangos.
package tcp

import (
	"net"

	mangos "github.com/funkygao/nano"
)

// options is used for shared GetOption/SetOption logic.
type options map[string]interface{}

// GetOption retrieves an option value.
func (o options) get(name string) (interface{}, error) {
	if v, ok := o[name]; !ok {
		return nil, mangos.ErrBadOption
	} else {
		return v, nil
	}
}

// SetOption sets an option.
func (o options) set(name string, val interface{}) error {
	switch name {
	case mangos.OptionNoDelay:
		fallthrough
	case mangos.OptionKeepAlive:
		switch v := val.(type) {
		case bool:
			o[name] = v
			return nil
		default:
			return mangos.ErrBadValue
		}
	}
	return mangos.ErrBadOption
}

func newOptions() options {
	o := make(map[string]interface{})
	o[mangos.OptionNoDelay] = true
	o[mangos.OptionKeepAlive] = true
	return options(o)
}

func (o options) configTCP(conn *net.TCPConn) error {
	if v, ok := o[mangos.OptionNoDelay]; ok {
		if err := conn.SetNoDelay(v.(bool)); err != nil {
			return err
		}
	}
	if v, ok := o[mangos.OptionKeepAlive]; ok {
		if err := conn.SetKeepAlive(v.(bool)); err != nil {
			return err
		}
	}
	return nil
}

type dialer struct {
	addr  *net.TCPAddr
	proto mangos.Protocol
	opts  options
}

func (d *dialer) Dial() (mangos.Pipe, error) {

	conn, err := net.DialTCP("tcp", nil, d.addr)
	if err != nil {
		return nil, err
	}
	if err = d.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return mangos.NewConnPipe(conn, d.proto)
}

func (d *dialer) SetOption(n string, v interface{}) error {
	return d.opts.set(n, v)
}

func (d *dialer) GetOption(n string) (interface{}, error) {
	return d.opts.get(n)
}

type listener struct {
	addr     *net.TCPAddr
	proto    mangos.Protocol
	listener *net.TCPListener
	opts     options
}

func (l *listener) Accept() (mangos.Pipe, error) {

	if l.listener == nil {
		return nil, mangos.ErrClosed
	}
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if err = l.opts.configTCP(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return mangos.NewConnPipe(conn, l.proto)
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

func (t *tcpTran) NewDialer(addr string, proto mangos.Protocol) (mangos.PipeDialer, error) {
	var err error
	d := &dialer{proto: proto, opts: newOptions()}

	if addr, err = mangos.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if d.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}
	return d, nil
}

func (t *tcpTran) NewListener(addr string, proto mangos.Protocol) (mangos.PipeListener, error) {
	var err error
	l := &listener{proto: proto, opts: newOptions()}

	if addr, err = mangos.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if l.addr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return nil, err
	}

	return l, nil
}

// NewTransport allocates a new TCP transport.
func NewTransport() mangos.Transport {
	return &tcpTran{}
}
