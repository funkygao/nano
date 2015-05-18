package test

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
)

func TestMessageDup(t *testing.T) {
	msg := nano.NewMessage(0)
	msg1 := msg.Dup()
	assert.Equal(t, true, msg == msg1)
}

func TestMessagePoolNormal(t *testing.T) {
	msg := nano.NewMessage(5)
	assert.Equal(t, 64, cap(msg.Body))
	msg.Free()
	msg = nano.NewMessage(1086)
	assert.Equal(t, 8192, cap(msg.Body))
	msg.Free()
}

func TestMessagePoolEdgeCase(t *testing.T) {
	msg := nano.NewMessage(64)
	assert.Equal(t, 64, cap(msg.Body))
	msg.Free()
	msg = nano.NewMessage(1024)
	assert.Equal(t, 1024, cap(msg.Body))
	msg.Free()
	msg = nano.NewMessage(8192)
	assert.Equal(t, 8192, cap(msg.Body))
}

func TestMessageRecyle(t *testing.T) {
	msg := nano.NewMessage(5)
	msg = msg.Dup()
	msg = msg.Dup()
	assert.Equal(t, false, msg.Free())
	assert.Equal(t, false, msg.Free())
	assert.Equal(t, true, msg.Free())

	// free on an already free'ed msg
	assert.Equal(t, true, msg.Free())
}

func TestMessageReadSlice(t *testing.T) {
	msg := nano.NewMessage(100)
	msg.Body = []byte("a b c d\ne f g,\n")
	l1 := msg.ReadSlice('\n')
	assert.Equal(t, "a b c d\n", string(l1))
	l2 := msg.ReadSlice('\n')
	assert.Equal(t, "e f g,\n", string(l2))
	l3 := msg.ReadSlice('\n')
	assert.Equal(t, "", string(l3))
}

func TestMessageReadFull(t *testing.T) {
	msg := nano.NewMessage(100)
	msg.Body = []byte("a b c d\ne f g,\n")
	msg.ReadSlice('\n')
	line := msg.ReadFull()
	assert.Equal(t, "e f g,\n", string(line))
	line = msg.ReadFull()
	assert.Equal(t, "", string(line))
}
