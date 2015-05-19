package nano

// Device is used to create a forwarding loop between two sockets.  If the
// same socket is listed (or either socket is nil), then a loopback device
// is established instead.  Note that the single socket case is only valid
// for protocols where the underlying protocol can peer for itself (e.g. PAIR,
// or BUS, but not REQ/REP or PUB/SUB!)  Both sockets will be placed into RAW
// mode.
//
// If the plumbing is successful, nil will be returned.  Two threads will be
// established to forward messages in each direction.  If either socket returns
// error on receive or send, the goroutine doing the forwarding will exit.
// This means that closing either socket will generally cause the goroutines
// to exit.  Apart from closing the socket(s), no futher operations should be
// performed against the socket.
func Device(s1 Socket, s2 Socket) error {
	// At least one must be non-nil
	if s1 == nil && s2 == nil {
		return ErrClosed
	}

	// Is one of the sockets nil? loopback
	if s1 == nil {
		s1 = s2
	}
	if s2 == nil {
		s2 = s1
	}

	p1 := s1.GetProtocol()
	p2 := s2.GetProtocol()
	if !ValidPeers(p1, p2) {
		return ErrBadProto
	}

	if err := s1.SetOption(OptionRaw, true); err != nil {
		return err
	}
	if err := s2.SetOption(OptionRaw, true); err != nil {
		return err
	}

	go forwarder(s1, s2)
	go forwarder(s2, s1)
	return nil
}

// Forwarder takes messages from one socket, and sends them to the other.
// The sockets must be of compatible types, and must be in Raw mode.
func forwarder(fromSock Socket, toSock Socket) {
	for {
		msg, err := fromSock.RecvMsg()
		if err != nil {
			// ErrClosed or ErrSendTimeout
			return
		}

		err = toSock.SendMsg(msg)
		if err != nil {
			// ErrClosed or ErrSendTimeout
			return
		}
	}
}
