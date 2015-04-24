package nano

import (
	"strings"
	"sync"
	"time"
)

// socket is the meaty part of the core information.
type socket struct {
	proto Protocol

	sync.Mutex

	sendChan     chan *Message
	sendChanSize int
	recvChan     chan *Message
	recvChanSize int
	closeChan    chan struct{} // closed when user requests close

	closing bool  // true if Socket was closed at API level
	active  bool  // true if either Dial or Listen has been successfully called
	recverr error // error to return on attempts to Recv()
	senderr error // error to return on attempts to Send()

	rdeadline  time.Duration
	wdeadline  time.Duration
	reconntime time.Duration // reconnect time after error or disconnect
	reconnmax  time.Duration // max reconnect interval before give up? TODO
	linger     time.Duration

	pipes []*pipe

	listeners []*listener

	transports map[string]Transport

	// These are conditional "type aliases" for our self
	sendhook ProtocolSendHook
	recvhook ProtocolRecvHook

	// Port hook -- called when a port is added or removed
	porthook PortHook
}

func (sock *socket) addPipe(tranpipe Pipe, d *dialer, l *listener) *pipe {
	p := newPipe(tranpipe)
	p.d = d
	p.l = l

	// Either listener or dialer is non-nil -- this could be an assert
	if l == nil && p == nil {
		p.Close()
		return nil
	}

	sock.Lock()
	if fn := sock.porthook; fn != nil {
		sock.Unlock()
		if !fn(PortActionAdd, p) {
			p.Close()
			return nil
		}
		sock.Lock()
	}
	p.sock = sock
	p.index = len(sock.pipes)
	sock.pipes = append(sock.pipes, p)
	sock.Unlock()
	sock.proto.AddEndpoint(p)
	return p
}

func (sock *socket) remPipe(p *pipe) {

	sock.proto.RemoveEndpoint(p)

	sock.Lock()
	if p.index >= 0 {
		sock.pipes[p.index] = sock.pipes[len(sock.pipes)-1]
		sock.pipes[p.index].index = p.index
		sock.pipes = sock.pipes[:len(sock.pipes)-1]
		p.index = -1
	}
	sock.Unlock()
}

func newSocket(proto Protocol) *socket {
	sock := new(socket)
	sock.sendChanSize = defaultQLen
	sock.recvChanSize = defaultQLen
	sock.sendChan = make(chan *Message, sock.sendChanSize)
	sock.recvChan = make(chan *Message, sock.recvChanSize)
	sock.closeChan = make(chan struct{})
	sock.reconntime = time.Millisecond * 100 // make it a tunable?
	sock.reconnmax = time.Minute             // TODO
	sock.linger = time.Second                // TODO
	sock.proto = proto
	sock.transports = make(map[string]Transport)

	// Add some conditionals now -- saves checks later
	if i, ok := interface{}(proto).(ProtocolRecvHook); ok {
		sock.recvhook = i.(ProtocolRecvHook)
	}
	if i, ok := interface{}(proto).(ProtocolSendHook); ok {
		sock.sendhook = i.(ProtocolSendHook)
	}

	proto.Init(sock)

	return sock
}

// MakeSocket is intended for use by Protocol implementations.  The intention
// is that they can wrap this to provide a "proto.NewSocket()" implementation.
func MakeSocket(proto Protocol) *socket {
	return newSocket(proto)
}

// Implementation of ProtocolSocket bits on socket.  This is the middle
// API presented to Protocol implementations.

func (sock *socket) SendChannel() <-chan *Message {
	return sock.sendChan
}

func (sock *socket) RecvChannel() chan<- *Message {
	return sock.recvChan
}

func (sock *socket) CloseChannel() <-chan struct{} {
	return sock.closeChan
}

func (sock *socket) SetSendError(err error) {
	sock.Lock()
	sock.senderr = err
	sock.Unlock()
}

func (sock *socket) SetRecvError(err error) {
	sock.Lock()
	sock.recverr = err
	sock.Unlock()
}

//
// Implementation of Socket bits on socket.  This is the upper API
// presented to applications.
//

func (sock *socket) Close() error {
	fin := time.Now().Add(sock.linger)

	DrainChannel(sock.sendChan, fin)

	sock.Lock()
	if sock.closing {
		sock.Unlock()
		return ErrClosed
	}
	sock.closing = true
	close(sock.closeChan)

	for _, l := range sock.listeners {
		l.l.Close()
	}
	pipes := append([]*pipe{}, sock.pipes...)
	sock.Unlock()

	// A second drain, just to be sure.  (We could have had device or
	// forwarded messages arrive since the last one.)
	DrainChannel(sock.sendChan, fin)

	// And tell the protocol to shutdown and drain its pipes too.
	sock.proto.Shutdown(fin)

	for _, p := range pipes {
		p.Close()
	}

	return nil
}

