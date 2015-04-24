package nano

// Listener is an interface to the underlying listener for a transport
// and address.
type Listener interface {
	// Close closes the listener, and removes it from any active socket.
	// Further operations on the Listener will return ErrClosed.
	Close() error

	// Listen starts listening for new connectons on the address.
	Listen() error

	// Address returns the string (full URL) of the Listener.
	Address() string

	// SetOption sets an option the Listener. Setting options
	// can only be done before Listen() has been called.
	SetOption(name string, value interface{}) error

	// GetOption gets an option value from the Listener.
	GetOption(name string) (interface{}, error)
}
