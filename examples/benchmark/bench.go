// benchmark throughput of push-pull
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/funkygao/golib/dashboard"
	"github.com/funkygao/golib/stress"
	"github.com/funkygao/nano"
	"github.com/funkygao/nano/protocol/pipeline"
	"github.com/funkygao/nano/transport"
	"github.com/pkg/browser"
)

var (
	//addr             = "ipc://xx"
	addr             = "tcp://127.0.0.1:1234"
	bodySize         = 1000
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
	fmt.Printf("Usage: %s <push|pull>\n", os.Args[0])
	os.Exit(0)
}

func runPull(seq int) {
	db := dashboard.New("qps", dashboardRefresh)
	g := db.AddGraph("qps")
	metrics := &metricsRecv{}
	g.AddLine("recv", metrics)
	go db.Launch(":8000")
	browser.OpenURL("http://localhost:8000/")

	sock := pipeline.NewPullSocket()
	transport.AddAll(sock)
	if err := sock.Listen(addr); err != nil {
		panic(err)
	}

	var buf *bytes.Buffer
	for {
		buf = nano.BufferPoolGet()
		_, err := sock.XRecv(buf)
		nano.BufferPoolPut(buf)
		if err != nil {
			fmt.Println(err)
			break
		}

		atomic.AddInt64(&metrics.total, 1)
	}

	if err := sock.Close(); err != nil {
		fmt.Println(err)
	}

}

func runPush(seq int) {
	sock := pipeline.NewPushSocket()
	transport.AddAll(sock)
	if err := sock.Dial(addr); err != nil {
		panic(err)
	}

	body := []byte(strings.Repeat("X", bodySize))
	for i := 0; i < 100000; i++ {
		sock.Send(body)
		stress.IncCounter("qps", 1)
	}
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {
	case "push":
		stress.RunStress(runPush)

	case "pull":
		runPull(0)

	default:
		usage()
	}

	select {}

}
