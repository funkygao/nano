package nano

import (
	"time"
)

// dialer implements the Dailer interface.
type dialer struct {
	d         PipeDialer // created by Transport
	sock      *socket
	addr      string // remote server addr
	closed    bool
	closeChan chan struct{}
}

func (this *dialer) Dial() error {
	this.sock.Lock()
	if this.sock.active {
		this.sock.Unlock()
		return ErrAddrInUse
	}

	this.closeChan = make(chan struct{})
	this.sock.active = true
	this.sock.Unlock()

	Debugf("sock is active, go dialing...")

	// keep dialing
	go this.dialing()

	return nil
}

func (this *dialer) Close() error {
	this.sock.Lock()
	if this.closed {
		this.sock.Unlock()
		return ErrClosed
	}

	Debugf("dialer closed")

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

// dialing is used to dial or redial from a goroutine.
// TODO OptionMaxRetry?
func (this *dialer) dialing() {
	rtime := this.sock.redialTime
	for {
		connPipe, err := this.d.Dial()
		if err == nil {
			// reset retry time
			rtime = this.sock.redialTime

			this.sock.Lock()
			if this.closed {
				this.sock.Unlock()
				connPipe.Close()
				return
			}
			this.sock.Unlock()

			if cp := this.sock.addPipe(connPipe, this, nil); cp != nil {
				// sleep till pipe broken, and then redial
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
			if rtime > this.sock.redialMax {
				rtime = this.sock.redialMax
			}
			Debugf("%s", rtime)
			continue
		}
	}
}
