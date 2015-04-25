// Package protocol implements nano.Protocol.
//
/*
type Protocol interface {

	// Init is called by the core to allow the protocol to perform
	// any initialization steps it needs.  It should save the handle
	// for future use, as well.
	Init(ProtocolSocket)

	// Shutdown is used to drain the send side.  It is only ever called
	// when the socket is being shutdown cleanly. Protocols should use
	// the linger time, and wait up to that time for sockets to drain.
	Shutdown(time.Time)

	// AddEndpoint is called when a new Endpoint is added to the socket.
	// Typically this is as a result of connect or accept completing.
	AddEndpoint(Endpoint)

	// RemoveEndpoint is called when an Endpoint is removed from the socket.
	// Typically this indicates a disconnected or closed connection.
	RemoveEndpoint(Endpoint)

	// ProtocolNumber returns a 16-bit value for the protocol number,
	// as assigned by the SP governing body. (IANA?)
	Number() uint16

	// Name returns our name.
	Name() string

	// PeerNumber() returns a 16-bit number for our peer protocol.
	PeerNumber() uint16

	// PeerName() returns the name of our peer protocol.
	PeerName() string

	// GetOption is used to retrieve the current value of an option.
	// If the protocol doesn't recognize the option, EBadOption should
	// be returned.
	GetOption(string) (interface{}, error)

	// SetOption is used to set an option.  EBadOption is returned if
	// the option name is not recognized, EBadValue if the value is
	// invalid.
	SetOption(string, interface{}) error
}
*/
package protocol
