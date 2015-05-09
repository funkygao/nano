package pubsub

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type pubEp struct {
	ep nano.Endpoint

	q chan *nano.Message
	w nano.Waiter
}

func (pe *pubEp) peerSender() {
	var msg *nano.Message
	for {
		msg = <-pe.q
		if msg == nil {
			break
		}

		if pe.ep.SendMsg(msg) != nil {
			msg.Free()
			break
		}
	}

	pe.w.Done()
}

// pub will listen and accept subs.
type pub struct {
	sock   nano.ProtocolSocket
	eps    map[nano.EndpointId]*pubEp // endpoints
	raw    bool
	waiter nano.Waiter

	sync.Mutex
}

func (p *pub) Init(sock nano.ProtocolSocket) {
	p.sock = sock
	p.eps = make(map[nano.EndpointId]*pubEp)

	// send only
	p.sock.SetRecvError(nano.ErrProtoOp)

	p.waiter.Init()
	p.waiter.Add()
	go p.sender()
}

func (p *pub) AddEndpoint(ep nano.Endpoint) {
	qlen := 16
	if i, err := p.sock.GetOption(nano.OptionWriteQLen); err == nil {
		qlen = i.(int)
	}

	subscriber := &pubEp{
		ep: ep,
		q:  make(chan *nano.Message, qlen),
	}
	p.Lock()
	p.eps[ep.Id()] = subscriber
	p.Unlock()

	subscriber.w.Init()
	subscriber.w.Add()
	go subscriber.peerSender()
	go nano.NullRecv(ep) // avoid goroutine leakage
}

func (p *pub) RemoveEndpoint(ep nano.Endpoint) {
	p.Lock()
	delete(p.eps, ep.Id())
	p.Unlock()
}

func (p *pub) sender() {
	defer p.waiter.Done()

	sendChan := p.sock.SendChannel()
	closeChan := p.sock.CloseChannel()
	var msg *nano.Message
	for {
		select {
		case <-closeChan:
			return

		case msg = <-sendChan:
			// copy message to each endpoints
			// if no subscribers, drop the msg
			p.Lock()
			for _, peer := range p.eps {
				m := msg.Dup()
				select {
				case peer.q <- m:

				default:
					// sub queue full, silently drop
					m.Free()
				}
			}
			msg.Free()
			p.Unlock()
		}
	}
}

func (p *pub) Shutdown(expire time.Time) {
	p.waiter.WaitAbsTimeout(expire)

	p.Lock()
	peers := p.eps
	p.eps = make(map[nano.EndpointId]*pubEp)
	p.Unlock()

	for id, peer := range peers {
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
		delete(peers, id)
	}
}

func (*pub) Number() uint16 {
	return nano.ProtoPub
}

func (*pub) PeerNumber() uint16 {
	return nano.ProtoSub
}

func (p *pub) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if p.raw, ok = v.(bool); !ok {
			return nano.ErrBadValue
		}
		return nil

	default:
		return nano.ErrBadOption
	}
}

func (p *pub) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return p.raw, nil

	default:
		return nil, nano.ErrBadOption
	}
}

// NewPubSocket allocates a new Socket using the PUB protocol.
func NewPubSocket() nano.Socket {
	return nano.MakeSocket(&pub{})
}
