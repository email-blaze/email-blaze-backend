package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
}

type rateLimiterEntry struct {
	limiter *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
	go rl.cleanupRoutine()
	return rl
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.limiters[key]
	if !exists {
		entry = &rateLimiterEntry{
			limiter: rate.NewLimiter(r.limit, r.burst),
			lastSeen: time.Now(),
		}
		r.limiters[key] = entry
	} else {
		entry.lastSeen = time.Now()
	}

	return entry.limiter.Allow()
}

func (r *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		r.cleanup()
	}
}

func (r *RateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, entry := range r.limiters {
		if time.Since(entry.lastSeen) > 1*time.Hour {
			delete(r.limiters, key)
		}
	}
}