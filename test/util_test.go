package test

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/transport/tcp"
)

func TestStripScheme(t *testing.T) {
	trans := tcp.NewTransport()
	s, err := nano.StripScheme(trans, "tcp://192.168.0.111:5555")
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.111:5555", s)
}

func BenchmarkStripScheme(b *testing.B) {
	trans := tcp.NewTransport()
	var addr = "tcp://192.168.0.111:5555"
	for i := 0; i < b.N; i++ {
		nano.StripScheme(trans, addr)
	}
}
