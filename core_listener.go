package nano

// listener implements the Listener interface.
type listener struct {
	l PipeListener // created by Transport

	sock *socket // local bind addr
	addr string
}

func (this *listener) Listen() error {
	this.sock.Lock()
	if this.sock.active {
		this.sock.Unlock()
		return ErrAddrInUse
	}

	this.sock.active = true
	this.sock.Unlock()

	Debugf("sock is active")

	if err := this.l.Listen(); err != nil {
		return err
	}

	// keep serving connections
	go this.serve()

	return nil
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

		connPipe, err := l.l.Accept() // will handshake
		if err == nil {
			Debugf("accepted: %+v", connPipe)
			l.sock.addPipe(connPipe, nil, l)
		} else {
			// If the underlying PipeListener is closed, or not
			// listening, we expect to return back with an error.
			if err == ErrClosed {
				return
			} else {
				// TODO
				Debugf("%v", err)
			}
		}

	}
}

func (this *listener) GetOption(name string) (interface{}, error) {
	return this.l.GetOption(name)
}

func (this *listener) SetOption(name string, val interface{}) error {
	return this.l.SetOption(name, val)
}

func (this *listener) Address() string {
	return this.addr
}

func (this *listener) Close() error {
	return this.l.Close()
}
