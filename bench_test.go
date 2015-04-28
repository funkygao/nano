package nano

import (
	"strings"
	"testing"
)

func BenchmarkStringIndex(b *testing.B) {
	addr := "tcp://eth0;124.56.74.124:1990"
	for i := 0; i < b.N; i++ {
		strings.Index(addr, "://")
	}
}
