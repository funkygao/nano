// benchmark throughput of push-pull
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/funkygao/golib/dashboard"
	"github.com/funkygao/golib/stress"
	"github.com/funkygao/golib/wg"
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pair"
	"github.com/funkygao/nano/transport"
	"github.com/pkg/browser"
)

var (
	addr             = "tcp://127.0.0.1:1234"
	bodySize         = 300
	dashboardRefresh = 10
)

type metricsRecv struct {
	last  int
	total int64
}

func (this *metricsRecv) Data() int {
	total := int(atomic.LoadInt64(&this.total))
	r := (total - this.last) / dashboardRefresh
	this.last = total
	return r
}

func init() {
	nano.Debug = false
}

func usage() {
	fmt.Printf("Usage: %s <server|client>\n", os.Args[0])
	os.Exit(0)
}

func runServer(seq int) {
	db := dashboard.New("qps", dashboardRefresh)
	g := db.AddGraph("qps")
	metrics := &metricsRecv{}
	g.AddLine("recv", metrics)
	go db.Launch(":8000")
	browser.OpenURL("http://localhost:8000/")

	sock, err := pair.NewSocket()
	if err != nil {
		panic(err)
	}

	transport.AddAll(sock)
	if err := sock.Listen(addr); err != nil {
		panic(err)
	}

	var body = []byte(strings.Repeat("x", bodySize))
	for {
		if _, err := sock.Recv(); err != nil {
			log.Println(err)
			break
		}
		atomic.AddInt64(&metrics.total, 1)

		if err := sock.Send(body); err != nil {
			log.Println(err)
			break
		}
		atomic.AddInt64(&metrics.total, 1)
	}

	if err := sock.Close(); err != nil {
		log.Println(err)
	}

}

func runClient(seq int) {
	sock, err := pair.NewSocket()
	if err != nil {
		panic(err)
	}

	transport.AddAll(sock)
	if err := sock.Dial(addr); err != nil {
		panic(err)
	}

	body := []byte(strings.Repeat("X", bodySize))
	var w wg.WaitGroupWrapper
	w.Wrap(func() {
		const x = 10000
		for i := 0; i < x; i++ {
			if err := sock.Send(body); err != nil {
				log.Println(err)
				break
			}
		}

		stress.IncCounter("qps", x)

		sock.Close()
	})
	w.Wrap(func() {
		for {
			_, err := sock.Recv()
			if err != nil {
				//log.Println(err)
				break
			}
			stress.IncCounter("qps", 1)
		}
	})
	w.Wait()
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {
	case "client":
		stress.RunStress(runClient)

	case "server":
		runServer(0)

	default:
		usage()
	}

	select {}

}
