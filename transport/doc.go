// Package transport implements nano.Transport interface.
//
/*
type Transport interface {
	// Scheme returns a string used as the prefix for SP "addresses".
	// This is similar to a URI scheme.  For example, schemes can be
	// "tcp" (for "tcp://xxx..."), "ipc", "inproc", etc.
	Scheme() string

	// NewDialer creates a new Dialer for this Transport.
	NewDialer(url string, protocol Protocol) (PipeDialer, error)

	// NewListener creates a new PipeListener for this Transport.
	// This generally also arranges for an OS-level file descriptor to be
	// opened, and bound to the the given address, as well as establishing
	// any "listen" backlog.
	NewListener(url string, protocol Protocol) (PipeListener, error)
}
*/
package transport
