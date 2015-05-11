// a one-way pipe with Device(intermediary node) feature.
//
// Usage:
// 3 nodes required:
// node1: $ pipeline pull tcp://127.0.0.1:1235
// node2: $ pipeline push tcp://127.0.0.1:1234
// node3: $ pipeline forward tcp://127.0.0.1:1234 tcp://127.0.0.1:1235
//
package main

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pipeline"
	"github.com/funkygao/nano/transport"
	"log"
	"os"
	"strings"
	"time"
)

func init() {
	nano.Debug = false
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func doForward(fromAddr, toAddr string) {
	pullSock := pipeline.NewPullSocket()
	transport.AddAll(pullSock)
	dieIfErr(pullSock.Listen(fromAddr))

	pushSock := pipeline.NewPushSocket()
	transport.AddAll(pushSock)
	dieIfErr(pushSock.Dial(toAddr))

	dieIfErr(nano.Device(pullSock, pushSock))
	select {}

	dieIfErr(pushSock.Close())
	dieIfErr(pullSock.Close())
}

func doPull(addr string) {
	// Topology establishment
	sock := pipeline.NewPullSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Listen(addr))

	// Message routing
	for {
		data, err := sock.Recv()
		dieIfErr(err)

		log.Println(string(data))
	}

	dieIfErr(sock.Close())

}

func doPush(addr string) {
	// Topology establishment
	sock := pipeline.NewPushSocket()
	transport.AddAll(sock)
	dieIfErr(sock.Dial(addr))

	// Message routing
	for {
		dieIfErr(sock.Send([]byte(strings.Repeat("X", 10))))
		time.Sleep(time.Second)
	}

	dieIfErr(sock.Close())
}

func usage() {
	log.Printf("Usage: %s <push|pull|forward> <url> <url2>", os.Args[0])
	os.Exit(0)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	switch os.Args[1] {
	case "push":
		doPush(os.Args[2])
	case "pull":
		doPull(os.Args[2])
	case "forward":
		if len(os.Args) != 4 {
			usage()
		}

		doForward(os.Args[2], os.Args[3])
	default:
		usage()
	}
}
