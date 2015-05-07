package main

import (
	"log"

	"github.com/funkygao/nano/protocol/pipeline"
	"github.com/funkygao/nano/protocol/pubsub"
	"github.com/funkygao/nano/transport"
)

type pubnub struct {
	ch chan []byte
}

func main() {
	pn := &pubnub{
		ch: make(chan []byte, 128),
	}

	go recvPub(pn)

	sock := pubsub.NewPubSocket()
	transport.AddAll(sock)
	addr := "tcp://:1910"
	if err := sock.Listen(addr); err != nil {
		panic(err)
	}
	log.Printf("pub listening on %s", addr)

	for data := range pn.ch {
		err := sock.Send(data)
		if err != nil {
			panic(err)
		}

	}

}

func recvPub(pn *pubnub) {
	sock := pipeline.NewPullSocket()
	transport.AddAll(sock)
	addr := "tcp://:1900"
	if err := sock.Listen(addr); err != nil {
		panic(err)
	}
	log.Printf("pipeline listening on %s", addr)

	for {
		data, err := sock.Recv()
		if err != nil {
			panic(err)
		}

		pn.ch <- data
	}

}
