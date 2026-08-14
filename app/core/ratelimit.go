package core

import (
	"sync"
	"time"
)

// RateLimiter is a fixed-window counter keyed by an arbitrary string such as a
// client IP. It is in-memory and per-process, which is sufficient for a single
// instance and is the obvious seam to swap for Redis once the server is
// replicated.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*rateBucket
	now     func() time.Time
}

type rateBucket struct {
	count     int
	expiresAt time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*rateBucket),
		now:     time.Now,
	}
}

// Allow records one attempt for key and reports whether it stays within the
// limit. A key with no bucket, or one whose window has elapsed, starts fresh.
func (r *RateLimiter) Allow(key string) bool {
	if r.limit <= 0 {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	r.sweep(now)

	bucket, ok := r.buckets[key]
	if !ok || now.After(bucket.expiresAt) {
		r.buckets[key] = &rateBucket{count: 1, expiresAt: now.Add(r.window)}
		return true
	}

	bucket.count++
	return bucket.count <= r.limit
}

// Reset clears a key, letting a successful login forgive earlier failures.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
}

// sweep drops expired buckets so the map cannot grow without bound under a
// spray of distinct keys. Callers already hold the mutex.
func (r *RateLimiter) sweep(now time.Time) {
	for key, bucket := range r.buckets {
		if now.After(bucket.expiresAt) {
			delete(r.buckets, key)
		}
	}
}
