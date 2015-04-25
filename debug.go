package nano

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var (
	Debug = true
)

func debugf(format string, args ...interface{}) {
	if Debug {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "<?>"
			line = 0
		} else {
			if i := strings.LastIndex(file, "/"); i >= 0 {
				file = file[i+1:]
			}
		}
		t := time.Now()
		hour, min, sec := t.Clock()
		nanosec := t.Nanosecond() / 1e3
		fmt.Printf("DEBUG: [%d:%d:%d.%04d] %s:%d: %s\n",
			hour, min, sec, nanosec,
			file, line,
			fmt.Sprintf(format, args...))
	}
}
