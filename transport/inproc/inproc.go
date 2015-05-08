// Package inproc implements an simple inproc transport for nano.
package inproc

import (
	"strings"
	"sync"

	"github.com/funkygao/nano"
)

// inproc implements the Pipe interface on top of channels.
type inproc struct {
	rq     chan *nano.Message
	wq     chan *nano.Message
	closeq chan struct{}
	readyq chan struct{}
	proto  nano.Protocol
	addr   addr
	peer   *inproc
}

type addr string

func (a addr) String() string {
	s := string(a)
	if strings.HasPrefix(s, "inproc://") {
		s = s[len("inproc://"):]
	}
	return s
}

func (addr) Network() string {
	return "inproc"
}

type listener struct {
	addr      string
	proto     nano.Protocol
	accepters []*inproc
}

type inprocTran struct{}

var listeners struct {
	// Who is listening, on which "address"?
	byAddr map[string]*listener
	cv     sync.Cond
	mx     sync.Mutex
}

func init() {
	listeners.byAddr = make(map[string]*listener)
	listeners.cv.L = &listeners.mx
}

func (p *inproc) RecvMsg() (*nano.Message, error) {
	if p.peer == nil {
		return nil, nano.ErrClosed
	}
	select {
	case m, ok := <-p.rq:
		if m == nil || !ok {
			return nil, nano.ErrClosed
		}
		// Upper protocols expect to have to pick header and
		// body part.  So mush them back together.
		//msg.Body = append(msg.Header, msg.Body...)
		//msg.Header = make([]byte, 0, 32)
		return m, nil
	case <-p.closeq:
		return nil, nano.ErrClosed
	}
}

func (p *inproc) Flush() error {
	return nil
}

func (p *inproc) SendMsg(m *nano.Message) error {
	if p.peer == nil {
		return nano.ErrClosed
	}

	// Upper protocols expect to have to pick header and body part.
	// Also we need to have a fresh copy of the message for receiver, to
	// break ownership.
	nmsg := nano.NewMessage(len(m.Header) + len(m.Body))
	nmsg.Body = append(nmsg.Body, m.Header...)
	nmsg.Body = append(nmsg.Body, m.Body...)
	select {
	case p.wq <- nmsg:
		return nil
	case <-p.closeq:
		nmsg.Free()
		return nano.ErrClosed
	}
}

func (p *inproc) LocalProtocol() uint16 {
	return p.proto.Number()
}

func (p *inproc) RemoteProtocol() uint16 {
	return p.proto.PeerNumber()
}

func (p *inproc) Close() error {
	close(p.closeq)
	return nil
}

func (p *inproc) IsOpen() bool {
	select {
	case <-p.closeq:
		return false
	default:
		return true
	}
}

func (p *inproc) GetProp(name string) (interface{}, error) {
	switch name {
	case nano.PropRemoteAddr:
		return p.addr, nil
	case nano.PropLocalAddr:
		return p.addr, nil
	}
	// We have no special properties
	return nil, nano.ErrBadProperty
}

type dialer struct {
	addr  string
	proto nano.Protocol
}

func (d *dialer) Dial() (nano.Pipe, error) {

	var server *inproc
	client := &inproc{proto: d.proto, addr: addr(d.addr)}
	client.readyq = make(chan struct{})
	client.closeq = make(chan struct{})

	listeners.mx.Lock()

	// NB: No timeouts here!
	for {
		var l *listener
		var ok bool
		if l, ok = listeners.byAddr[d.addr]; !ok || l == nil {
			listeners.mx.Unlock()
			return nil, nano.ErrConnRefused
		}

		if !nano.ValidPeers(client.proto, l.proto) {
			return nil, nano.ErrBadProto
		}

		if len(l.accepters) != 0 {
			server = l.accepters[len(l.accepters)-1]
			l.accepters = l.accepters[:len(l.accepters)-1]
			break
		}

		listeners.cv.Wait()
		continue
	}

	listeners.mx.Unlock()

	server.wq = make(chan *nano.Message)
	server.rq = make(chan *nano.Message)
	client.rq = server.wq
	client.wq = server.rq
	server.peer = client
	client.peer = server

	close(server.readyq)
	close(client.readyq)
	return client, nil
}

func (*dialer) SetOption(string, interface{}) error {
	return nano.ErrBadOption
}

func (*dialer) GetOption(string) (interface{}, error) {
	return nil, nano.ErrBadOption
}

func (l *listener) Listen() error {
	listeners.mx.Lock()
	if _, ok := listeners.byAddr[l.addr]; ok {
		listeners.mx.Unlock()
		return nano.ErrAddrInUse
	}
	listeners.byAddr[l.addr] = l
	listeners.cv.Broadcast()
	listeners.mx.Unlock()
	return nil
}

func (l *listener) Accept() (nano.Pipe, error) {
	server := &inproc{proto: l.proto, addr: addr(l.addr)}
	server.readyq = make(chan struct{})
	server.closeq = make(chan struct{})

	listeners.mx.Lock()
	l.accepters = append(l.accepters, server)
	listeners.cv.Broadcast()
	listeners.mx.Unlock()

	select {
	case <-server.readyq:
		return server, nil
	case <-server.closeq:
		return nil, nano.ErrClosed
	}
}

func (*listener) SetOption(string, interface{}) error {
	return nano.ErrBadOption
}

func (*listener) GetOption(string) (interface{}, error) {
	return nil, nano.ErrBadOption
}

func (l *listener) Close() error {
	listeners.mx.Lock()
	if listeners.byAddr[l.addr] == l {
		delete(listeners.byAddr, l.addr)
	}
	servers := l.accepters
	l.accepters = nil
	listeners.cv.Broadcast()
	listeners.mx.Unlock()

	for _, s := range servers {
		close(s.closeq)
	}

	return nil
}

func (t *inprocTran) Scheme() string {
	return "inproc"
}

func (t *inprocTran) NewDialer(addr string, proto nano.Protocol) (nano.PipeDialer, error) {
	if _, err := nano.StripScheme(t, addr); err != nil {
		return nil, err
	}
	return &dialer{addr: addr, proto: proto}, nil
}

func (t *inprocTran) NewListener(addr string, proto nano.Protocol) (nano.PipeListener, error) {
	if _, err := nano.StripScheme(t, addr); err != nil {
		return nil, err
	}
	l := &listener{addr: addr, proto: proto}
	return l, nil
}

// NewTransport allocates a new inproc:// transport.
func NewTransport() nano.Transport {
	return &inprocTran{}
}
