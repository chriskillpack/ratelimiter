package ratelimiter

import "time"

type fakeclock struct {
	nowCalled, afterCalled bool
	fakeNow                time.Time
}

func newFakeClock(now time.Time) *fakeclock {
	return &fakeclock{fakeNow: now}
}

func (fc *fakeclock) Now() time.Time {
	fc.nowCalled = true
	return fc.fakeNow
}

func (fc *fakeclock) After(d time.Duration) <-chan time.Time {
	fc.afterCalled = true
	fc.fakeNow = fc.fakeNow.Add(d)
	c := make(chan time.Time)
	go func() {
		c <- fc.fakeNow
	}()
	return c
}

func (fc *fakeclock) Advance(d time.Duration) time.Time {
	fc.fakeNow = fc.fakeNow.Add(d)
	return fc.fakeNow
}
