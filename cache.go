package axios4go

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Cache defines the interface for cache implementations.
// Users can implement this interface for custom storage backends (Redis, file, etc.)
type Cache interface {
	// Get retrieves a cached response by key. Returns nil if not found or expired.
	Get(key string) *CacheEntry

	// Set stores a response in the cache with the given key and TTL.
	Set(key string, entry *CacheEntry, ttl time.Duration)

	// Delete removes an entry from the cache.
	Delete(key string)

	// Clear removes all entries from the cache.
	Clear()

	// Stats returns cache statistics.
	Stats() CacheStats
}

// CacheEntry represents a cached HTTP response
type CacheEntry struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CacheStats provides cache statistics
type CacheStats struct {
	Hits   int64
	Misses int64
	Size   int64
}

// CacheKeyFunc is a function type for generating cache keys
type CacheKeyFunc func(method, fullURL string, headers map[string]string) string

// CacheConfig holds the global cache configuration for a Client
type CacheConfig struct {
	// Cache is the cache implementation to use
	Cache Cache

	// DefaultTTL is the default TTL for cached responses when per-request TTL is not specified
	// A value of 0 means no default TTL (must be specified per-request)
	DefaultTTL time.Duration

	// KeyFunc is a custom function to generate cache keys
	// If nil, the default key function (Method + URL) is used
	KeyFunc CacheKeyFunc

	// CacheableMethods defines which HTTP methods can be cached
	// If nil, defaults to []string{"GET"}
	CacheableMethods []string
}

// RequestCacheOptions holds per-request cache configuration
type RequestCacheOptions struct {
	// Enabled explicitly enables/disables caching for this request
	// If nil, caching is disabled (opt-in model)
	Enabled *bool

	// TTL sets the TTL for this specific request
	// Overrides the global DefaultTTL if set
	TTL time.Duration

	// ForceRefresh bypasses the cache and fetches fresh data
	// The fresh response will still be cached
	ForceRefresh bool

	// CustomKey allows overriding the cache key for this request
	CustomKey string
}

// Bool is a helper function to create a pointer to a bool value
func Bool(v bool) *bool {
	return &v
}

// CacheEnabled returns a RequestCacheOptions with caching enabled
func CacheEnabled(ttl time.Duration) *RequestCacheOptions {
	enabled := true
	return &RequestCacheOptions{
		Enabled: &enabled,
		TTL:     ttl,
	}
}

// CacheDisabled returns a RequestCacheOptions with caching disabled
func CacheDisabled() *RequestCacheOptions {
	enabled := false
	return &RequestCacheOptions{
		Enabled: &enabled,
	}
}

// DefaultCacheKeyFunc generates a cache key from method and full URL
func DefaultCacheKeyFunc(method, fullURL string, _ map[string]string) string {
	return method + ":" + fullURL
}

// MemoryCache is a thread-safe in-memory cache implementation
type MemoryCache struct {
	entries         map[string]*memoryCacheEntry
	mu              sync.RWMutex
	hits            int64
	misses          int64
	maxSize         int
	cleanupInterval time.Duration
	stopChan        chan struct{}
	stopped         bool
}

type memoryCacheEntry struct {
	entry *CacheEntry
}

// MemoryCacheOptions configures the MemoryCache
type MemoryCacheOptions struct {
	// MaxSize is the maximum number of entries (0 = unlimited)
	MaxSize int

	// CleanupInterval is how often to clean expired entries (default: 5 minutes)
	CleanupInterval time.Duration
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(opts *MemoryCacheOptions) *MemoryCache {
	if opts == nil {
		opts = &MemoryCacheOptions{}
	}

	cleanupInterval := opts.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = 5 * time.Minute
	}

	mc := &MemoryCache{
		entries:         make(map[string]*memoryCacheEntry),
		maxSize:         opts.MaxSize,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}

	// Start background cleanup goroutine
	go mc.cleanupLoop()

	return mc
}

