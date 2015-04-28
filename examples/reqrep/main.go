package main

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/rep"
	"github.com/funkygao/nano/protocol/req"
	"github.com/funkygao/nano/transport/tcp"
	"log"
	"os"
	"strings"
	"time"
)

const (
	addr = "tcp://127.0.0.1:1234"
)

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func usage() {
	log.Printf("Usage: %s <rep | req>", os.Args[0])
	os.Exit(0)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {
	case "rep":
		reply()
	case "req":
		request()
	default:
		usage()
	}
}

func request() {
	socket, err := req.NewSocket()
	dieIfErr(err)

	socket.AddTransport(tcp.NewTransport())
	err = socket.Dial(addr)
	dieIfErr(err)
	dieIfErr(socket.SetOption(nano.OptionSendDeadline, time.Second))

	for i := 0; i < 1; i++ {
		err = socket.Send([]byte(strings.Repeat("X", 100)))
		dieIfErr(err)

		time.Sleep(time.Second)

		msg, err := socket.Recv()
		dieIfErr(err)
		log.Println(string(msg))
	}

	dieIfErr(socket.Close())
}

func reply() {
	socket, err := rep.NewSocket()
	dieIfErr(err)

	socket.AddTransport(tcp.NewTransport())
	dieIfErr(socket.Listen(addr))

	for {
		data, err := socket.Recv()
		dieIfErr(err)

		log.Println(string(data))

		socket.Send([]byte("world"))
	}

}
