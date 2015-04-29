package main

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/rep"
	"github.com/funkygao/nano/protocol/req"
	"github.com/funkygao/nano/transport"
	"log"
	"os"
	"strings"
	"time"
)

var (
	addr string
)

func init() {
	nano.Debug = false
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func request() {
	sock, err := req.NewSocket()
	dieIfErr(err)

	transport.AddAll(sock)
	err = sock.Dial(addr)
	dieIfErr(err)
	dieIfErr(sock.SetOption(nano.OptionSendDeadline, time.Second))

	for i := 0; i < 10<<20; i++ {
		err = sock.Send([]byte(strings.Repeat("X", 10)))
		dieIfErr(err)

		msg, err := sock.Recv()
		dieIfErr(err)
		log.Println(i, string(msg))
	}

	dieIfErr(sock.Close())

}

func reply() {
	sock, err := rep.NewSocket()
	dieIfErr(err)

	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))
	log.Printf("listening on %s", addr)

	for {
		data, err := sock.Recv()
		dieIfErr(err)

		log.Println(string(data))

		sock.Send([]byte("world"))
	}

}

func usage() {
	log.Printf("Usage: %s <rep|req> <url>", os.Args[0])
	log.Println("url example: tcp://127.0.0.1:1234  ipc://x.sock  inproc://test  tlstcp://127.0.0.1:1234")
	os.Exit(0)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}

	addr = os.Args[2]

	switch os.Args[1] {
	case "rep":
		reply()
	case "req":
		request()
	default:
		usage()
	}
}
