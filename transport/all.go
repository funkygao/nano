// Package all is used to register all transports.  This allows a program
// to support all known transports as well as supporting as yet-unknown
// transports, with a single import.
package transport

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/transport/inproc"
	"github.com/funkygao/nano/transport/ipc"
	"github.com/funkygao/nano/transport/tcp"
	"github.com/funkygao/nano/transport/tlstcp"
)

// AddAll adds all known transports to the given socket.
func AddAll(sock nano.Socket) {
	sock.AddTransport(tcp.NewTransport())
	sock.AddTransport(inproc.NewTransport())
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tlstcp.NewTransport())
}
