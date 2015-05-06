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
	return nano.NewConnPipeIPC(conn, d.proto)
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
	return nano.NewConnPipeIPC(conn, l.proto)
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

type ipcTran struct{}

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

	d := &dialer{proto: proto, opts: nil}
	if d.addr, err = net.ResolveUnixAddr("unix", addr); err != nil {
		return nil, err
	}
	return d, nil
}

// NewListener implements the Transport NewListener method.
func (t *ipcTran) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
	var err error
	l := &listener{proto: proto}

	if addr, err = nano.StripScheme(t, addr); err != nil {
		return nil, err
	}

	if l.addr, err = net.ResolveUnixAddr("unix", addr); err != nil {
		return nil, err
	}

	return l, nil
}

// NewTransport allocates a new IPC transport.
func NewTransport() nano.Transport {
	return &ipcTran{}
}
