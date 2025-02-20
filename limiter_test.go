package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	fakeclock := newFakeClock(time.Now())
	clock = fakeclock
	t.Cleanup(func() { clock = &pkgclock{} })

	const nt = 5
	l := New(nt, time.Minute)
	for range nt {
		if err := l.Acquire(t.Context()); err != nil {
			t.Fatalf("Unexpected error on Acquire() - %s", err)
		}
	}

	if fakeclock.afterCalled {
		t.Errorf("The limiter should not have blocked but it did")
	}
	fakeclock.afterCalled = false

	// Lock one more time, which will cause the limiter to block
	if err := l.Acquire(t.Context()); err != nil {
		t.Fatalf("Unexpected error on Acquire() - %s", err)
	}

	if !fakeclock.afterCalled {
		t.Errorf("The limiter should have blocked but it did not")
	}
}

func TestContextCancel(t *testing.T) {
	l := New(1, 15*time.Second)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	// Consume the only available token.
	if got := l.Acquire(ctx); got != nil {
		t.Errorf("Expected no error but got %s", got)
	}

	// Attempt to do more work, this will block
	if got := l.Acquire(ctx); got == nil || got.Error() != "context canceled" {
		t.Errorf("Expected context canceled error, got %s", got)
	}
}
