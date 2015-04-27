package nano

import (
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

func StripScheme(t Transport, addr string) (string, error) {
	s := t.Scheme() + "://"
	if !strings.HasPrefix(addr, s) {
		return addr, ErrBadTran
	}
	return addr[len(s):], nil
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

func ProtocolName(number uint16) string {
	names := map[uint16]string{
		ProtoPair:       "pair",
		ProtoPub:        "pub",
		ProtoSub:        "sub",
		ProtoReq:        "req",
		ProtoRep:        "rep",
		ProtoPush:       "push",
		ProtoPull:       "pull",
		ProtoSurveyor:   "surveyor",
		ProtoRespondent: "respondent",
		ProtoBus:        "bus"}
	return names[number]
}

// ValidPeers returns true if the two sockets are capable of
// peering to one another.  For example, REQ can peer with REP,
// but not with BUS.
func ValidPeers(p1, p2 Protocol) bool {
	if p1.Number() != p2.PeerNumber() {
		return false
	}
	if p2.Number() != p1.PeerNumber() {
		return false
	}
	return true
}

// NullRecv simply loops, receiving and discarding messages, until the
// Endpoint returns back a nil message.  This allows the Endpoint to notice
// a dropped connection.  It is intended for use by Protocols that are write
// only -- it lets them become aware of a loss of connectivity even when they
// have no data to send.
func NullRecv(ep Endpoint) {
	for {
		if m := ep.RecvMsg(); m == nil {
			return
		} else {
			m.Free()
		}
	}
}