func (sock *socket) SendMsg(msg *Message) error {
	sock.Lock()
	e := sock.senderr
	if e != nil {
		sock.Unlock()
		return e
	}
	sock.Unlock()
	if sock.sendhook != nil {
		if ok := sock.sendhook.SendHook(msg); !ok {
			// just drop it silently
			msg.Free()
			return nil
		}
	}
	sock.Lock()
	timeout := mkTimer(sock.wdeadline)
	sock.Unlock()
	select {
	case <-timeout:
		return ErrSendTimeout
	case <-sock.closeChan:
		return ErrClosed
	case sock.sendChan <- msg:
		return nil
	}
}

func (sock *socket) Send(b []byte) error {
	msg := &Message{Body: b, Header: nil, refcnt: 1}
	return sock.SendMsg(msg)
}

func (sock *socket) RecvMsg() (*Message, error) {
	sock.Lock()
	timeout := mkTimer(sock.rdeadline)
	e := sock.recverr
	sock.Unlock()

	if e != nil {
		return nil, e
	}

	for {
		select {
		case <-timeout:
			return nil, ErrRecvTimeout
		case msg := <-sock.recvChan:
			if sock.recvhook != nil {
				if ok := sock.recvhook.RecvHook(msg); ok {
					return msg, nil
				} // else loop
				msg.Free()
			} else {
				return msg, nil
			}
		case <-sock.closeChan:
			return nil, ErrClosed
		}
	}
}

func (sock *socket) Recv() ([]byte, error) {
	msg, err := sock.RecvMsg()
	if err != nil {
		return nil, err
	}
	return msg.Body, nil
}

func (sock *socket) getTransport(addr string) Transport {
	var i int

	sock.Lock()
	defer sock.Unlock()

	if i = strings.Index(addr, "://"); i < 0 {
		return nil
	}
	scheme := addr[:i]
	t, ok := sock.transports[scheme]
	if t != nil && ok {
		return t
	}
	return nil
}

func (sock *socket) AddTransport(t Transport) {
	sock.Lock()
	sock.transports[t.Scheme()] = t
	sock.Unlock()
}

func (sock *socket) DialOptions(addr string, opts map[string]interface{}) error {

	d, err := sock.NewDialer(addr, opts)
	if err != nil {
		return err
	}
	return d.Dial()
}

func (sock *socket) Dial(addr string) error {
	return sock.DialOptions(addr, nil)
}

