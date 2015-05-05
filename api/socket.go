package api

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pub"
	"github.com/funkygao/nano/protocol/rep"
	"github.com/funkygao/nano/protocol/req"
	"github.com/funkygao/nano/protocol/sub"
	"github.com/funkygao/nano/transport"
)

// Domain is the socket domain or address family.  We use it to indicate
// either normal or raw mode sockets.
type Domain int

const (
	AF_SP Domain = iota
	AF_SP_RAW
)

// Protocol is the numeric abstraction to the various protocols or patterns
// that nano supports
type Protocol int

const (
	PUSH       = Protocol(nano.ProtoPush)
	PULL       = Protocol(nano.ProtoPull)
	PUB        = Protocol(nano.ProtoPub)
	SUB        = Protocol(nano.ProtoSub)
	REQ        = Protocol(nano.ProtoReq)
	REP        = Protocol(nano.ProtoRep)
	SURVEYOR   = Protocol(nano.ProtoSurveyor)
	RESPONDENT = Protocol(nano.ProtoRespondent)
	BUS        = Protocol(nano.ProtoBus)
	PAIR       = Protocol(nano.ProtoPair)
)

type Socket struct {
	sock nano.Socket

	protocol Protocol
	domain   Domain
}

// TODO getsockopt/setsockopt: sendbuf, nodelay, keepalive, etc
func NewSocket(d Domain, p Protocol) (*Socket, error) {
	var err error
	sock := &Socket{protocol: p, domain: d}
	switch p {
	case PUB:
		sock.sock, err = pub.NewSocket()
	case SUB:
		sock.sock, err = sub.NewSocket()
	case REQ:
		sock.sock, err = req.NewSocket()
	case REP:
		sock.sock, err = rep.NewSocket()
	default:
		err = ErrBadProtocol
	}
	if err != nil {
		sock.sock.Close()
		return nil, err
	}

	switch d {
	case AF_SP:
	case AF_SP_RAW:
		err = sock.sock.SetOption(nano.OptionRaw, true)
	default:
		err = ErrBadDomain
	}
	if err != nil {
		sock.sock.Close()
		return nil, err
	}

	transport.AddAll(sock.sock)

	return sock, nil
}

func (this *Socket) Close() error {
	return this.sock.Close()
}

func (this *Socket) Bind(addr string) error {
	return this.sock.Listen(addr)
}

func (this *Socket) Connect(addr string) error {
	return this.sock.Dial(addr)
}

func (this *Socket) Recv() ([]byte, error) {
	msg, err := this.sock.RecvMsg()
	if err != nil {
		return nil, err
	}

	// TODO mem copy overhead
	var b []byte
	if this.domain == AF_SP_RAW {
		b = make([]byte, 0, len(msg.Header)+len(msg.Body))
		copy(b, msg.Header)
		b = append(b, msg.Body...)
	} else {
		b = make([]byte, len(msg.Body))
		copy(b, msg.Body)
	}
	msg.Free()
	return b, nil
}

func (this *Socket) Send(b []byte, flags int) (int, error) {
	msg := nano.NewMessage(len(b))
	msg.Body = append(msg.Body, b...)
	return len(b), this.sock.SendMsg(msg)
	return 0, nil
}
