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
	sock := mq.NewBrokerSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))

	for {
		sock.Recv()

	}

}

func sub() {
	sock := mq.NewSubSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Dial(addr))

	for {
		data, err := sock.Recv()
		if err != nil {
			println(err)
			return
		}

		fmt.Println(string(data))
	}

}

func pub() {
	sock := mq.NewPubSocket()
	sock.SetOption(nano.OptionWriteQLen, 2)
	transport.AddAll(sock)
	// can dial N servers, but sock.Send only pick one server
	// because the send channel is shared within the socket
	dieIfErr(sock.Dial(addr))
	dieIfErr(sock.Dial(addr))

	i := 0
	for {
		i++
		time.Sleep(time.Second * 5)

		body := []byte(fmt.Sprintf("PUB x\nhello %d", i))
		dieIfErr(sock.Send(body))

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
