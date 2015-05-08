package transport

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/transport/inproc"
	"github.com/funkygao/nano/transport/ipc"
	"github.com/funkygao/nano/transport/tcp"
	"github.com/funkygao/nano/transport/tlstcp"
)

// AddAll registers all known transports to the given socket.
// This allows a program to support all known transports as
// well as supporting as yet-unknown transports, with a single command.
func AddAll(sock nano.Socket) {
	sock.AddTransport(tcp.NewTransport())
	sock.AddTransport(inproc.NewTransport())
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tlstcp.NewTransport())
}

func AddAllOptions(sock nano.Socket, opts ...interface{}) {
	sock.AddTransport(tcp.NewTransport(opts...))
	sock.AddTransport(ipc.NewTransport(opts...))
	sock.AddTransport(inproc.NewTransport()) // TODO
	sock.AddTransport(tlstcp.NewTransport())
}
