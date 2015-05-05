package pubsub

import (
	"sync"
	"time"

	"github.com/funkygao/nano"
)

type pubEp struct {
	ep nano.Endpoint
	q  chan *nano.Message
	p  *pub
	w  nano.Waiter
}

type pub struct {
	sock nano.ProtocolSocket
	eps  map[uint32]*pubEp
	raw  bool
	w    nano.Waiter
	init sync.Once

	sync.Mutex
}

func (p *pub) Init(sock nano.ProtocolSocket) {
	p.sock = sock
	p.eps = make(map[uint32]*pubEp)
	p.sock.SetRecvError(nano.ErrProtoOp)
	p.w.Init()
}

func (p *pub) Shutdown(expire time.Time) {
	p.w.WaitAbsTimeout(expire)

	p.Lock()
	peers := p.eps
	p.eps = make(map[uint32]*pubEp)
	p.Unlock()

	for id, peer := range peers {
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
		delete(peers, id)
	}
}

// Bottom sender.
func (pe *pubEp) peerSender() {
	for {
		m := <-pe.q
		if m == nil {
			break
		}

		if pe.ep.SendMsg(m) != nil {
			m.Free()
			break
		}
	}
}

// Top sender.
func (p *pub) sender() {
	defer p.w.Done()

	sq := p.sock.SendChannel()
	cq := p.sock.CloseChannel()

	for {
		select {
		case <-cq:
			return

		case m := <-sq:

			// copy message to each endpoints
			p.Lock()
			for _, peer := range p.eps {
				m := m.Dup()
				select {
				case peer.q <- m:
				default:
					m.Free()
				}
			}
			p.Unlock()
		}
	}
}

func (p *pub) AddEndpoint(ep nano.Endpoint) {
	p.init.Do(func() {
		p.w.Add()
		go p.sender()
	})
	depth := 16
	if i, err := p.sock.GetOption(nano.OptionWriteQLen); err == nil {
		depth = i.(int)
	}
	pe := &pubEp{ep: ep, p: p, q: make(chan *nano.Message, depth)}
	pe.w.Init()
	p.Lock()
	p.eps[ep.Id()] = pe
	p.Unlock()

	pe.w.Add()
	go pe.peerSender()
	go nano.NullRecv(ep)
}

func (p *pub) RemoveEndpoint(ep nano.Endpoint) {
	p.Lock()
	delete(p.eps, ep.Id())
	p.Unlock()
}

func (*pub) Number() uint16 {
	return nano.ProtoPub
}

func (*pub) PeerNumber() uint16 {
	return nano.ProtoSub
}

func (*pub) Name() string {
	return "pub"
}

func (*pub) PeerName() string {
	return "sub"
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