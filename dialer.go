package nano

// Dialer is an interface to the underlying dialer for a transport
// and address.
type Dialer interface {

	// Close closes the dialer, and removes it from any active socket.
	// Further operations on the Dialer will return ErrClosed.
	Close() error

	// Dial starts connecting on the address.  If a connection fails,
	// it will restart.
	Dial() error

	// Address returns the string (full URL) of the Listener.
	Address() string

	// SetOption sets an option the Listener. Setting options
	// can only be done before Listen() has been called.
	SetOption(name string, value interface{}) error

	// GetOption gets an option value from the Listener.
	GetOption(name string) (interface{}, error)
}