// Get retrieves a cached response by key
func (c *MemoryCache) Get(key string) *CacheEntry {
	c.mu.RLock()
	entry, exists := c.entries[key]
	if !exists {
		c.mu.RUnlock()
		atomic.AddInt64(&c.misses, 1)
		return nil
	}

	// Copy the entry while holding the lock to avoid race conditions
	entryCopy := &CacheEntry{
		Body:       entry.entry.Body,
		StatusCode: entry.entry.StatusCode,
		Headers:    entry.entry.Headers,
		CreatedAt:  entry.entry.CreatedAt,
		ExpiresAt:  entry.entry.ExpiresAt,
	}
	c.mu.RUnlock()

	if entryCopy.IsExpired() {
		// Delete expired entry
		c.Delete(key)
		atomic.AddInt64(&c.misses, 1)
		return nil
	}

	atomic.AddInt64(&c.hits, 1)
	return entryCopy
}

// Set stores a response in the cache
func (c *MemoryCache) Set(key string, entry *CacheEntry, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check max size and evict oldest if necessary
	if c.maxSize > 0 && len(c.entries) >= c.maxSize {
		// Simple eviction: remove first expired or oldest entry
		c.evictOne()
	}

	entry.ExpiresAt = time.Now().Add(ttl)
	c.entries[key] = &memoryCacheEntry{
		entry: entry,
	}
}

// Delete removes an entry from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*memoryCacheEntry)
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	size := int64(len(c.entries))
	c.mu.RUnlock()

	return CacheStats{
		Hits:   atomic.LoadInt64(&c.hits),
		Misses: atomic.LoadInt64(&c.misses),
		Size:   size,
	}
}

// Close stops the background cleanup goroutine
func (c *MemoryCache) Close() {
	c.mu.Lock()
	if !c.stopped {
		c.stopped = true
		close(c.stopChan)
	}
	c.mu.Unlock()
}

// evictOne removes one entry to make room (called with lock held)
func (c *MemoryCache) evictOne() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.entries {
		// First, try to evict expired entries
		if entry.entry.IsExpired() {
			delete(c.entries, key)
			return
		}

		// Track oldest entry
		if first || entry.entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.entry.CreatedAt
			first = false
		}
	}

	// Evict oldest if no expired entries found
	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// cleanupLoop runs periodically to remove expired entries
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopChan:
			return
		}
	}
}

// cleanupExpired removes all expired entries
func (c *MemoryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.entry.IsExpired() {
			delete(c.entries, key)
		}
	}
}

// shouldCacheRequest determines if a request should be cached
func shouldCacheRequest(cacheConfig *CacheConfig, options *RequestOptions) bool {
	// Check if client has cache configured
	if cacheConfig == nil || cacheConfig.Cache == nil {
		return false
	}

	// Check if per-request cache is explicitly disabled
	if options.Cache != nil && options.Cache.Enabled != nil && !*options.Cache.Enabled {
		return false
	}

	// Check if per-request cache is explicitly enabled
	if options.Cache != nil && options.Cache.Enabled != nil && *options.Cache.Enabled {
		return isMethodCacheable(cacheConfig, options.Method)
	}

	// Default: caching is disabled unless explicitly enabled per-request
	return false
}

// isMethodCacheable checks if the HTTP method can be cached
func isMethodCacheable(cacheConfig *CacheConfig, method string) bool {
	methods := cacheConfig.CacheableMethods
	if methods == nil {
		methods = []string{"GET"}
	}
	for _, m := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// shouldForceRefresh checks if the request should bypass cache
func shouldForceRefresh(options *RequestOptions) bool {
	return options.Cache != nil && options.Cache.ForceRefresh
}

// generateCacheKey generates the cache key for a request
func generateCacheKey(cacheConfig *CacheConfig, options *RequestOptions, fullURL string) string {
	// Check for custom key in request options
	if options.Cache != nil && options.Cache.CustomKey != "" {
		return options.Cache.CustomKey
	}

	// Use custom key function if configured
	if cacheConfig.KeyFunc != nil {
		return cacheConfig.KeyFunc(options.Method, fullURL, options.Headers)
	}

	// Use default key function
	return DefaultCacheKeyFunc(options.Method, fullURL, options.Headers)
}

// getCacheTTL determines the TTL for a cached response
func getCacheTTL(cacheConfig *CacheConfig, options *RequestOptions) time.Duration {
	// Per-request TTL takes precedence
	if options.Cache != nil && options.Cache.TTL > 0 {
		return options.Cache.TTL
	}

	// Fall back to client default TTL
	return cacheConfig.DefaultTTL
}
