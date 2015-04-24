package nano

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// mkTimer creates a timer based upon a duration.  If however
// a zero valued duration is passed, then a nil channel is passed
// i.e. never selectable.  This allows the output to be readily used
// with deadlines in network connections, etc.
func mkTimer(deadline time.Duration) <-chan time.Time {

	if deadline == 0 {
		return nil
	}

	return time.After(deadline)
}

var debug = true

func debugf(format string, args ...interface{}) {
	if debug {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "<?>"
			line = 0
		} else {
			if i := strings.LastIndex(file, "/"); i >= 0 {
				file = file[i+1:]
			}
		}
		fmt.Printf("DEBUG: %s:%d [%s]: %s\n", file, line,
			time.Now().String(), fmt.Sprintf(format, args...))
	}
}

func DrainChannel(ch chan<- *Message, expire time.Time) bool {
	var dur time.Duration = time.Millisecond * 10

	for {
		if len(ch) == 0 {
			return true
		}
		now := time.Now()
		if now.After(expire) {
			return false
		}
		// We sleep the lesser of the remaining time, or
		// 10 milliseconds.  This polling is kind of suboptimal for
		// draining, but its far far less complicated than trying to
		// arrange special messages to force notification, etc.
		dur = expire.Sub(now)
		if dur > time.Millisecond*10 {
			dur = time.Millisecond * 10
		}
		time.Sleep(dur)
	}
}
