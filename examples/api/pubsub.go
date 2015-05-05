package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/api"
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
	s, e := api.NewPubSocket()
	dieIfErr(e)
	dieIfErr(s.Bind(addr))

	body := []byte(strings.Repeat("X", 100))
	for {
		n, err := s.Send(body, 0)
		dieIfErr(err)

		fmt.Println(n)
	}

}

func doSub() {
	s, e := api.NewSubSocket()
	dieIfErr(e)
	dieIfErr(s.Connect(addr))
	dieIfErr(s.Subscribe(""))

	var i int64
	for {
		data, err := s.Recv()
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
