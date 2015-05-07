package nano

import (
	"strings"
	"sync"
	"time"
)

// socket implements Socket & ProtocolSocket interfaces.
type socket struct {
	proto      Protocol
	transports map[string]Transport

	sync.RWMutex

	sendChan     chan *Message
	sendChanSize int
	recvChan     chan *Message
	recvChanSize int
	closeChan    chan struct{} // closed when user requests close

	closing bool // true if Socket was closed at API level
	active  bool // true if either Dial or Listen has been successfully called

	recvErr error // error to return on attempts to Recv()
	sendErr error // error to return on attempts to Send()

	readDeadline  time.Duration // read deadline, default 0
	writeDeadline time.Duration // write deadline, default 0
	redialTime    time.Duration // reconnect time after error or disconnect
	redialMax     time.Duration // max reconnect interval before give up? TODO SetOption
	linger        time.Duration // wait up to that time for sockets to drain

	// a socket can have multiple endpoints:
	// a listener can accept multiple inbound connections(endpoints);
	// a dialer can dial multiple servers(endpoints).
	//
	// we store pipes so that socket close will gracefully close all endpoints.
	pipes []*pipeEndpoint

	sendHook ProtocolSendHook // pipeline/reqrep/survey are hooking
	recvHook ProtocolRecvHook // bus/reqrep/survey are hooking

	portHook PortHook
}

// MakeSocket is intended for use by Protocol implementations.  The intention
// is that they can wrap this to provide a "proto.NewSocket()" implementation.
func MakeSocket(proto Protocol) Socket {
	Debugf("proto: %v", proto.Name())
	sock := &socket{
		proto:      proto,
		transports: make(map[string]Transport, 1),

		sendChanSize: defaultChanLen,
		sendChan:     make(chan *Message, defaultChanLen), // 128
		recvChanSize: defaultChanLen,
		recvChan:     make(chan *Message, defaultChanLen), // 128
		closeChan:    make(chan struct{}),

		redialTime: defaultRedialTime, // 100ms, backoff with double redial time
		redialMax:  defaultRedialMax,  // 1m
		linger:     defaultLingerTime, // 1s

		pipes: make([]*pipeEndpoint, 0), // TODO
	}

	if hook, ok := proto.(ProtocolRecvHook); ok {
		sock.recvHook = hook
	}
	if hook, ok := proto.(ProtocolSendHook); ok {
		sock.sendHook = hook
	}

	Debugf("sock:%+v", *sock)

	// let protocol plugin initialize
	proto.Init(sock)

	return sock
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
	t, e := sock.getTransport(addr)
	if e != nil {
		return nil, e
	}

	var (
		err error
		d   = &dialer{
			sock: sock,
			addr: addr,
		}
	)
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

	return l.Listen()
}

func (sock *socket) Listen(addr string) error {
	return sock.ListenOptions(addr, nil)
}

// NewListener wraps PipeListener created in Transport.
func (sock *socket) NewListener(addr string, options map[string]interface{}) (Listener, error) {
	// This function sets up a goroutine to accept inbound connections.
	// The accepted connection will be added to a list of accepted
	// connections.  The Listener just needs to listen continuously,
	// as we assume that we want to continue to receive inbound
	// connections without limit.
	t, e := sock.getTransport(addr)
	if e != nil {
		return nil, e
	}

	l := &listener{
		sock: sock,
		addr: addr,
	}
	var err error
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

func (sock *socket) getTransport(addr string) (Transport, error) {
	var i int
	if i = strings.Index(addr, "://"); i < 0 {
		return nil, ErrBadTran
	}

	scheme := addr[:i]
	sock.RLock()
	t, present := sock.transports[scheme]
	sock.RUnlock()
	if present {
		return t, nil
	}

	return nil, ErrBadTran
}

func (sock *socket) AddTransport(t Transport) {
	sock.Lock()
	sock.transports[t.Scheme()] = t
	sock.Unlock()
}

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
	sock.Lock() // sync with SendMsg
	sock.sendErr = err
	sock.Unlock()
}

func (sock *socket) SetRecvError(err error) {
	sock.Lock() // sync with RecvMsg
	sock.recvErr = err
	sock.Unlock()
}

func (sock *socket) Close() error {
	expire := time.Now().Add(sock.linger)
	DrainChannel(sock.sendChan, expire)

	sock.Lock()
	if sock.closing {
		sock.Unlock()
		return ErrClosed
	}

	sock.closing = true
	close(sock.closeChan) // broadcast

	pipes := append([]*pipeEndpoint{}, sock.pipes...)
	sock.Unlock()

	// A second drain, just to be sure.  (We could have had device or
	// forwarded messages arrive since the last one.)
	DrainChannel(sock.sendChan, expire)

	// And tell the protocol to shutdown and drain its pipes too.
	sock.proto.Shutdown(expire)

	Debugf("closing all pipes: %#v", pipes)
	for _, p := range pipes {
		p.Close()
	}

	return nil
}

