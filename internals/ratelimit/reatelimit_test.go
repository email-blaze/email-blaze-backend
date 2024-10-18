package ratelimit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	t.Run("Allow within limit", func(t *testing.T) {
		rl := NewRateLimiter(10, 1)
		assert.True(t, rl.Allow("test_key"))
	})

	t.Run("Exceed rate limit", func(t *testing.T) {
		rl := NewRateLimiter(1, 1)
		assert.True(t, rl.Allow("test_key"))
		assert.False(t, rl.Allow("test_key"))
	})

	t.Run("Different keys", func(t *testing.T) {
		rl := NewRateLimiter(1, 1)
		assert.True(t, rl.Allow("key1"))
		assert.True(t, rl.Allow("key2"))
	})

	t.Run("Cleanup", func(t *testing.T) {
		rl := NewRateLimiter(1, 1)
		rl.Allow("test_key")

		// Simulate time passing
		rl.limiters["test_key"].lastSeen = time.Now().Add(-2 * time.Hour)

		rl.cleanup()
		assert.Empty(t, rl.limiters)
	})
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(100, 10)
	concurrentRequests := 1000
	allowedCount := 0

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("test_key") {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, allowedCount, 110) // Allow for some flexibility due to timing
	assert.GreaterOrEqual(t, allowedCount, 90)
}
