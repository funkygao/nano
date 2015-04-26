package nano

import (
	"math/rand"
	"sync"
	"time"
)

var pipes struct {
	byid   map[uint32]*pipeEndpoint
	nextid uint32
	sync.Mutex
}

// pipe wraps the Pipe data structure with the stuff we need to keep
// for the core.  It implements the Endpoint interface.
type pipeEndpoint struct {
	pipe      Pipe
	closeChan chan struct{} // only closed, never passes data
	id        uint32
	index     int // index in master list of pipes for socket

	listener *listener
	dialer   *dialer
	sock     *socket
	closing  bool // true if we were closed

	sync.Mutex
}

func init() {
	pipes.byid = make(map[uint32]*pipeEndpoint)
	pipes.nextid = uint32(rand.NewSource(time.Now().UnixNano()).Int63())
}

func newPipe(tranpipe Pipe) *pipeEndpoint {
	this := &pipeEndpoint{pipe: tranpipe, index: -1}
	this.closeChan = make(chan struct{})
	for {
		pipes.Lock()
		this.id = pipes.nextid & 0x7fffffff
		pipes.nextid++
		Debugf("%d %d", this.id, pipes.nextid)
		if this.id != 0 && pipes.byid[this.id] == nil {
			pipes.byid[this.id] = this
			pipes.Unlock()
			break
		}
		pipes.Unlock()
	}

	Debugf("%+v", pipes)
	return this
}

func (this *pipeEndpoint) Id() uint32 {
	pipes.Lock()
	defer pipes.Unlock()
	return this.id
}

func (this *pipeEndpoint) Close() error {
	var hook PortHook
	this.Lock()
	sock := this.sock
	if sock != nil {
		hook = sock.porthook
	}
	if this.closing {
		return nil
	}
	this.closing = true
	this.Unlock()
	close(this.closeChan)
	if sock != nil {
		sock.removePipe(this)
	}
	this.pipe.Close()
	pipes.Lock()
	delete(pipes.byid, this.id)
	this.id = 0 // safety
	pipes.Unlock()
	if hook != nil {
		hook(PortActionRemove, this)
	}

	Debugf("%+v", pipes)
	return nil
}

func (this *pipeEndpoint) SendMsg(msg *Message) error {
	Debugf("msg: %+v", *msg)
	if err := this.pipe.SendMsg(msg); err != nil {
		this.Close()
		return err
	}
	return nil
}

func (this *pipeEndpoint) RecvMsg() *Message {
	Debugf("RecvMsg: %#v", this.pipe)
	msg, err := this.pipe.RecvMsg()
	if err != nil {
		this.Close()
		return nil
	}
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
