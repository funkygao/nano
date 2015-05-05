package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pubsub"
	"github.com/funkygao/nano/transport"
)

var (
	addr = "tcp://127.0.0.1:1234"
)

func init() {
	nano.Debug = false
}

func usage() {
	fmt.Printf("Usage: %s <pub|sub>\n", os.Args[0])
	os.Exit(0)
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func doPub() {
	sock := pubsub.NewPubSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))

	body := []byte(strings.Repeat("X", 100))
	for {
		dieIfErr(sock.Send(body))
	}

}

func doSub() {
	sock := pubsub.NewSubSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Dial(addr))
	// Empty byte array effectively subscribes to everything
	dieIfErr(sock.SetOption(nano.OptionSubscribe, []byte("")))

	var i int64
	for {
		data, err := sock.Recv()
		dieIfErr(err)

		i++
		fmt.Println(i, string(data))
	}

}

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	switch os.Args[1] {
	case "pub":
		doPub()
	case "sub":
		doSub()
	}

}
