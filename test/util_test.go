package test

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/transport/tcp"
)

func TestStripScheme(t *testing.T) {
	trans := tcp.NewTransport()
	addr, err := nano.StripScheme(trans, "tcp://192.168.0.111:5555")
	assert.Equal(t, nil, err)
	assert.Equal(t, "192.168.0.111:5555", addr)

	addr, err = nano.StripScheme(trans, "tcp://eth1;127.0.0.1:1234")
	assert.Equal(t, "eth1;127.0.0.1:1234", addr)
	assert.Equal(t, nil, err)
}

func TestFlattenOptions(t *testing.T) {
	opts := map[string]interface{}{
		"nodelay":   true,
		"keepalive": false,
	}
	r := nano.FlattenOptions(opts)
	t.Logf("%+v", r)
	assert.Equal(t, 4, len(r))
	for idx, v := range r {
		if v == "nodelay" {
			assert.Equal(t, true, r[idx+1].(bool))
		}
		if v == "keepalive" {
			assert.Equal(t, false, r[idx+1].(bool))
		}
	}
}

func BenchmarkStripScheme(b *testing.B) {
	trans := tcp.NewTransport()
	var addr = "tcp://192.168.0.111:5555"
	for i := 0; i < b.N; i++ {
		nano.StripScheme(trans, addr)
	}
}
