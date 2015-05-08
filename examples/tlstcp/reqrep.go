// TODO still NOT workable
package main

import (
	"crypto/tls"

	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/reqrep"
	"github.com/funkygao/nano/transport"
	"log"
	"os"
	"strings"
	"time"
)

var (
	addr     = "tls+tcp://127.0.0.1:1234"
	tlscfg   tls.Config
	certFile = "a.cert"
	keyFile  = "a.key"
)

func init() {
	nano.Debug = false

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	dieIfErr(err)
	tlscfg.Certificates = make([]tls.Certificate, 0, 1)
	tlscfg.Certificates = append(tlscfg.Certificates, cert)
}

func dieIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func request() {
	sock := reqrep.NewReqSocket()
	transport.AddAll(sock)
	dieIfErr(sock.DialOptions(addr, map[string]interface{}{nano.OptionTlsConfig: &tlscfg}))
	dieIfErr(sock.SetOption(nano.OptionSendDeadline, time.Second))

	for {
		msg := strings.Repeat("X", 10)
		log.Printf("sending %s", msg)
		err := sock.Send([]byte(msg))
		dieIfErr(err)

		_, err = sock.Recv()
		dieIfErr(err)

		//time.Sleep(time.Second)
	}

	dieIfErr(sock.Close())
}

func reply() {
	sock := reqrep.NewRepSocket()
	transport.AddAll(sock)
	dieIfErr(sock.ListenOptions(addr, map[string]interface{}{nano.OptionTlsConfig: &tlscfg}))
	log.Printf("listening on %s", addr)

	for {
		data, err := sock.Recv()
		dieIfErr(err)

		log.Println(string(data))

		dieIfErr(sock.Send([]byte("world")))
	}

}

func usage() {
	log.Printf("Usage: %s <rep|req>", os.Args[0])
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
