package main

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/reqrep"
	"github.com/funkygao/nano/transport"
	"log"
	"os"
	"strings"
	"time"
)

func init() {
	nano.Debug = true
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func request(addr string) {
	sock := reqrep.NewReqSocket()
	transport.AddAll(sock)
	dieIfErr(sock.SetOption(nano.OptionReadQLen, 4<<10)) // must be before Dial
	dieIfErr(sock.SetOption(nano.OptionWriteQLen, 4<<10))
	dieIfErr(sock.Dial(addr))
	dieIfErr(sock.SetOption(nano.OptionSendDeadline, time.Second))

	for i := 0; i < 2; i++ {
		err := sock.Send([]byte(strings.Repeat("X", 10)))
		dieIfErr(err)

		msg, err := sock.Recv()
		dieIfErr(err)
		log.Println(i, string(msg))

		time.Sleep(time.Second)
	}

	dieIfErr(sock.Close())
}

func reply(addr string) {
	sock := reqrep.NewRepSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))
	log.Printf("listening on %s", addr)

	for {
		data, err := sock.Recv()
		dieIfErr(err)

		log.Println(string(data))

		dieIfErr(sock.Send([]byte("world")))
	}

}

func usage() {
	log.Printf("Usage: %s <rep|req> <url>", os.Args[0])
	log.Println("url example: tcp://127.0.0.1:1234  ipc://x.sock  inproc://test  tls+tcp://127.0.0.1:1234")
	os.Exit(0)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}

	switch os.Args[1] {
	case "rep":
		reply(os.Args[2])
	case "req":
		request(os.Args[2])
	default:
		usage()
	}
}
