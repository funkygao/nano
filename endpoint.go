package nano

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

type EndpointId uint32

var endpointPool struct {
	byid       map[EndpointId]*pipeEndpoint
	nextidChan chan EndpointId

	sync.Mutex
}

func endpointIdGenerator() {
	var nextid = EndpointId(rand.NewSource(time.Now().UnixNano()).Int63())
	var id EndpointId
	for {
		id = nextid & 0x7fffffff // will never conflict with REQ.id
		nextid++
		if id == 0 {
			continue
		}

		endpointPool.nextidChan <- id
	}
}

// pipe wraps the Pipe data structure with the stuff we need to keep
// for the core.  It implements the Endpoint interface.
type pipeEndpoint struct {
	pipe Pipe // connPipe

	listener *listener
	dialer   *dialer
	sock     *socket

	closing   bool          // true if we were closed
	closeChan chan struct{} // notify dialer to redial
	id        EndpointId
	index     int

	sync.Mutex
}

// TODO use sync.Pool
func newPipeEndpoint(connPipe Pipe, d *dialer, l *listener) *pipeEndpoint {
	this := &pipeEndpoint{
		pipe:      connPipe,
		dialer:    d,
		listener:  l,
		index:     -1,
		closeChan: make(chan struct{}),
	}
	for {
		this.id = <-endpointPool.nextidChan

		endpointPool.Lock()
		if _, present := endpointPool.byid[this.id]; !present {
			endpointPool.byid[this.id] = this
			endpointPool.Unlock()
			return this
		}
		endpointPool.Unlock()
	}

	return this
}

func (this *pipeEndpoint) Id() EndpointId {
	return this.id
}

func (this *pipeEndpoint) LocalAddr() net.Addr {
	if addr, err := this.pipe.GetProp(PropLocalAddr); err == nil {
		return addr.(net.Addr)
	}
	return nil
}

func (this *pipeEndpoint) RemoteAddr() net.Addr {
	if addr, err := this.pipe.GetProp(PropRemoteAddr); err == nil {
		return addr.(net.Addr)
	}
	return nil
}

func (this *pipeEndpoint) Close() error {
	var hook PortHook
	this.Lock()
	if this.closing {
		this.Unlock()
		return nil // TODO ErrClosed?
	}
	sock := this.sock
	if sock != nil {
		hook = sock.portHook
	}
	this.closing = true
	this.Unlock()

	close(this.closeChan)
	if sock != nil {
		sock.removePipe(this)
	}
	this.pipe.Close()

	endpointPool.Lock()
	delete(endpointPool.byid, this.id)
	endpointPool.Unlock()

	if hook != nil {
		hook(PortActionRemove, this)
	}

	Debugf("%+v", endpointPool)
	return nil
}

func (this *pipeEndpoint) Flush() error {
	Debugf("flushing")

	if err := this.pipe.Flush(); err != nil {
		this.Close()
		return err
	}

	return nil
}

func (this *pipeEndpoint) SendMsg(msg *Message) error {
	Debugf("msg: %+v, calling %T.SendMsg", *msg, this.pipe)
	if err := this.pipe.SendMsg(msg); err != nil {
		// FIXME error will lead to close?
		this.Close()
		return err
	}

	return nil
}

func (this *pipeEndpoint) RecvMsg() *Message {
	msg, err := this.pipe.RecvMsg()
	if err != nil {
		// e,g connection reset by peer: read a socket that was closed by peer, RST
		// e,g broken pipe: write to socket that was closed by peer
		// e,g read tcp i/o timeout
		// FIXME error will lead to close?
		Debugf("recv msg err: %v, close myself", err)
		this.Close()
		return nil
	}

	Debugf("RecvMsg: %+v", *msg)
	return msg
}

func (this *pipeEndpoint) Address() string {
	switch {
	case this.listener != nil:
		return this.listener.Address()

	case this.dialer != nil:
		return this.dialer.Address()
	}
	return ""
}

func (this *pipeEndpoint) GetProp(name string) (interface{}, error) {
	return this.pipe.GetProp(name)
}

func (this *pipeEndpoint) IsOpen() bool {
	return this.pipe.IsOpen()
}

func (this *pipeEndpoint) IsClient() bool {
	return this.dialer != nil
}

func (this *pipeEndpoint) IsServer() bool {
	return this.listener != nil
}

func (this *pipeEndpoint) LocalProtocol() uint16 {
	return this.pipe.LocalProtocol()
}

func (this *pipeEndpoint) RemoteProtocol() uint16 {
	return this.pipe.RemoteProtocol()
}

func (this *pipeEndpoint) Dialer() Dialer {
	return this.dialer
}

func (this *pipeEndpoint) Listener() Listener {
	return this.listener
}
