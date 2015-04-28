package nano

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"sync"
)

// connPipe implements the Pipe interface on top of net.Conn.  The
// assumption is that transports using this have similar wire protocols,
// and connPipe is meant to be used as a building block.
type connPipe struct {
	c      net.Conn
	rlock  sync.Mutex
	wlock  sync.Mutex
	reader *bufio.Reader
	writer *bufio.Writer
	proto  Protocol
	open   bool // true after handshake
	props  map[string]interface{}
}

// NewConnPipe allocates a new Pipe using the supplied net.Conn, and
// initializes it.  It performs the handshake required at the SP layer,
// only returning the Pipe once the SP layer negotiation is complete.
//
// Stream oriented transports can utilize this to implement a Transport.
func NewConnPipe(c net.Conn, proto Protocol, props ...interface{}) (Pipe, error) {
	this := &connPipe{
		c:      c,
		reader: bufio.NewReaderSize(c, defaultBufferSize),
		writer: bufio.NewWriterSize(c, defaultBufferSize),
		proto:  proto,
		props:  make(map[string]interface{}),
	}

	this.props[PropLocalAddr] = c.LocalAddr()
	this.props[PropRemoteAddr] = c.RemoteAddr()
	if len(props)%2 != 0 {
		return nil, ErrBadOption
	}
	for i := 0; i+1 < len(props); i += 2 {
		this.props[props[i].(string)] = props[i+1]
	}

	Debugf("proto:%s, props:%v", proto.Name(), this.props)

	if err := this.handshake(); err != nil {
		return nil, err
	}

	return this, nil
}

// handshake establishes an SP connection between peers.  Both sides must
// send the header, then both sides must wait for the peer's header.
// As a side effect, the peer's protocol number is stored in the conn.
func (this *connPipe) handshake() error {
	type connHeader struct {
		Zero    byte   // must be zero
		S       byte   // 'S'
		P       byte   // 'P'
		Version byte   // only zero at present
		Proto   uint16 // protocol type
		Rsvd    uint16 // always zero at present
	}

	var err error
	var header = connHeader{S: 'S', P: 'P', Proto: this.proto.Number()}
	if err = binary.Write(this.c, binary.BigEndian, &header); err != nil {
		return err
	}
	Debugf("send header: %v", header)

	if err = binary.Read(this.c, binary.BigEndian, &header); err != nil {
		this.c.Close()
		return err
	}
	if header.Zero != 0 || header.S != 'S' || header.P != 'P' || header.Rsvd != 0 {
		this.c.Close()
		return ErrBadHeader
	}
	// The only version number we support at present is "0"
	if header.Version != 0 {
		this.c.Close()
		return ErrBadVersion
	}

	// The protocol number lives as 16-bits (big-endian)
	if header.Proto != this.proto.PeerNumber() {
		this.c.Close()
		return ErrBadProto
	}

	Debugf("recv header: %v", header)
	this.open = true
	return nil
}

// RecvMsg implements the Pipe RecvMsg method.  The message received is expected as
// a 64-bit size (network byte order) followed by the message itself.
func (this *connPipe) RecvMsg() (*Message, error) {
	var sz int64
	var err error
	var msg *Message

	// prevent interleaved reads
	this.rlock.Lock()

	// TODO bufio will read sz and body at once if msg is not too big
	if err = binary.Read(this.c, binary.BigEndian, &sz); err != nil {
		this.rlock.Unlock()
		return nil, err
	}

	Debugf("sz: %d", sz)

	// TODO: This fixed limit is kind of silly, but it keeps
	// a bogus peer from causing us to try to allocate ridiculous
	// amounts of memory.  If you don't like it, then prealloc
	// a buffer.  But for protocols that only use small messages
	// this can actually be more efficient since we don't allocate
	// any more space than our peer says we need to.
	if sz > 1024*1024 || sz < 0 {
		this.c.Close()
		this.rlock.Unlock()
		return nil, ErrTooLong
	}

	msg = NewMessage(int(sz))
	msg.Body = msg.Body[0:sz] // the msg may be pulled from message pool, so reset body
	if _, err = io.ReadFull(this.c, msg.Body); err != nil {
		msg.Free()
		this.rlock.Unlock()
		return nil, err
	}

	this.rlock.Unlock()
	return msg, nil
}

