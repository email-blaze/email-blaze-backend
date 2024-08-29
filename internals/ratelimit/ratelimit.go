package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	limiter, exists := r.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(r.limit, r.burst)
		r.limiters[key] = limiter
	}
	r.mu.Unlock()

	return limiter.Allow()
}