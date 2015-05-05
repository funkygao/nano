package nano

import (
	"math/rand"
	"time"
)

func init() {
	pipes.byid = make(map[uint32]*pipeEndpoint)
	pipes.nextid = uint32(rand.NewSource(time.Now().UnixNano()).Int63())

	go monitorMessagePool()
}
