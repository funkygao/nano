package nano

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkStringIndex(b *testing.B) {
	addr := "tcp://eth0;124.56.74.124:1990"
	for i := 0; i < b.N; i++ {
		strings.Index(addr, "://")
	}
}

func BenchmarkMakeTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mkTimer(time.Nanosecond)
	}
}

func BenchmarkCAS(b *testing.B) {
	var v int32
	for i := 0; i < b.N; i++ {
		atomic.CompareAndSwapInt32(&v, 0, 1)
	}
}
