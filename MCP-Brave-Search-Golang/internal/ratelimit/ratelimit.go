package ratelimit

import (
	"errors"
	"sync"
	"time"
)

// RateLimits defines the rate limits
type RateLimits struct {
	PerSecond int
	PerMonth  int
}

// RateLimiter manages rate limiting for API requests
type RateLimiter struct {
	limits      RateLimits
	requestCount struct {
		second int
		month  int
	}
	lastReset time.Time
	mu        sync.Mutex
}

// ErrRateLimitExceeded is returned when the rate limit is exceeded
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// NewRateLimiter creates a new rate limiter with the given limits
func NewRateLimiter(limits RateLimits) *RateLimiter {
	return &RateLimiter{
		limits:    limits,
		lastReset: time.Now(),
	}
}

// CheckLimit checks if the request is within rate limits and increments counters
func (r *RateLimiter) CheckLimit() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	
	// Reset second counter if it's been more than a second
	if now.Sub(r.lastReset) > time.Second {
		r.requestCount.second = 0
		r.lastReset = now
	}

	// Check if we're over limits
	if r.requestCount.second >= r.limits.PerSecond ||
		r.requestCount.month >= r.limits.PerMonth {
		return ErrRateLimitExceeded
	}

	// Increment counters
	r.requestCount.second++
	r.requestCount.month++

	return nil
}

// ResetMonthlyCounter resets the monthly counter
// This should be called on a schedule (e.g., first day of month)
func (r *RateLimiter) ResetMonthlyCounter() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requestCount.month = 0
}
