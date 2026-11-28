package api

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestResolveLimiterCapsConcurrency(t *testing.T) {
	const limit = 2
	l := newResolveLimiter(limit)
	var inFlight, peak int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			l.enter()
			defer l.leave()
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				p := atomic.LoadInt32(&peak)
				if cur <= p || atomic.CompareAndSwapInt32(&peak, p, cur) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
		})
	}
	wg.Wait()
	if peak > limit {
		t.Fatalf("peak concurrency %d exceeds limit %d", peak, limit)
	}
}

func TestResolveLimiterWaitedOnlyWhenQueued(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := newResolveLimiter(1)
		if waited := l.enter(); waited {
			t.Fatal("first request must not wait")
		}
		done := make(chan bool, 1)
		go func() { done <- l.enter() }()
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("second request entered while the slot was held")
		default:
		}
		l.leave()
		if waited := <-done; !waited {
			t.Fatal("queued request must report waited")
		}
		l.leave()
	})
}

func TestResolveLimiterResizeAppliesImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := newResolveLimiter(4)
		l.enter()
		l.enter()
		l.resize(1)
		entered := make(chan struct{})
		go func() {
			l.enter()
			close(entered)
		}()
		l.leave()
		synctest.Wait()
		select {
		case <-entered:
			t.Error("enter proceeded while the remaining slot was held")
		default:
		}
		l.leave()
		synctest.Wait()
		select {
		case <-entered:
			l.leave()
		default:
			t.Fatal("queued enter did not proceed after a slot freed")
		}
	})
}

func TestResolveLimiterResizeIgnoresInvalidLimits(t *testing.T) {
	l := newResolveLimiter(2)
	l.resize(0)
	l.resize(-3)
	if l.limit != 2 {
		t.Fatalf("invalid resize changed limit to %d", l.limit)
	}
}
