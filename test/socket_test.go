package test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/req"
)

func TestSocketCloseMoreThanOnce(t *testing.T) {
	sock, _ := req.NewSocket()
	err := sock.Close()
	assert.Equal(t, nil, err)
	err = sock.Close()
	assert.Equal(t, nano.ErrClosed, err)
}

func TestSocketConcurrentClose(t *testing.T) {
	sock, _ := req.NewSocket()
	var n int32
	var wg sync.WaitGroup
	const c = 10
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			if sock.Close() != nil {
				atomic.AddInt32(&n, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(c-1), n)
}
