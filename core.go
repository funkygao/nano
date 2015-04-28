package nano

import (
	"strings"
	"sync"
	"time"
)

// socket is the meaty part of the core information.
type socket struct {
	proto      Protocol
	transports map[string]Transport // key is transport scheme

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

	rdeadline  time.Duration // read deadline
	wdeadline  time.Duration // write deadline
	redialTime time.Duration // reconnect time after error or disconnect
	redialMax  time.Duration // max reconnect interval before give up? TODO SetOption
	linger     time.Duration // wait up to that time for sockets to drain

	pipes []*pipeEndpoint

	listeners []*listener // TODO discard this, pipes already got it

	sendhook ProtocolSendHook
	recvhook ProtocolRecvHook

	porthook PortHook
}

// MakeSocket is intended for use by Protocol implementations.  The intention
// is that they can wrap this to provide a "proto.NewSocket()" implementation.
func MakeSocket(proto Protocol) *socket {
	sock := &socket{
		proto:        proto,
		sendChanSize: defaultChanLen,
		recvChanSize: defaultChanLen,
		sendChan:     make(chan *Message, defaultChanLen),
		recvChan:     make(chan *Message, defaultChanLen),
		closeChan:    make(chan struct{}),
		redialTime:   defaultRedialTime,
		redialMax:    defaultRedialMax,
		linger:       defaultLingerTime,
		transports:   make(map[string]Transport, 1),
		pipes:        make([]*pipeEndpoint, 0, 2),
		listeners:    make([]*listener, 0), // TODO why no dialers?
	}

	// Add some conditionals now -- saves checks later
	if hook, ok := proto.(ProtocolRecvHook); ok {
		sock.recvhook = hook
		Debugf("got recvhook")
	}
	if hook, ok := proto.(ProtocolSendHook); ok {
		sock.sendhook = hook
		Debugf("got sendhook")
	}

	Debugf("sock:%+v, proto.Init(sock)...", *sock)

	proto.Init(sock)

	return sock
}

func (sock *socket) addPipe(tranpipe Pipe, d *dialer, l *listener) *pipeEndpoint {
	p := newPipe(tranpipe)
	p.dialer = d
	p.listener = l

	// dialer or listener must be provied only one
	if l == nil && d == nil {
		p.Close()
		return nil
	}
	if l != nil && d != nil {
		p.Close()
		return nil
	}

	Debugf("t:%#v, d:%#v, l:%#v", tranpipe, d, l)

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

func (sock *socket) removePipe(p *pipeEndpoint) {
	sock.proto.RemoveEndpoint(p)

	Debugf("%#v", *p)

	sock.Lock()
	if p.index >= 0 {
		sock.pipes[p.index] = sock.pipes[len(sock.pipes)-1]
		sock.pipes[p.index].index = p.index
		sock.pipes = sock.pipes[:len(sock.pipes)-1]
		p.index = -1
	}
	sock.Unlock()
}

func (sock *socket) SendChannel() <-chan *Message {
	return sock.sendChan
}

// RecvChannel is used for Protocol implementations.  The standard data
// flow is: Protocol.AddEndpoint will register the Endpoint(low level
// data stream implementation), Protocol will use this Endpoint to
// RecvMsg and put the message back to its socket RecvChannel.
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
	pipes := append([]*pipeEndpoint{}, sock.pipes...)
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
	e := sock.senderr // copy err
	sock.Unlock()
	if e != nil {
		return e
	}

	if sock.sendhook != nil {
		Debugf("SendHook: %+v", *msg)
		if ok := sock.sendhook.SendHook(msg); !ok {
			// just drop it silently TODO ErrSendHook?
			msg.Free()
			return nil
		}
	}

	sock.Lock() // TODO need lock?
	timeout := mkTimer(sock.wdeadline)
	sock.Unlock()
	select {
	case <-timeout:
		return ErrSendTimeout
	case <-sock.closeChan:
		return ErrClosed
	case sock.sendChan <- msg:
		Debugf("put to send chan: %+v", *msg)
		return nil
	}
}

