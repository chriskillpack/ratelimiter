# ratelimiter
A simple Go rate limiter

## Usage

```
import "github.com/chriskillpack/ratelimiter"

l := ratelimiter.New(10, time.Minute)  // A limiter that allows 10 requests a minute

// Block until rate limit under configured amount
if err := l.Acquire(context); err != nil {
    // handle error, only happens when context is cancelled or Done
}
f()
```