func (sock *socket) SendMsg(msg *Message) error {
	sock.RLock()
	err := sock.sendErr
	sock.RUnlock()
	if err != nil {
		return err
	}

	Debugf("msg: %+v", *msg)

	if sock.sendHook != nil {
		Debugf("sendHook: %+v", *msg)
		if ok := sock.sendHook.SendHook(msg); !ok {
			// just drop it silently
			Debugf("hook fail: %+v", *msg)
			msg.Free() // safe to recycle
			return nil
		}
	}

	select {
	case <-mkTimer(sock.writeDeadline):
		return ErrSendTimeout

	case <-sock.closeChan:
		return ErrClosed

	case sock.sendChan <- msg:
		return nil
	}
}

func (sock *socket) RecvMsg() (*Message, error) {
	sock.RLock()
	err := sock.recvErr
	sock.RUnlock()
	if err != nil {
		return nil, err
	}

	var (
		timeout = mkTimer(sock.readDeadline)
		msg     *Message
	)
	for {
		select {
		case <-timeout:
			return nil, ErrRecvTimeout

		case msg = <-sock.recvChan:
			Debugf("recv msg: %+v", *msg)

			if sock.recvHook != nil {
				if ok := sock.recvHook.RecvHook(msg); ok {
					Debugf("after RecvHook: %+v", *msg)
					return msg, nil
				} else {
					// drop this msg and get next msg
					msg.Free()
				}
			} else {
				return msg, nil
			}

		case <-sock.closeChan:
			return nil, ErrClosed
		}
	}
}

func (sock *socket) Send(b []byte) error {
	// msg is allocated on stack instead of heap
	// so needn't NewMessage
	msg := &Message{Body: b, Header: nil, refCount: 1}
	return sock.SendMsg(msg)
}

func (sock *socket) Recv() ([]byte, error) {
	msg, err := sock.RecvMsg()
	if err != nil {
		return nil, err
	}

	return msg.Body, nil // FIXME when to msg.Free?
}

func (sock *socket) SetOption(name string, value interface{}) error {
	matched := false
	err := sock.proto.SetOption(name, value)
	if err == nil {
		matched = true
	} else if err != ErrBadOption {
		return err
	}

	sock.Lock()
	defer sock.Unlock()
	switch name {
	case OptionRecvDeadline:
		sock.readDeadline = value.(time.Duration)
		return nil

	case OptionSendDeadline:
		sock.writeDeadline = value.(time.Duration)
		return nil

	case OptionLinger:
		sock.linger = value.(time.Duration)
		return nil

	case OptionWriteQLen:
		if sock.active {
			// will lose data, so forbidden
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
		if sock.active {
			// will lose data, so forbidden
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

	sock.Lock()
	defer sock.Unlock()
	switch name {
	case OptionRecvDeadline:
		return sock.readDeadline, nil

	case OptionSendDeadline:
		return sock.writeDeadline, nil

	case OptionLinger:
		return sock.linger, nil

	case OptionWriteQLen:
		return sock.sendChanSize, nil

	case OptionReadQLen:
		return sock.recvChanSize, nil
	}
	return nil, ErrBadOption
}

func (sock *socket) GetProtocol() Protocol {
	return sock.proto
}

func (sock *socket) SetPortHook(newhook PortHook) PortHook {
	sock.Lock()
	oldhook := sock.portHook
	sock.portHook = newhook
	sock.Unlock()
	return oldhook
}

func (sock *socket) addPipe(connPipe Pipe, d *dialer, l *listener) *pipeEndpoint {
	pe := newPipeEndpoint(connPipe, d, l)
	sock.Lock()
	if fn := sock.portHook; fn != nil {
		sock.Unlock()
		if !fn(PortActionAdd, pe) {
			pe.Close()
			return nil
		}
		sock.Lock()
	}
	pe.sock = sock
	pe.index = len(sock.pipes)
	sock.pipes = append(sock.pipes, pe)
	sock.Unlock()

	// let Protocol register this endpoint
	sock.proto.AddEndpoint(pe)

	return pe
}

func (sock *socket) removePipe(p *pipeEndpoint) {
	sock.proto.RemoveEndpoint(p)

	Debugf("%#v", *p)

	sock.Lock()
	if p.index >= 0 {
		// switch between p and pipes slice last item
		sock.pipes[p.index] = sock.pipes[len(sock.pipes)-1]
		sock.pipes[p.index].index = p.index
		sock.pipes = sock.pipes[:len(sock.pipes)-1]
		p.index = -1
	}
	sock.Unlock()
}
