package ipc

import (
	"runtime"
	"testing"

	"github.com/funkygao/nano/test"
)

var tt = test.NewTranTest(NewTransport(), "ipc:///tmp/test1234")

func TestIpcListenAndAccept(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("IPC not supported on Windows")
	case "plan9":
		t.Skip("IPC not supported on Plan9")
	default:
		tt.TranTestListenAndAccept(t)
	}
}

func TestIpcDuplicateListen(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("IPC not supported on Windows")
	case "plan9":
		t.Skip("IPC not supported on Plan9")
	default:
		tt.TranTestDuplicateListen(t)
	}
}

func TestIpcConnRefused(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("IPC not supported on Windows")
	case "plan9":
		t.Skip("IPC not supported on Plan9")
	default:
		tt.TranTestConnRefused(t)
	}
}

func TestIpcSendRecv(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("IPC not supported on Windows")
	case "plan9":
		t.Skip("IPC not supported on Plan9")
	default:
		tt.TranTestSendRecv(t)
	}
}

func TestIpcAll(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		t.Skip("IPC not supported on Windows")
	case "plan9":
		t.Skip("IPC not supported on Plan9")
	default:
		tt.TranTestAll(t)
	}
}
