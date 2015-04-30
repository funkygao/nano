// Package transport implements nano.Transport interface.
// Examples of transport url:
// tcp://*:5678  ipc://x.sock  inproc://test  tlstcp://12.1.22.1:5678
/*
type Transport interface {

	Scheme() string

	NewDialer(url string, protocol Protocol) (PipeDialer, error)

	NewListener(url string, protocol Protocol) (PipeListener, error)
}
*/
package transport
