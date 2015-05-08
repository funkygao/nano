// Package ipc implements the IPC transport on top of UNIX domain sockets.
package ipc

import (
	"net"

	"github.com/funkygao/nano"
)

// options is used for shared GetOption/SetOption logic.
type options map[string]interface{}

// GetOption retrieves an option value.
func (o options) get(name string) (interface{}, error) {
	if o == nil {
		return nil, nano.ErrBadOption
	}
	if v, ok := o[name]; !ok {
		return nil, nano.ErrBadOption
	} else {
		return v, nil
	}
}

// SetOption sets an option.  We have none, so just ErrBadOption.
func (o options) set(string, interface{}) error {
	return nano.ErrBadOption
}

type dialer struct {
	t     *ipcTran
	addr  *net.UnixAddr
	proto nano.Protocol
	opts  options
}

// Dial implements the PipeDialer Dial method
func (d *dialer) Dial() (nano.Pipe, error) {
	conn, err := net.DialUnix("unix", nil, d.addr)
	if err != nil {
		return nil, err
	}

	props := make([]interface{}, 0)
	for n, v := range d.t.opts {
		props = append(props, n, v)
	}
	return nano.NewConnPipeIPC(conn, d.proto, props...)
}

// SetOption implements a stub PipeDialer SetOption method.
func (d *dialer) SetOption(n string, v interface{}) error {
	return d.opts.set(n, v)
}

// GetOption implements a stub PipeDialer GetOption method.
func (d *dialer) GetOption(n string) (interface{}, error) {
	return d.opts.get(n)
}

type listener struct {
	t        *ipcTran
	addr     *net.UnixAddr
	proto    nano.Protocol
	listener *net.UnixListener
	opts     options
}

// Listen implements the PipeListener Listen method.
func (l *listener) Listen() error {
	if listener, err := net.ListenUnix("unix", l.addr); err != nil {
		return err
	} else {
		l.listener = listener
	}
	return nil
}

// Accept implements the the PipeListener Accept method.
func (l *listener) Accept() (nano.Pipe, error) {
	conn, err := l.listener.AcceptUnix()
	if err != nil {
		return nil, err
	}

	props := make([]interface{}, 0)
	for n, v := range l.t.opts {
		props = append(props, n, v)
	}
	return nano.NewConnPipeIPC(conn, l.proto, props...)
}

// Close implements the PipeListener Close method.
func (l *listener) Close() error {
	l.listener.Close()
	return nil
}

// SetOption implements a stub PipeListener SetOption method.
func (l *listener) SetOption(n string, v interface{}) error {
	return l.opts.set(n, v)
}

// GetOption implements a stub PipeListener GetOption method.
func (l *listener) GetOption(n string) (interface{}, error) {
	return l.opts.get(n)
}

type ipcTran struct {
	opts map[string]interface{}
}

// Scheme implements the Transport Scheme method.
func (t *ipcTran) Scheme() string {
	return "ipc"
}

// NewDialer implements the Transport NewDialer method.
func (t *ipcTran) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
	var err error

	if addr, err = nano.StripScheme(t, addr); err != nil {
		return nil, err
	}

	d := &dialer{t: t, proto: proto, opts: nil}
	if d.addr, err = net.ResolveUnixAddr("unix", addr); err != nil {
		return nil, err
	}
	return d, nil
}

// NewListener implements the Transport NewListener method.
func (t *ipcTran) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
	var err error
	l := &listener{t: t, proto: proto}

	if addr, err = nano.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if l.addr, err = net.ResolveUnixAddr("unix", addr); err != nil {
		return nil, err
	}

	return l, nil
}

var validOpts = map[string]bool{
	nano.OptionDisableHandshake: true,
	nano.OptionDeflate:          true,
	nano.OptionSnappy:           true,
}

// NewTransport allocates a new IPC transport.
func NewTransport(opts ...interface{}) nano.Transport {
	t := &ipcTran{opts: make(map[string]interface{})}
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
