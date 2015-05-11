package main

import (
	"fmt"
	"os"
	"time"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/bus"
	"github.com/funkygao/nano/transport"
)

func init() {
	nano.Debug = false
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func usage() {
	fmt.Printf("Usage: %s <name> <listen_url> <dial_url> <dial_url>...\n", os.Args[0])
	os.Exit(0)
}

func run(args []string) {
	// Topology establishment
	sock := bus.NewSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(args[2]))

	fmt.Printf("%s -> %v\n", args[1], args[3:])
	time.Sleep(time.Second)

	// Message routing
	for i := 3; i < len(args); i++ {
		dieIfErr(sock.Dial(args[i]))
	}

	time.Sleep(time.Second)

	fmt.Printf("[%s] ===> %s\n", args[1], args[1])
	dieIfErr(sock.Send([]byte(args[1])))
	var msg []byte
	var err error
	for {
		msg, err = sock.Recv()
		dieIfErr(err)

		fmt.Printf("    [%s] <==== %s\n", args[1], string(msg))
	}

}

func main() {
	if len(os.Args) > 3 {
		run(os.Args)
	}

	usage()
}
