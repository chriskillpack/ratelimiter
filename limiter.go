// Package ratelimiter provides a concurrency safe rate limiter.
//
// Create a new rate limiter with NewLimiter() providing the limiter with the
// number of attempts allowed over a window of time. Clients should then call
// Acquire() prior to doing a rate limited piece of work. The Limiter will block
// until either sufficient time has passed or the provided Context has been
// closed.
//
// Internally the rate limit uses a simple token bucket approach which is both
// simple and handles average and bursty loads well.
package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// A simple rate limiter that uses the token bucket algorithm.
type Limiter struct {
	mu       sync.Mutex // protect access to lastTime and tokens
	lastTime time.Time
	tokens   int

	window time.Duration
	rate   int
}

// NewLimiter creates a new rate limiter for the given number of tokens
// over the provided time window. E.g. NewLimiter(10, time.Minute) will
// allow 10 units of work to happen over a minute. The limiter is already full
// so the caller can immediately get all
func New(rate int, window time.Duration) *Limiter {
	return &Limiter{
		window:   window,
		rate:     rate,
		lastTime: clock.Now(),
		tokens:   rate,
	}
}

// Acquire returns nil if work can proceed immediately. If the provided context
// is Done Acquire will return context.Err(). If the bucket is empty, Acquire
// will block until at least one unit of work can be executed.
func (l *Limiter) Acquire(ctx context.Context) error {
	for {
		if ok := l.tryAcquire(); ok {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-clock.After(l.window / time.Duration(l.rate)):
			// If tryAcquire() returned false the token bucket is empty.
			// Assuming an even distribution of tokens across the window, wait
			// 1/Nth of the window duration to allow at least one token to
			// accumulate. And then try again.
		}
	}
}

func (l *Limiter) tryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// How much time has elapsed?
	now := clock.Now()
	elapsed := now.Sub(l.lastTime)
	l.lastTime = now

	// Put tokens into the bucket, the number proportional to the duration since
	// last called.
	l.tokens += int(elapsed.Nanoseconds() * int64(l.rate) / l.window.Nanoseconds())
	l.tokens = min(l.tokens, l.rate)

	// If the bucket is exhausted then the caller cannot proceed immediately.
	if l.tokens <= 0 {
		return false
	}

	// Success, remove a token.
	l.tokens--
	return true
}

// clocker defines an interface through which to access time package functions.
// This exists purely for testing. If testing/synctest lands then hopefully
// this dance won't be necessary anymore.
type clocker interface {
	Now() time.Time

	After(d time.Duration) <-chan time.Time
}

// The default implementation of clocker just calls the package level functions
type pkgclock struct{}

func (p *pkgclock) Now() time.Time {
	return time.Now()
}

func (p *pkgclock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// This variable holds the clock implementation that will be used in the
// limiter. It will only be overriden in tests.
var clock clocker = &pkgclock{}