func (sock *socket) Send(b []byte) error {
	// TODO borrow from message pool
	msg := &Message{Body: b, Header: nil, refCount: 1}
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

	var msg *Message
	for {
		select {
		case <-timeout:
			return nil, ErrRecvTimeout
		case msg = <-sock.recvChan:
			if sock.recvhook != nil {
				Debugf("RecvHook: %+v", *msg)
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
	sock.Lock()
	defer sock.Unlock()

	var i int
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
	Debugf("transports: %v", sock.transports)
	sock.Unlock()
}

func (sock *socket) DialOptions(addr string, options map[string]interface{}) error {
	d, err := sock.NewDialer(addr, options)
	if err != nil {
		return err
	}

	Debugf("addr:%s, opt:%v, dialing...", addr, options)

	return d.Dial()
}

func (sock *socket) Dial(addr string) error {
	return sock.DialOptions(addr, nil)
}

func (sock *socket) NewDialer(addr string, options map[string]interface{}) (Dialer, error) {
	d := &dialer{sock: sock, addr: addr, closeChan: make(chan struct{})}
	t := sock.getTransport(addr)
	if t == nil {
		return nil, ErrBadTran
	}

	var err error
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

	Debugf("addr:%s, opt:%v, listen...", addr, options)

	if err = l.Listen(); err != nil {
		return err
	}
	return nil
}

func (sock *socket) Listen(addr string) error {
	return sock.ListenOptions(addr, nil)
}

// NewListener wraps  PipeListener created in Transport.
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

// dialer implements the Dailer interface.
type dialer struct {
	d         PipeDialer
	sock      *socket
	addr      string
	closed    bool
	active    bool // TDOO dup with socket.active?
	closeChan chan struct{}
}

func (this *dialer) Dial() error {
	this.sock.Lock()
	if this.active {
		this.sock.Unlock()
		return ErrAddrInUse
	}

	this.closeChan = make(chan struct{})
	this.sock.active = true
	this.active = true
	this.sock.Unlock()

	Debugf("sock is active, go dialer...")

	go this.dialer()
	return nil
}

func (this *dialer) Close() error {
	this.sock.Lock()
	if this.closed {
		this.sock.Unlock()
		return ErrClosed
	}

	this.closed = true
	close(this.closeChan)
	this.sock.Unlock()
	return nil
}

func (this *dialer) GetOption(name string) (interface{}, error) {
	return this.d.GetOption(name)
}

func (this *dialer) SetOption(name string, val interface{}) error {
	return this.d.SetOption(name, val)
}

func (this *dialer) Address() string {
	return this.addr
}

// dialer is used to dial or redial from a goroutine.
func (this *dialer) dialer() {
	rtime := this.sock.redialTime
	rtmax := this.sock.redialMax
	for {
		pipe, err := this.d.Dial()
		if err == nil {
			// reset retry time
			rtime = this.sock.redialTime
			this.sock.Lock()
			if this.closed {
				// TODO this.sock.Unlock() ?
				pipe.Close()
				return
			}

			this.sock.Unlock()
			if cp := this.sock.addPipe(pipe, this, nil); cp != nil {
				select {
				case <-this.sock.closeChan: // parent socket closed
				case <-cp.closeChan: // disconnect event
				case <-this.closeChan: // dialer closed
				}
			}
		}

		// we're redialing here
		select {
		case <-this.closeChan: // dialer closed
			return
		case <-this.sock.closeChan: // exit if parent socket closed
			return
		case <-time.After(rtime):
			rtime *= 2
			Debugf("%s", rtime)
			if rtime > rtmax {
				rtime = rtmax
			}
			continue
		}
	}
}

// listener implements the Listener interface.
type listener struct {
	l    PipeListener
	sock *socket
	addr string
}

func (this *listener) GetOption(name string) (interface{}, error) {
	return this.l.GetOption(name)
}

func (this *listener) SetOption(name string, val interface{}) error {
	return this.l.SetOption(name, val)
}

// serve spins in a loop, calling the accepter's Accept routine.
func (l *listener) serve() {
	Debugf("serve: %+v", *l)

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
		} // TODO else
	}
}

func (this *listener) Listen() error {
	// This function sets up a goroutine to accept inbound connections.
	// The accepted connection will be added to a list of accepted
	// connections.  The Listener just needs to listen continuously,
	// as we assume that we want to continue to receive inbound
	// connections without limit.

	if err := this.l.Listen(); err != nil {
		return err
	}

	this.sock.listeners = append(this.sock.listeners, this)
	this.sock.Lock()
	this.sock.active = true
	this.sock.Unlock()
	Debugf("sock is active")

	go this.serve()
	return nil
}

func (this *listener) Address() string {
	return this.addr
}

func (this *listener) Close() error {
	return this.l.Close()
}
