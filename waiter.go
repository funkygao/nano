package nano

import (
	"sync"
	"time"
)

// condTimed is a condition variable (ala sync.Cond) but inclues a timeout.
type condTimed struct {
	sync.Cond
}

// WaitRelTimeout is like Wait, but it times out.  The fact that
// it timed out can be determined by checking the return value.  True
// indicates that it woke up without a timeout (signaled another way),
// whereas false indicates a timeout occurred.
func (this *condTimed) WaitRelTimeout(when time.Duration) bool {
	timer := time.AfterFunc(when, func() {
		this.L.Lock()
		this.Broadcast()
		this.L.Unlock()
	})
	this.Wait()
	return timer.Stop()
}

// WaitAbsTimeout is like WaitRelTimeout, but expires on an absolute time
// instead of a relative one.
func (this *condTimed) WaitAbsTimeout(when time.Time) bool {
	now := time.Now()
	if when.After(now) {
		return this.WaitRelTimeout(when.Sub(now))
	} else {
		return this.WaitRelTimeout(0)
	}
}

// Waiter is a way to wait for completion, but it includes a timeout.  It
// is similar in some respects to sync.WaitGroup.
type Waiter struct {
	cv    condTimed // conditional variable
	count int
	sync.Mutex
}

// Init must be called to initialize the Waiter.
func (this *Waiter) Init() {
	this.cv.L = this
	this.count = 0
}

// Add adds a new go routine/item to wait for. This should be called before
// starting go routines you want to wait for, for example.
// TODO atomic.Add?
func (this *Waiter) Add() {
	this.Lock()
	this.count++
	this.Unlock()
}

// Done is called when the item to wait for is done. There should be a one to
// one correspondance between Add and Done.  When the count drops to zero,
// any callers blocked in Wait() are woken.  If the count drops below zero,
// it panics.
func (this *Waiter) Done() {
	this.Lock()
	this.count--
	if this.count < 0 {
		// should never happen
		panic("wait count dropped < 0")
	}
	if this.count == 0 {
		this.cv.Broadcast()
	}
	this.Unlock()
}

// Wait waits without a timeout.  It only completes when the count drops
// to zero.
func (this *Waiter) Wait() {
	this.Lock()
	for this.count != 0 {
		this.cv.Wait()
	}
	this.Unlock()
}

// WaitRelTimeout waits until either the count drops to zero, or the timeout
// expires.  It returns true if the count is zero, false otherwise.
func (this *Waiter) WaitRelTimeout(d time.Duration) bool {
	this.Lock()
	for this.count != 0 {
		if !this.cv.WaitRelTimeout(d) {
			break
		}
	}
	done := this.count == 0
	this.Unlock()
	return done
}

// WaitAbsTimeout is like WaitRelTimeout, but waits until an absolute time.
func (this *Waiter) WaitAbsTimeout(t time.Time) bool {
	this.Lock()
	for this.count != 0 {
		if !this.cv.WaitAbsTimeout(t) {
			break
		}
	}
	done := this.count == 0
	this.Unlock()
	return done
}
