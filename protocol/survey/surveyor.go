package survey

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/funkygao/nano"
)

const defaultSurveyTime = time.Second

type surveyor struct {
	sock     nano.ProtocolSocket
	peers    map[uint32]*surveyorP
	raw      bool
	nextID   uint32
	surveyID uint32
	duration time.Duration
	timeout  time.Time
	timer    *time.Timer
	w        nano.Waiter
	init     sync.Once

	sync.Mutex
}

type surveyorP struct {
	q  chan *nano.Message
	ep nano.Endpoint
	x  *surveyor
}

func (x *surveyor) Init(sock nano.ProtocolSocket) {
	x.sock = sock
	x.peers = make(map[uint32]*surveyorP)
	x.sock.SetRecvError(nano.ErrProtoState)
	x.timer = time.AfterFunc(x.duration,
		func() { x.sock.SetRecvError(nano.ErrProtoState) })
	x.timer.Stop()
	x.w.Init()
}

func (x *surveyor) Shutdown(expire time.Time) {

	x.w.WaitAbsTimeout(expire)
	x.Lock()
	peers := x.peers
	x.peers = make(map[uint32]*surveyorP)
	x.Unlock()

	for id, peer := range peers {
		delete(peers, id)
		nano.DrainChannel(peer.q, expire)
		close(peer.q)
	}
}

func (x *surveyor) sender() {
	defer x.w.Done()
	sq := x.sock.SendChannel()
	cq := x.sock.CloseChannel()
	for {
		var m *nano.Message
		select {
		case m = <-sq:
		case <-cq:
			return
		}

		x.Lock()
		for _, pe := range x.peers {
			m := m.Dup()
			select {
			case pe.q <- m:
			default:
				m.Free()
			}
		}
		x.Unlock()
	}
}

// When sending, we should have the survey ID in the header.
func (peer *surveyorP) sender() {
	for {
		if m := <-peer.q; m == nil {
			break
		} else {
			if peer.ep.SendMsg(m) != nil {
				m.Free()
				return
			}
		}
	}
}

func (peer *surveyorP) receiver() {

	rq := peer.x.sock.RecvChannel()
	cq := peer.x.sock.CloseChannel()

	for {
		m := peer.ep.RecvMsg()
		if m == nil {
			return
		}
		if len(m.Body) < 4 {
			m.Free()
			continue
		}

		// Get survery ID -- this will be passed in the header up
		// to the application.  It should include that in the response.
		m.Header = append(m.Header, m.Body[:4]...)
		m.Body = m.Body[4:]

		select {
		case rq <- m:
		case <-cq:
			return
		}
	}
}

func (x *surveyor) AddEndpoint(ep nano.Endpoint) {
	peer := &surveyorP{ep: ep, x: x, q: make(chan *nano.Message, 1)}
	x.init.Do(func() {
		x.w.Add()
		go x.sender()
	})
	x.Lock()
	x.peers[ep.Id()] = peer
	go peer.receiver()
	go peer.sender()
	x.Unlock()
}

func (x *surveyor) RemoveEndpoint(ep nano.Endpoint) {
	x.Lock()
	defer x.Unlock()
	peer := x.peers[ep.Id()]
	if peer == nil {
		return
	}
	delete(x.peers, ep.Id())
}

func (*surveyor) Number() uint16 {
	return nano.ProtoSurveyor
}

func (*surveyor) PeerNumber() uint16 {
	return nano.ProtoRespondent
}

func (*surveyor) Name() string {
	return "surveyor"
}

func (*surveyor) PeerName() string {
	return "respondent"
}

func (*surveyor) Handshake() bool {
	return true
}

func (x *surveyor) SendHook(m *nano.Message) bool {

	if x.raw {
		return true
	}

	x.Lock()
	x.surveyID = x.nextID | 0x80000000
	x.nextID++
	x.sock.SetRecvError(nil)
	v := x.surveyID
	m.Header = append(m.Header,
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))

	if x.duration > 0 {
		x.timer.Reset(x.duration)
	}
	x.Unlock()

	return true
}

func (x *surveyor) RecvHook(m *nano.Message) bool {
	if x.raw {
		return true
	}

	x.Lock()
	defer x.Unlock()

	if len(m.Header) < 4 {
		return false
	}
	if binary.BigEndian.Uint32(m.Header) != x.surveyID {
		return false
	}
	m.Header = m.Header[4:]
	return true
}

func (x *surveyor) SetOption(name string, val interface{}) error {
	var ok bool
	switch name {
	case nano.OptionRaw:
		if x.raw, ok = val.(bool); !ok {
			return nano.ErrBadValue
		}
		if x.raw {
			x.timer.Stop()
			x.sock.SetRecvError(nil)
		} else {
			x.sock.SetRecvError(nano.ErrProtoState)
		}
		return nil
	case nano.OptionSurveyTime:
		x.Lock()
		x.duration, ok = val.(time.Duration)
		x.Unlock()
		if !ok {
			return nano.ErrBadValue
		}
		return nil
	default:
		return nano.ErrBadOption
	}
}

func (x *surveyor) GetOption(name string) (interface{}, error) {
	switch name {
	case nano.OptionRaw:
		return x.raw, nil
	case nano.OptionSurveyTime:
		x.Lock()
		d := x.duration
		x.Unlock()
		return d, nil
	default:
		return nil, nano.ErrBadOption
	}
}

// NewSurveyorSocket allocates a new Socket using the SURVEYOR protocol.
func NewSurveyorSocket() nano.Socket {
	return nano.MakeSocket(&surveyor{duration: defaultSurveyTime})
}
