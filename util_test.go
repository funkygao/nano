package nano

import (
	"testing"

	"github.com/funkygao/assert"
)

type dummyTransport struct{}

func (this *dummyTransport) Scheme() string {
	return "tcp"
}

func (this *dummyTransport) NewDialer(url string, protocol Protocol) (PipeDialer, error) {
	return nil, nil
}

func (this *dummyTransport) NewListener(url string, protocol Protocol) (PipeListener, error) {
	return nil, nil
}

func TestStripScheme(t *testing.T) {
	trans := &dummyTransport{}
	addr, err := StripScheme(trans, "tcp://127.0.0.1:1234")
	assert.Equal(t, "127.0.0.1:1234", addr)
	assert.Equal(t, nil, err)

	addr, err = StripScheme(trans, "tcp://eth1;127.0.0.1:1234")
	assert.Equal(t, "eth1;127.0.0.1:1234", addr)
	assert.Equal(t, nil, err)

}
