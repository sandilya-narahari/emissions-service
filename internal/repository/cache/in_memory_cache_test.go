package cache_test

import (
	"testing"
	"time"

	"emissions-cache-service/internal/repository/cache"
)

func TestCacheSetGet(t *testing.T) {
	ttl := 2 * time.Second
	cleanup := 1 * time.Second
	cacheRepo := cache.NewInMemoryCache(ttl, cleanup, 0)

	key := "test-key"
	value := "test-value"

	// Set a value without priority.
	cacheRepo.Set(key, value, false)
	if v, found := cacheRepo.Get(key); !found {
		t.Errorf("Expected to find key %s in cache", key)
	} else if v != value {
		t.Errorf("Expected value %v, got %v", value, v)
	}

	// Wait for expiration.
	time.Sleep(3 * time.Second)
	if _, found := cacheRepo.Get(key); found {
		t.Errorf("Expected key %s to be expired", key)
	}

	// Test priority: set with isPriority true, so it never expires.
	priorityKey := "priority-key"
	cacheRepo.Set(priorityKey, value, true)
	time.Sleep(3 * time.Second)
	if v, found := cacheRepo.Get(priorityKey); !found {
		t.Errorf("Expected to find priority key %s in cache", priorityKey)
	} else if v != value {
		t.Errorf("Expected value %v, got %v", value, v)
	}
}

func TestCacheEviction(t *testing.T) {
	ttl := 1 * time.Second
	cleanup := 500 * time.Millisecond
	cacheRepo := cache.NewInMemoryCache(ttl, cleanup, 0)

	key := "test-key"
	value := "test-value"

	cacheRepo.Set(key, value, false)

	// Ensure value is in cache initially
	if _, found := cacheRepo.Get(key); !found {
		t.Errorf("Expected key %s in cache", key)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Value should be expired
	if _, found := cacheRepo.Get(key); found {
		t.Errorf("Expected key %s to be expired", key)
	}
}

func TestCacheTTLVariants(t *testing.T) {
	cacheRepo := cache.NewInMemoryCache(2*time.Second, 1*time.Second, 0)

	// Regular TTL
	cacheRepo.Set("short-lived", "value", false)
	// Permanent (priority) cache entry
	cacheRepo.Set("permanent", "value", true)

	time.Sleep(3 * time.Second)

	// Short-lived entry should be expired
	if _, found := cacheRepo.Get("short-lived"); found {
		t.Errorf("Expected key 'short-lived' to expire")
	}

	// Priority entry should still exist
	if _, found := cacheRepo.Get("permanent"); !found {
		t.Errorf("Expected key 'permanent' to persist")
	}
}
