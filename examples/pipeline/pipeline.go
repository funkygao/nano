// a one-way pipe with Device(intermediary node) feature.
package main

import (
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pipeline"
	"github.com/funkygao/nano/transport"
	"log"
	"os"
	"strings"
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
	log.Printf("forward listen on: %s", fromAddr)

	pushSock := pipeline.NewPushSocket()
	transport.AddAll(pushSock)
	dieIfErr(pushSock.Dial(toAddr))
	log.Printf("forward dial to: %s", toAddr)

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
	log.Printf("pull listen on: %s", addr)

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
	log.Printf("push dial to: %s", addr)

	// Message routing
	msg := strings.Repeat("X", 10)
	for {
		dieIfErr(sock.Send([]byte(msg)))
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
