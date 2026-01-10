package axios4go

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		CreatedAt:  time.Now(),
	}

	// Test Set and Get
	cache.Set("test-key", entry, 1*time.Minute)

	retrieved := cache.Get("test-key")
	if retrieved == nil {
		t.Fatal("Expected to get cached entry, got nil")
	}

	if string(retrieved.Body) != string(entry.Body) {
		t.Errorf("Expected body %s, got %s", entry.Body, retrieved.Body)
	}

	if retrieved.StatusCode != entry.StatusCode {
		t.Errorf("Expected status code %d, got %d", entry.StatusCode, retrieved.StatusCode)
	}
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	retrieved := cache.Get("non-existent-key")
	if retrieved != nil {
		t.Error("Expected nil for non-existent key")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(&MemoryCacheOptions{
		CleanupInterval: 100 * time.Millisecond,
	})
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	// Set with short TTL
	cache.Set("expiring-key", entry, 50*time.Millisecond)

	// Should be available immediately
	retrieved := cache.Get("expiring-key")
	if retrieved == nil {
		t.Fatal("Expected to get cached entry before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	retrieved = cache.Get("expiring-key")
	if retrieved != nil {
		t.Error("Expected nil for expired entry")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	cache.Set("delete-key", entry, 1*time.Minute)

	// Verify it exists
	if cache.Get("delete-key") == nil {
		t.Fatal("Expected entry to exist before delete")
	}

	// Delete
	cache.Delete("delete-key")

	// Verify it's gone
	if cache.Get("delete-key") != nil {
		t.Error("Expected entry to be deleted")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	// Add multiple entries
	cache.Set("key1", entry, 1*time.Minute)
	cache.Set("key2", entry, 1*time.Minute)
	cache.Set("key3", entry, 1*time.Minute)

	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}

	// Clear
	cache.Clear()

	stats = cache.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.Size)
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	cache.Set("stats-key", entry, 1*time.Minute)

	// Hit
	cache.Get("stats-key")
	cache.Get("stats-key")

	// Miss
	cache.Get("non-existent")

	stats := cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
}

func TestMemoryCache_MaxSize(t *testing.T) {
	cache := NewMemoryCache(&MemoryCacheOptions{
		MaxSize: 2,
	})
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	cache.Set("key1", entry, 1*time.Minute)
	time.Sleep(10 * time.Millisecond) // Ensure different CreatedAt
	cache.Set("key2", entry, 1*time.Minute)
	time.Sleep(10 * time.Millisecond)
	cache.Set("key3", entry, 1*time.Minute) // Should evict key1

	stats := cache.Stats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2 (max), got %d", stats.Size)
	}

	// key1 should be evicted (oldest)
	if cache.Get("key1") != nil {
		t.Error("Expected key1 to be evicted")
	}
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set("concurrent-key", entry, 1*time.Minute)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get("concurrent-key")
		}()
	}

	wg.Wait()

	// Should not panic or have race conditions
	retrieved := cache.Get("concurrent-key")
	if retrieved == nil {
		t.Error("Expected to get cached entry after concurrent operations")
	}
}

