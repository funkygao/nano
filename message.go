package nano

import (
	"sync/atomic"
	"time"
)

// Message encapsulates the messages that we exchange back and forth.  The
// meaning of the Header and Body fields, and where the splits occur, will
// vary depending on the protocol.  Note however that any headers applied by
// transport layers (including TCP/ethernet headers, and SP protocol
// independent length headers), are *not* included in the Header.
type Message struct {
	Header []byte
	Body   []byte

	headerBuf []byte
	bodyBuf   []byte

	slabSize int
	refCount int32
}

type messageSlab struct {
	maxBody int
	cache   chan *Message
}

var messagePool = []messageSlab{
	{maxBody: 64, cache: make(chan *Message, 2048)},   // 128K
	{maxBody: 128, cache: make(chan *Message, 1024)},  // 128K
	{maxBody: 1024, cache: make(chan *Message, 1024)}, // 1 MB
	{maxBody: 8192, cache: make(chan *Message, 256)},  // 2 MB
	{maxBody: 65536, cache: make(chan *Message, 64)},  // 4 MB
}

// Free decrements the reference count on a message, and releases its
// resources if no further references remain.  While this is not
// strictly necessary thanks to GC, doing so allows for the resources to
// be recycled without engaging GC.  This can have rather substantial
// benefits for performance.
func (this *Message) Free() {
	if refCount := atomic.AddInt32(&this.refCount, -1); refCount > 0 {
		return
	}

	// safe to put back message pool for later reuse
	var ch chan *Message
	for _, slab := range messagePool {
		if this.slabSize == slab.maxBody {
			ch = slab.cache
			break
		}
	}

	select {
	case ch <- this:
	default:
		// message pool is full, just discard it
	}
}

// Dup creates a "duplicate" message.  What it really does is simply
// increment the reference count on the message.  Note that since the
// underlying message is actually shared, consumers must take care not
// to modify the message.  (We might revise this API in the future to
// add a copy-on-write facility, but for now modification is neither
// needed nor supported.)  Applications should *NOT* make use of this
// function -- it is intended for Protocol, Transport and internal use only.
func (this *Message) Dup() *Message {
	atomic.AddInt32(&this.refCount, 1)
	return this
}

// NewMessage is the supported way to obtain a new Message.  This makes
// use of a "slab allocator" which greatly reduces the load on the
// garbage collector.
func NewMessage(sz int) *Message {
	var msg *Message
	var ch chan *Message
	for _, slab := range messagePool {
		if sz <= slab.maxBody {
			ch = slab.cache
			sz = slab.maxBody
			break
		}
	}

	select {
	case msg = <-ch:
	default:
		// message pool empty
		msg = &Message{}
		msg.slabSize = sz
		msg.bodyBuf = make([]byte, 0, msg.slabSize)
		msg.headerBuf = make([]byte, 0, 32) // TODO
	}

	msg.refCount = 1
	msg.Body = msg.bodyBuf
	msg.Header = msg.headerBuf
	return msg
}

func monitorMessagePool() {
	if !Debug {
		return
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for _ = range ticker.C {
		for _, mp := range messagePool {
			Debugf("%5d: %d", mp.maxBody, len(mp.cache))
		}
	}
}
