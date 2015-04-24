package tlstcp

import (
	"testing"

	"github.com/funkygao/nano/test"
)

var tt = test.NewTranTest(NewTransport(), "tls+tcp://127.0.0.1:3334")

func TestTLSListenAndAccept(t *testing.T) {
	tt.TranTestListenAndAccept(t)
}

func TestTLSDuplicateListen(t *testing.T) {
	tt.TranTestDuplicateListen(t)
}

func TestTLSConnRefused(t *testing.T) {
	tt.TranTestConnRefused(t)
}

func TestTLSSendRecv(t *testing.T) {
	tt.TranTestSendRecv(t)
}

func TestTLSAll(t *testing.T) {
	tt.TranTestAll(t)
}
