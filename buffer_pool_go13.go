// +build go1.3

package nano

import (
	"bytes"
	"sync"
)

var bufferPool sync.Pool

func init() {
	bufferPool.New = func() interface{} {
		return &bytes.Buffer{}
	}
}

func bufferPoolGet() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func bufferPoolPut(b *bytes.Buffer) {
	bufferPool.Put(b)
}
