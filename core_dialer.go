package nano

import (
	"time"
)

// dialer implements the Dailer interface.
type dialer struct {
	d PipeDialer // created by Transport

	sock      *socket
	addr      string // remote server addr
	closed    bool
	active    bool
	closeChan chan struct{}
}

func (this *dialer) Dial() error {
	this.sock.Lock()
	if this.active {
		this.sock.Unlock()
		return ErrAddrInUse
	}

	this.active = true
	this.sock.active = true
	this.sock.Unlock()

	this.closeChan = make(chan struct{})

	Debugf("sock is active, go dialing...")

	// keep dialing
	go this.dialing()

	return nil
}

// dialing is used to dial or redial from a goroutine.
// TODO OptionMaxRetry?
func (this *dialer) dialing() {
	retry := this.sock.redialTime
	for {
		connPipe, err := this.d.Dial() // will handshake
		if err == nil {
			// reset retry time
			retry = this.sock.redialTime

			this.sock.Lock()
			if this.closed {
				this.sock.Unlock()
				connPipe.Close()
				return
			}
			this.sock.Unlock()

			// add the new endpoint
			cp := this.sock.addPipe(connPipe, this, nil)
			// sleep till pipe broken, and then redial
			select {
			case <-cp.closeChan:
			case <-this.sock.closeChan:
			case <-this.closeChan:
			}
		} else {
			// dial error
			Debugf("%v", err)
		}

		// we're redialing here
		select {
		case <-this.closeChan: // dialer closed
			return

		case <-this.sock.closeChan: // exit if parent socket closed
			return

		case <-time.After(retry):
			retry *= 2
			if retry > this.sock.redialMax {
				retry = this.sock.redialMax
			}
			Debugf("%s", retry)
			continue
		}
	}
}

func (this *dialer) Close() error {
	this.sock.Lock()
	if this.closed {
		this.sock.Unlock()
		return ErrClosed
	}

	this.closed = true
	this.sock.Unlock()

	Debugf("dialer closed")

	close(this.closeChan)
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
