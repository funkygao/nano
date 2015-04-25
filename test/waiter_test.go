package test

import (
	"testing"
	"time"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
)

func TestWaiterRelTimeout(t *testing.T) {
	var w nano.Waiter
	w.Init()
	w.Add()
	w.Add()
	assert.Equal(t, false, w.WaitRelTimeout(time.Microsecond))
}

func TestWaiterAbsTimeout(t *testing.T) {
	var w nano.Waiter
	w.Init()
	w.Add()
	when := time.Now().Add(time.Microsecond)
	assert.Equal(t, false, w.WaitAbsTimeout(when))
}

func TestWaiterAllDone(t *testing.T) {
	var w nano.Waiter
	w.Init()
	w.Add()
	w.Add()
	w.Done()
	w.Done()
	t1 := time.Now()
	w.Wait()
	assert.Equal(t, true, time.Since(t1) < time.Second)
}
