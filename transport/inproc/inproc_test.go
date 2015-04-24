package inproc

import (
	"testing"

	"github.com/funkygao/nano/test"
)

var tt = test.NewTranTest(NewTransport(), "inproc://testname")

func TestInpListenAndAccept(t *testing.T) {
	tt.TranTestListenAndAccept(t)
}

func TestInpDuplicateListen(t *testing.T) {
	tt.TranTestDuplicateListen(t)
}

func TestInpConnRefused(t *testing.T) {
	tt.TranTestConnRefused(t)
}

func TestInpSendRecv(t *testing.T) {
	tt.TranTestSendRecv(t)
}

func TestInpSchem(t *testing.T) {
	tt.TranTestScheme(t)
}

func TestInp(t *testing.T) {
	tt.TranTestAll(t)
}
