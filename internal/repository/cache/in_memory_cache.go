package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// EmissionsCache is a concurrency-safe in-memory cache for emissions data.
// NOTE: This implementation uses an in-memory cache which is suitable for single-instance deployments.
// For a distributed system, replacing this with a Redis-backed cache.
type EmissionsCache struct {
	store *cache.Cache
}

// NewInMemoryCache creates a new in-memory cache with a default TTL and cleanup interval.
func NewInMemoryCache(defaultTTL, cleanupInterval time.Duration, _ int) *EmissionsCache {
	c := cache.New(defaultTTL, cleanupInterval)
	return &EmissionsCache{
		store: c,
	}
}

// Set stores a value in the cache.
// If isPriority is true, the value never expires.
func (ec *EmissionsCache) Set(key string, value interface{}, isPriority bool) {
	ttl := cache.DefaultExpiration
	if isPriority {
		ttl = cache.NoExpiration
	}
	ec.store.Set(key, value, ttl)
}

// Get retrieves a value from the cache.
func (ec *EmissionsCache) Get(key string) (interface{}, bool) {
	return ec.store.Get(key)
}
