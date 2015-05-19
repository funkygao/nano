package main

import (
	"fmt"
	"os"
	"time"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/mq"
	"github.com/funkygao/nano/transport"
)

const (
	addr = "tcp://127.0.0.1:1234"
)

func init() {
	nano.Debug = true
}

func broker() {
	sock := mq.NewSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))

	for {
		sock.Recv()

	}

}

func sub() {

}

func pub() {
	sock := mq.NewSocket()
	sock.SetOption(nano.OptionWriteQLen, 2)
	transport.AddAll(sock)
	dieIfErr(sock.Dial(addr))

	body := []byte("PUB x\nhello")
	for {
		sock.Send(body)

		time.Sleep(time.Second * 5)
	}

}

func usage() {
	fmt.Printf("Usage: %s <broker|sub|pub>\n", os.Args[0])
	os.Exit(0)
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "broker":
		broker()

	case "pub":
		pub()

	case "sub":
		sub()

	}

}