func TestMemoryCache_ZeroTTL(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	entry := &CacheEntry{
		Body:       []byte(`{"message":"test"}`),
		StatusCode: 200,
		Headers:    http.Header{},
		CreatedAt:  time.Now(),
	}

	// Set with zero TTL should not store
	cache.Set("zero-ttl", entry, 0)

	if cache.Get("zero-ttl") != nil {
		t.Error("Expected nil for zero TTL entry")
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	// Not expired
	entry := &CacheEntry{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if entry.IsExpired() {
		t.Error("Expected entry to not be expired")
	}

	// Expired
	expiredEntry := &CacheEntry{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !expiredEntry.IsExpired() {
		t.Error("Expected entry to be expired")
	}
}

func TestCacheEnabled(t *testing.T) {
	opts := CacheEnabled(5 * time.Minute)

	if opts.Enabled == nil || !*opts.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if opts.TTL != 5*time.Minute {
		t.Errorf("Expected TTL 5m, got %v", opts.TTL)
	}
}

func TestCacheDisabled(t *testing.T) {
	opts := CacheDisabled()

	if opts.Enabled == nil || *opts.Enabled {
		t.Error("Expected Enabled to be false")
	}
}

func TestDefaultCacheKeyFunc(t *testing.T) {
	key := DefaultCacheKeyFunc("GET", "https://api.example.com/users", nil)
	expected := "GET:https://api.example.com/users"

	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestShouldCacheRequest(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	cacheConfig := &CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	}

	t.Run("NoCacheConfig", func(t *testing.T) {
		options := &RequestOptions{Method: "GET"}
		if shouldCacheRequest(nil, options) {
			t.Error("Expected false when no cache config")
		}
	})

	t.Run("ExplicitlyEnabled", func(t *testing.T) {
		options := &RequestOptions{
			Method: "GET",
			Cache:  CacheEnabled(time.Minute),
		}
		if !shouldCacheRequest(cacheConfig, options) {
			t.Error("Expected true when explicitly enabled")
		}
	})

	t.Run("ExplicitlyDisabled", func(t *testing.T) {
		options := &RequestOptions{
			Method: "GET",
			Cache:  CacheDisabled(),
		}
		if shouldCacheRequest(cacheConfig, options) {
			t.Error("Expected false when explicitly disabled")
		}
	})

	t.Run("NotEnabled", func(t *testing.T) {
		options := &RequestOptions{Method: "GET"}
		if shouldCacheRequest(cacheConfig, options) {
			t.Error("Expected false when not explicitly enabled (opt-in model)")
		}
	})

	t.Run("NonCacheableMethod", func(t *testing.T) {
		options := &RequestOptions{
			Method: "POST",
			Cache:  CacheEnabled(time.Minute),
		}
		if shouldCacheRequest(cacheConfig, options) {
			t.Error("Expected false for POST method")
		}
	})
}

func TestIsMethodCacheable(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	t.Run("DefaultMethods", func(t *testing.T) {
		config := &CacheConfig{Cache: cache}

		if !isMethodCacheable(config, "GET") {
			t.Error("GET should be cacheable by default")
		}
		if !isMethodCacheable(config, "get") {
			t.Error("get (lowercase) should be cacheable")
		}
		if isMethodCacheable(config, "POST") {
			t.Error("POST should not be cacheable by default")
		}
	})

	t.Run("CustomMethods", func(t *testing.T) {
		config := &CacheConfig{
			Cache:            cache,
			CacheableMethods: []string{"GET", "HEAD"},
		}

		if !isMethodCacheable(config, "HEAD") {
			t.Error("HEAD should be cacheable when configured")
		}
		if isMethodCacheable(config, "POST") {
			t.Error("POST should not be cacheable")
		}
	})
}

func TestGenerateCacheKey(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	t.Run("DefaultKey", func(t *testing.T) {
		config := &CacheConfig{Cache: cache}
		options := &RequestOptions{Method: "GET"}

		key := generateCacheKey(config, options, "https://api.example.com/users")
		if key != "GET:https://api.example.com/users" {
			t.Errorf("Unexpected key: %s", key)
		}
	})

	t.Run("CustomKey", func(t *testing.T) {
		config := &CacheConfig{Cache: cache}
		options := &RequestOptions{
			Method: "GET",
			Cache: &RequestCacheOptions{
				CustomKey: "my-custom-key",
			},
		}

		key := generateCacheKey(config, options, "https://api.example.com/users")
		if key != "my-custom-key" {
			t.Errorf("Expected custom key, got: %s", key)
		}
	})

	t.Run("CustomKeyFunc", func(t *testing.T) {
		config := &CacheConfig{
			Cache: cache,
			KeyFunc: func(method, fullURL string, headers map[string]string) string {
				return "custom-func:" + method + ":" + fullURL
			},
		}
		options := &RequestOptions{Method: "GET"}

		key := generateCacheKey(config, options, "https://api.example.com/users")
		if key != "custom-func:GET:https://api.example.com/users" {
			t.Errorf("Unexpected key: %s", key)
		}
	})
}

func TestGetCacheTTL(t *testing.T) {
	cache := NewMemoryCache(nil)
	defer cache.Close()

	config := &CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	}

	t.Run("DefaultTTL", func(t *testing.T) {
		options := &RequestOptions{}
		ttl := getCacheTTL(config, options)
		if ttl != 5*time.Minute {
			t.Errorf("Expected default TTL 5m, got %v", ttl)
		}
	})

	t.Run("PerRequestTTL", func(t *testing.T) {
		options := &RequestOptions{
			Cache: &RequestCacheOptions{
				TTL: 10 * time.Minute,
			},
		}
		ttl := getCacheTTL(config, options)
		if ttl != 10*time.Minute {
			t.Errorf("Expected per-request TTL 10m, got %v", ttl)
		}
	})
}

func TestShouldForceRefresh(t *testing.T) {
	t.Run("NoCache", func(t *testing.T) {
		options := &RequestOptions{}
		if shouldForceRefresh(options) {
			t.Error("Expected false when no cache options")
		}
	})

	t.Run("ForceRefreshTrue", func(t *testing.T) {
		options := &RequestOptions{
			Cache: &RequestCacheOptions{
				ForceRefresh: true,
			},
		}
		if !shouldForceRefresh(options) {
			t.Error("Expected true when ForceRefresh is true")
		}
	})

	t.Run("ForceRefreshFalse", func(t *testing.T) {
		options := &RequestOptions{
			Cache: &RequestCacheOptions{
				ForceRefresh: false,
			},
		}
		if shouldForceRefresh(options) {
			t.Error("Expected false when ForceRefresh is false")
		}
	})
}
