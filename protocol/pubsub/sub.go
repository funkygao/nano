package pubsub

import (
	"bytes"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type sub struct {
	sock nano.ProtocolSocket
	subs [][]byte
	raw  bool
	sync.Mutex
}

func (s *sub) Init(sock nano.ProtocolSocket) {
	s.sock = sock
	s.subs = [][]byte{}
	s.sock.SetSendError(nano.ErrProtoOp)
}

func (s *sub) AddEndpoint(ep nano.Endpoint) {
	go s.receiver(ep)
}

func (*sub) RemoveEndpoint(nano.Endpoint) {}

func (s *sub) receiver(ep nano.Endpoint) {
	recvChan := s.sock.RecvChannel()
	closeChan := s.sock.CloseChannel()
	var msg *nano.Message
	for {
		msg = ep.RecvMsg()
		if msg == nil {
			// endpoint closed
			return
		}

		var matched = false
		s.Lock()
		for _, sub := range s.subs {
			if bytes.HasPrefix(msg.Body, sub) {
				// Matched, send it up.  Best effort.
				matched = true
				break
			}
		}
		s.Unlock()

		if !matched {
			msg.Free()
			continue
		}

		select {
		case recvChan <- msg:

		case <-closeChan:
			msg.Free()
			return

		default: // no room, drop it
			msg.Free()
		}
	}
}

func (*sub) Shutdown(time.Time) {} // No sender to drain.

func (*sub) Number() uint16 {
	return nano.ProtoSub
}

func (*sub) PeerNumber() uint16 {
	return nano.ProtoPub
}

func (*sub) Name() string {
	return "sub"
}

func (*sub) PeerName() string {
	return "pub"
}

func (*sub) Handshake() bool {
	return true
}

func (s *sub) SetOption(name string, value interface{}) error {
	s.Lock()
	defer s.Unlock()

	var vb []byte
	var ok bool

	// Check names first, because type check below is only valid for
	// subscription options.
	switch name {
	case nano.OptionRaw:
		if s.raw, ok = value.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil
	case nano.OptionSubscribe:
	case nano.OptionUnsubscribe:
	default:
		return nano.ErrBadOption
	}

	switch v := value.(type) {
	case []byte:
		vb = v
	case string:
		vb = []byte(v)
	default:
		return nano.ErrBadValue
	}
	switch name {
	case nano.OptionSubscribe:
		for _, sub := range s.subs {
			if bytes.Equal(sub, vb) {
				// Already present
				return nil
			}
		}
		s.subs = append(s.subs, vb)
		return nil

	case nano.OptionUnsubscribe:
		for i, sub := range s.subs {
			if bytes.Equal(sub, vb) {
				s.subs[i] = s.subs[len(s.subs)-1]
				s.subs = s.subs[:len(s.subs)-1]
				return nil
			}
		}
		// Subscription not present
		return nano.ErrBadValue

	default:
		return nano.ErrBadOption
	}
}

func (s *sub) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return s.raw, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewSubSocket allocates a new Socket using the SUB protocol.
func NewSubSocket() nano.Socket {
	return nano.MakeSocket(&sub{})
}