func (sock *socket) NewDialer(addr string, options map[string]interface{}) (Dialer, error) {
	var err error
	d := &dialer{sock: sock, addr: addr, closeChan: make(chan struct{})}
	t := sock.getTransport(addr)
	if t == nil {
		return nil, ErrBadTran
	}
	if d.d, err = t.NewDialer(addr, sock.proto); err != nil {
		return nil, err
	}
	for n, v := range options {
		if err = d.d.SetOption(n, v); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (sock *socket) ListenOptions(addr string, options map[string]interface{}) error {
	l, err := sock.NewListener(addr, options)
	if err != nil {
		return err
	}
	if err = l.Listen(); err != nil {
		return err
	}
	return nil
}

func (sock *socket) Listen(addr string) error {
	return sock.ListenOptions(addr, nil)
}

func (sock *socket) NewListener(addr string, options map[string]interface{}) (Listener, error) {
	// This function sets up a goroutine to accept inbound connections.
	// The accepted connection will be added to a list of accepted
	// connections.  The Listener just needs to listen continuously,
	// as we assume that we want to continue to receive inbound
	// connections without limit.
	t := sock.getTransport(addr)
	if t == nil {
		return nil, ErrBadTran
	}
	var err error
	l := &listener{sock: sock, addr: addr}
	l.l, err = t.NewListener(addr, sock.proto)
	if err != nil {
		return nil, err
	}
	for n, v := range options {
		if err = l.l.SetOption(n, v); err != nil {
			l.l.Close()
			return nil, err
		}
	}
	return l, nil
}

func (sock *socket) SetOption(name string, value interface{}) error {
	matched := false
	err := sock.proto.SetOption(name, value)
	if err == nil {
		matched = true
	} else if err != ErrBadOption {
		return err
	}
	switch name {
	case OptionRecvDeadline:
		sock.Lock()
		sock.rdeadline = value.(time.Duration)
		sock.Unlock()
		return nil
	case OptionSendDeadline:
		sock.Lock()
		sock.wdeadline = value.(time.Duration)
		sock.Unlock()
		return nil
	case OptionLinger:
		sock.Lock()
		sock.linger = value.(time.Duration)
		sock.Unlock()
		return nil
	case OptionWriteQLen:
		sock.Lock()
		defer sock.Unlock()
		if sock.active {
			return ErrBadOption
		}
		length := value.(int)
		if length < 0 {
			return ErrBadValue
		}
		sock.sendChanSize = length
		sock.sendChan = make(chan *Message, sock.sendChanSize)
		return nil
	case OptionReadQLen:
		sock.Lock()
		defer sock.Unlock()
		if sock.active {
			return ErrBadOption
		}
		length := value.(int)
		if length < 0 {
			return ErrBadValue
		}
		sock.recvChanSize = length
		sock.recvChan = make(chan *Message, sock.recvChanSize)
		return nil
	}
	if matched {
		return nil
	}
	return ErrBadOption
}

func (sock *socket) GetOption(name string) (interface{}, error) {
	val, err := sock.proto.GetOption(name)
	if err == nil {
		return val, nil
	}
	if err != ErrBadOption {
		return nil, err
	}

	switch name {
	case OptionRecvDeadline:
		sock.Lock()
		defer sock.Unlock()
		return sock.rdeadline, nil
	case OptionSendDeadline:
		sock.Lock()
		defer sock.Unlock()
		return sock.wdeadline, nil
	case OptionLinger:
		sock.Lock()
		defer sock.Unlock()
		return sock.linger, nil
	case OptionWriteQLen:
		sock.Lock()
		defer sock.Unlock()
		return sock.sendChanSize, nil
	case OptionReadQLen:
		sock.Lock()
		defer sock.Unlock()
		return sock.recvChanSize, nil
	}
	return nil, ErrBadOption
}

func (sock *socket) GetProtocol() Protocol {
	return sock.proto
}

func (sock *socket) SetPortHook(newhook PortHook) PortHook {
	sock.Lock()
	oldhook := sock.porthook
	sock.porthook = newhook
	sock.Unlock()
	return oldhook
}

type dialer struct {
	d         PipeDialer
	sock      *socket
	addr      string
	closed    bool
	active    bool
	closeChan chan struct{}
}

func (d *dialer) Dial() error {
	d.sock.Lock()
	if d.active {
		d.sock.Unlock()
		return ErrAddrInUse
	}
	d.closeChan = make(chan struct{})
	d.sock.active = true
	d.active = true
	d.sock.Unlock()
	go d.dialer()
	return nil
}

func (d *dialer) Close() error {
	d.sock.Lock()
	if d.closed {
		d.sock.Unlock()
		return ErrClosed
	}
	d.closed = true
	close(d.closeChan)
	d.sock.Unlock()
	return nil
}

func (d *dialer) GetOption(n string) (interface{}, error) {
	return d.d.GetOption(n)
}

func (d *dialer) SetOption(n string, v interface{}) error {
	return d.d.SetOption(n, v)
}

func (d *dialer) Address() string {
	return d.addr
}

// dialer is used to dial or redial from a goroutine.
func (d *dialer) dialer() {
	rtime := d.sock.reconntime
	rtmax := d.sock.reconnmax
	for {
		p, err := d.d.Dial()
		if err == nil {
			// reset retry time
			rtime = d.sock.reconntime
			d.sock.Lock()
			if d.closed {
				p.Close()
				return
			}
			d.sock.Unlock()
			if cp := d.sock.addPipe(p, d, nil); cp != nil {
				select {
				case <-d.sock.closeChan: // parent socket closed
				case <-cp.closeq: // disconnect event
				case <-d.closeChan: // dialer closed
				}
			}
		}

		// we're redialing here
		select {
		case <-d.closeChan: // dialer closed
			return
		case <-d.sock.closeChan: // exit if parent socket closed
			return
		case <-time.After(rtime):
			rtime *= 2
			if rtime > rtmax {
				rtime = rtmax
			}
			continue
		}
	}
}

type listener struct {
	l    PipeListener
	sock *socket
	addr string
}

func (l *listener) GetOption(n string) (interface{}, error) {
	return l.l.GetOption(n)
}

func (l *listener) SetOption(n string, v interface{}) error {
	return l.l.SetOption(n, v)
}

// serve spins in a loop, calling the accepter's Accept routine.
func (l *listener) serve() {
	for {
		select {
		case <-l.sock.closeChan:
			return
		default:
		}

		// If the underlying PipeListener is closed, or not
		// listening, we expect to return back with an error.
		if pipe, err := l.l.Accept(); err == nil {
			l.sock.addPipe(pipe, nil, l)
		} else if err == ErrClosed {
			return
		}
	}
}

func (l *listener) Listen() error {
	// This function sets up a goroutine to accept inbound connections.
	// The accepted connection will be added to a list of accepted
	// connections.  The Listener just needs to listen continuously,
	// as we assume that we want to continue to receive inbound
	// connections without limit.

	if err := l.l.Listen(); err != nil {
		return err
	}
	l.sock.listeners = append(l.sock.listeners, l)
	l.sock.Lock()
	l.sock.active = true
	l.sock.Unlock()
	go l.serve()
	return nil
}

func (l *listener) Address() string {
	return l.addr
}

func (l *listener) Close() error {
	return l.l.Close()
}
