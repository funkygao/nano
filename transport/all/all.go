// Package all is used to register all transports.  This allows a program
// to support all known transports as well as supporting as yet-unknown
// transports, with a single import.
package all

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/transport/inproc"
	"github.com/funkygao/nano/transport/ipc"
	"github.com/funkygao/nano/transport/tcp"
	"github.com/funkygao/nano/transport/tlstcp"
)

// AddTransports adds all known transports to the given socket.
func AddTransports(sock nano.Socket) {
	sock.AddTransport(tcp.NewTransport())
	sock.AddTransport(inproc.NewTransport())
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tlstcp.NewTransport())
}