// SendMsg implements the Pipe SendMsg method.  The message is sent as a 64-bit
// size (network byte order) followed by the message itself.
// TODO SendMsg seems not to match RecvMsg contents, header ignored?
func (this *connPipe) SendMsg(msg *Message) error {
	sz := uint64(len(msg.Header) + len(msg.Body))

	// prevent interleaved writes
	this.wlock.Lock()

	// send length header TODO bufio
	if err := binary.Write(this.c, binary.BigEndian, sz); err != nil {
		this.wlock.Unlock()
		return err
	}

	if _, err := this.c.Write(msg.Header); err != nil {
		this.wlock.Unlock()
		return err
	}
	if _, err := this.c.Write(msg.Body); err != nil {
		this.wlock.Unlock()
		return err
	}

	Debugf("sz: %d, h:%v, b:%s", sz, msg.Header, string(msg.Body))

	this.wlock.Unlock()
	msg.Free()
	return nil
}

// LocalProtocol returns our local protocol number.
func (this *connPipe) LocalProtocol() uint16 {
	return this.proto.Number()
}

// RemoteProtocol returns our peer's protocol number.
func (this *connPipe) RemoteProtocol() uint16 {
	return this.proto.PeerNumber()
}

// Close implements the Pipe Close method.
func (this *connPipe) Close() error {
	this.open = false
	return this.c.Close()
}

// IsOpen implements the PipeIsOpen method.
func (this *connPipe) IsOpen() bool {
	return this.open
}

func (this *connPipe) GetProp(name string) (interface{}, error) {
	if v, ok := this.props[name]; ok {
		return v, nil
	}
	return nil, ErrBadProperty
}

// connPipeIpc is *almost* like a regular connPipe, but the IPC protocol insists
// on stuffing a leading byte (valued 1) in front of messages.  This is for
// compatibility with nanomsg -- the value cannot ever be anything but 1.
type connPipeIpc struct {
	connPipe
}

// NewConnPipeIPC allocates a new Pipe using the IPC exchange protocol.
func NewConnPipeIPC(c net.Conn, proto Protocol, props ...interface{}) (Pipe, error) {
	this := &connPipeIpc{connPipe: connPipe{
		c:     c,
		proto: proto,
		props: make(map[string]interface{}),
	}}

	this.props[PropLocalAddr] = c.LocalAddr()
	this.props[PropRemoteAddr] = c.RemoteAddr()
	if len(props)%2 != 0 {
		return nil, ErrBadOption
	}
	for i := 0; i+1 < len(props); i += 2 {
		this.props[props[i].(string)] = props[i+1]
	}

	if err := this.handshake(); err != nil {
		return nil, err
	}

	return this, nil
}

func (this *connPipeIpc) SendMsg(msg *Message) error {
	sz := uint64(len(msg.Header) + len(msg.Body))
	one := [1]byte{1}
	var err error

	// prevent interleaved writes
	this.wlock.Lock()

	// send length header
	if _, err = this.c.Write(one[:]); err != nil {
		this.wlock.Unlock()
		return err
	}
	if err = binary.Write(this.c, binary.BigEndian, sz); err != nil {
		this.wlock.Unlock()
		return err
	}
	if _, err = this.c.Write(msg.Header); err != nil {
		this.wlock.Unlock()
		return err
	}
	if _, err = this.c.Write(msg.Body); err != nil {
		this.wlock.Unlock()
		return err
	}

	this.wlock.Unlock()
	msg.Free()
	return nil
}

func (this *connPipeIpc) RecvMsg() (*Message, error) {
	var sz int64
	var err error
	var msg *Message
	var one [1]byte

	// prevent interleaved reads
	this.rlock.Lock()

	if _, err = this.c.Read(one[:]); err != nil {
		this.rlock.Unlock()
		return nil, err
	}
	if err = binary.Read(this.c, binary.BigEndian, &sz); err != nil {
		this.rlock.Unlock()
		return nil, err
	}

	// TODO
	if sz > 1024*1024 || sz < 0 {
		this.c.Close()
		this.rlock.Unlock()
		return nil, ErrTooLong
	}

	msg = NewMessage(int(sz))
	msg.Body = msg.Body[0:sz]
	if _, err = io.ReadFull(this.c, msg.Body); err != nil {
		this.rlock.Unlock()
		msg.Free()
		return nil, err
	}

	this.rlock.Unlock()
	return msg, nil
}
