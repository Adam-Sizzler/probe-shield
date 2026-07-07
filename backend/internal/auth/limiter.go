package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	maxFailures int
	failures    map[string][]time.Time
}

func NewRateLimiter(window time.Duration, maxFailures int) *RateLimiter {
	return &RateLimiter{
		window:      window,
		maxFailures: maxFailures,
		failures:    make(map[string][]time.Time),
	}
}

func (r *RateLimiter) IsBlocked(key string) (bool, time.Duration) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	kept := r.keepRecentLocked(key, now)
	if len(kept) < r.maxFailures {
		return false, 0
	}
	oldest := kept[0]
	retryAfter := r.window - now.Sub(oldest)
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	return true, retryAfter
}

func (r *RateLimiter) RecordFailure(key string) (bool, time.Duration) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	kept := r.keepRecentLocked(key, now)
	kept = append(kept, now)
	r.failures[key] = kept
	if len(kept) < r.maxFailures {
		return false, 0
	}
	oldest := kept[0]
	retryAfter := r.window - now.Sub(oldest)
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	return true, retryAfter
}

func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, key)
}

func (r *RateLimiter) keepRecentLocked(key string, now time.Time) []time.Time {
	cutoff := now.Add(-r.window)
	items := r.failures[key]
	kept := items[:0]
	for _, item := range items {
		if item.After(cutoff) {
			kept = append(kept, item)
		}
	}
	if len(kept) == 0 {
		delete(r.failures, key)
		return nil
	}
	r.failures[key] = kept
	return kept
}
