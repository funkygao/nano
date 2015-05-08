// +build !go1.3

package nano

import (
	"bytes"
)

func BufferPoolGet() *bytes.Buffer {
	return &bytes.Buffer{}
}

func BufferPoolPut(b *bytes.Buffer) {}
