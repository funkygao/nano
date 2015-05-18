package nano

import (
	"bytes"
)

// Socket is the main access handle applications use to access the SP
// system.  It is an abstraction of an application's "connection" to a
// messaging topology.  Applications can have more than one Socket open
// at a time.
// A single Socket might have connections to multiple endpoints, which is
// the basic difference from POSIX socket.
type Socket interface {

	// Close closes the open Socket.  Further operations on the socket
	// will return ErrClosed.
	Close() error

	// Send puts the message on the outbound send.  It always succeeds,
	// unless the buffer(s) are full.  Once the system takes ownership of
	// the message, it guarantees to deliver the message or keep trying as
	// long as the Socket is open.
	// Depending on the protocol, a single Send might send message to
	// multiple endpoints.
	Send([]byte) error

	// XSend is same as Recv except that it accept buffer as arg.
	XSend(*bytes.Buffer) error

	// Recv receives a complete message.  The entire message is received.
	// Depending on the protocol, a single Recv might receive message from
	// multiple endpoints.
	Recv() ([]byte, error)

	// XRecv is same as Recv except that it can make use of buffer to
	// reduce GC presure.
	XRecv(*bytes.Buffer) (n int, err error)

	// SendMsg puts the message on the outbound queue.  It works like Send,
	// but allows the caller to supply message headers.  AGAIN, the Socket
	// ASSUMES OWNERSHIP OF THE MESSAGE.
	SendMsg(*Message) error

	// RecvMsg receives a complete message, including the message header,
	// which is useful for protocols in raw mode.
	RecvMsg() (*Message, error)

	// Dial connects a remote endpoint to the Socket.  The function
	// returns immediately, and an asynchronous goroutine is started to
	// establish and maintain the connection, reconnecting as needed.
	// If the address is invalid, then an error is returned.
	Dial(addr string) error

	DialOptions(addr string, options map[string]interface{}) error

	// NewDialer returns a Dialer object which can be used to get
	// access to the underlying configuration for dialing.
	NewDialer(addr string, options map[string]interface{}) (Dialer, error)

	// Listen connects a local endpoint to the Socket.  Remote peers
	// may connect (e.g. with Dial) and will each be "connected" to
	// the Socket.  The accepter logic is run in a separate goroutine.
	// The only error possible is if the address is invalid.
	Listen(addr string) error

	ListenOptions(addr string, options map[string]interface{}) error

	NewListener(addr string, options map[string]interface{}) (Listener, error)

	// GetOption is used to retrieve an option for a socket.
	GetOption(name string) (interface{}, error)

	// SetOption is used to set an option for a socket.
	SetOption(name string, value interface{}) error

	// Protocol is used to get the underlying Protocol.
	GetProtocol() Protocol

	// AddTransport adds a new Transport to the socket.  Transport specific
	// options may have been configured on the Transport prior to this.
	AddTransport(Transport)

	// SetPortHook sets a PortHook function to be called when a Port is
	// added or removed from this socket (connect/disconnect).  The previous
	// hook is returned (nil if none.)
	SetPortHook(PortHook) PortHook
}
