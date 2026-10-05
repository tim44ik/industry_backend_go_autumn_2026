package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{clock: clock, rate: ratePerSec, burst: burst, tokens: float64(burst)}
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.burst <= 0 || l.clock == nil {
		return false
	}

	if n <= 0 || n > l.burst {
		return false
	}

	if l.rate > 0 {
		l.add()
	}

	if l.tokens-float64(n) < 0 {
		return false
	}
	l.tokens = l.tokens - float64(n)
	return true
}

func (l *Limiter) add() {
	diff := l.clock.Now().Sub(l.last).Seconds()
	if diff < 0 {
		return
	}

	if l.tokens < float64(l.burst) {
		l.tokens += l.rate * diff
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}

	l.last = l.clock.Now()
}
