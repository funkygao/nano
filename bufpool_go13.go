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

func BufferPoolGet() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func BufferPoolPut(b *bytes.Buffer) {
	bufferPool.Put(b)
}
