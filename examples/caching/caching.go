package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rezmoss/axios4go"
)

func main() {
	fmt.Println("=== axios4go Caching Examples ===")

	// Example 1: Basic caching setup
	basicCachingExample()

	// Example 2: Per-request TTL
	perRequestTTLExample()

	// Example 3: Force refresh (bypass cache)
	forceRefreshExample()

	// Example 4: Custom cache key function
	customKeyFuncExample()

	// Example 5: Custom cache key per-request
	customKeyPerRequestExample()

	// Example 6: Cache statistics
	cacheStatisticsExample()

	// Example 7: Explicitly disable cache for specific request
	disableCacheExample()

	fmt.Println("\n=== All caching examples completed! ===")
}

// Example 1: Basic caching setup with MemoryCache
func basicCachingExample() {
	fmt.Println("--- Example 1: Basic Caching ---")

	// Create a memory cache with options
	cache := axios4go.NewMemoryCache(&axios4go.MemoryCacheOptions{
		MaxSize:         100,             // Maximum 100 entries
		CleanupInterval: 1 * time.Minute, // Clean expired entries every minute
	})
	defer cache.Close() // Always close to stop cleanup goroutine

	// Create client with cache configuration
	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute, // Default TTL for all cached responses
	})

	// First request - will be a cache MISS
	fmt.Println("Making first request (cache miss expected)...")
	resp1, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/rezmoss",
		Cache:  axios4go.CacheEnabled(5 * time.Minute), // Enable caching with 5 min TTL
	})
	if err != nil {
		log.Printf("First request error: %v\n", err)
		return
	}
	fmt.Printf("First request status: %d\n", resp1.StatusCode)

	// Second request - will be a cache HIT
	fmt.Println("Making second request (cache hit expected)...")
	resp2, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/rezmoss",
		Cache:  axios4go.CacheEnabled(5 * time.Minute),
	})
	if err != nil {
		log.Printf("Second request error: %v\n", err)
		return
	}
	fmt.Printf("Second request status: %d\n", resp2.StatusCode)

	// Check stats
	stats := client.CacheStats()
	fmt.Printf("Cache stats - Hits: %d, Misses: %d, Size: %d\n\n", stats.Hits, stats.Misses, stats.Size)
}

// Example 2: Per-request TTL that overrides default
func perRequestTTLExample() {
	fmt.Println("--- Example 2: Per-Request TTL ---")

	cache := axios4go.NewMemoryCache(nil)
	defer cache.Close()

	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 10 * time.Minute, // Default 10 minutes
	})

	// Request with short TTL (overrides default)
	fmt.Println("Making request with 30-second TTL...")
	_, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/golang",
		Cache:  axios4go.CacheEnabled(30 * time.Second), // Override with 30 seconds
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}
	fmt.Println("Request completed with custom 30-second TTL")

	// Request with longer TTL
	fmt.Println("Making request with 1-hour TTL...")
	_, err = client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/google",
		Cache:  axios4go.CacheEnabled(1 * time.Hour), // Override with 1 hour
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}
	fmt.Println("Request completed with custom 1-hour TTL")
}

// Example 3: Force refresh to bypass cache
func forceRefreshExample() {
	fmt.Println("--- Example 3: Force Refresh ---")

	cache := axios4go.NewMemoryCache(nil)
	defer cache.Close()

	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	})

	// First request - populate cache
	fmt.Println("Making initial request to populate cache...")
	_, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Cache:  axios4go.CacheEnabled(5 * time.Minute),
	})
	if err != nil {
		log.Printf("First request error: %v\n", err)
		return
	}

	stats := client.CacheStats()
	fmt.Printf("After first request - Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)

	// Second request with ForceRefresh - bypasses cache
	fmt.Println("Making request with ForceRefresh (bypasses cache)...")
	_, err = client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Cache: &axios4go.RequestCacheOptions{
			Enabled:      axios4go.Bool(true),
			TTL:          5 * time.Minute,
			ForceRefresh: true, // Bypass cache, fetch fresh data
		},
	})
	if err != nil {
		log.Printf("Force refresh request error: %v\n", err)
		return
	}

	stats = client.CacheStats()
	fmt.Printf("After force refresh - Hits: %d, Misses: %d (still 1 miss because ForceRefresh doesn't check cache)\n\n", stats.Hits, stats.Misses)
}

// Example 4: Custom cache key function
func customKeyFuncExample() {
	fmt.Println("--- Example 4: Custom Cache Key Function ---")

	cache := axios4go.NewMemoryCache(nil)
	defer cache.Close()

	// Create client with custom key function that includes Authorization header
	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
		KeyFunc: func(method, fullURL string, headers map[string]string) string {
			// Include Authorization header in cache key
			// This ensures different users get different cached responses
			auth := ""
			if headers != nil {
				auth = headers["Authorization"]
			}
			key := fmt.Sprintf("%s:%s:auth=%s", method, fullURL, auth)
			fmt.Printf("Generated cache key: %s\n", key)
			return key
		},
	})

	// Request without auth
	fmt.Println("Making request without auth header...")
	_, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Cache:  axios4go.CacheEnabled(5 * time.Minute),
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
	}

	// Request with auth (different cache key)
	fmt.Println("Making request with auth header (different cache key)...")
	_, err = client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Headers: map[string]string{
			"Authorization": "Bearer fake-token",
		},
		Cache: axios4go.CacheEnabled(5 * time.Minute),
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
	}

	stats := client.CacheStats()
	fmt.Printf("Cache size: %d (2 entries for different auth headers)\n\n", stats.Size)
}

// Example 5: Custom cache key per-request
func customKeyPerRequestExample() {
	fmt.Println("--- Example 5: Custom Cache Key Per-Request ---")

	cache := axios4go.NewMemoryCache(nil)
	defer cache.Close()

	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	})

	// Request with custom cache key
	fmt.Println("Making request with custom cache key 'my-zen-cache'...")
	_, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Cache: &axios4go.RequestCacheOptions{
			Enabled:   axios4go.Bool(true),
			TTL:       5 * time.Minute,
			CustomKey: "my-zen-cache", // Use custom key instead of auto-generated
		},
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}

	// Verify custom key was used
	entry := cache.Get("my-zen-cache")
	if entry != nil {
		fmt.Println("Successfully retrieved cache entry using custom key 'my-zen-cache'")
		fmt.Printf("Cached response status: %d\n\n", entry.StatusCode)
	}
}

// Example 6: Cache statistics and management
func cacheStatisticsExample() {
	fmt.Println("--- Example 6: Cache Statistics ---")

	cache := axios4go.NewMemoryCache(&axios4go.MemoryCacheOptions{
		MaxSize: 10,
	})
	defer cache.Close()

	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	})

	// Make several requests
	endpoints := []string{"/zen", "/users/rezmoss", "/users/golang"}
	for _, endpoint := range endpoints {
		_, err := client.Request(&axios4go.RequestOptions{
			Method: "GET",
			URL:    endpoint,
			Cache:  axios4go.CacheEnabled(5 * time.Minute),
		})
		if err != nil {
			log.Printf("Request to %s error: %v\n", endpoint, err)
		}
	}

	// Check statistics
	stats := client.CacheStats()
	fmt.Printf("Cache Statistics:\n")
	fmt.Printf("  - Hits: %d\n", stats.Hits)
	fmt.Printf("  - Misses: %d\n", stats.Misses)
	fmt.Printf("  - Size: %d entries\n", stats.Size)

	// Clear cache
	fmt.Println("\nClearing cache...")
	client.ClearCache()

	stats = client.CacheStats()
	fmt.Printf("After clear - Size: %d entries\n\n", stats.Size)
}

// Example 7: Explicitly disable cache for a specific request
func disableCacheExample() {
	fmt.Println("--- Example 7: Disable Cache for Specific Request ---")

	cache := axios4go.NewMemoryCache(nil)
	defer cache.Close()

	client := axios4go.NewClientWithCache("https://api.github.com", &axios4go.CacheConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
	})

	// First request - cache enabled
	fmt.Println("Making request with cache enabled...")
	_, err := client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/zen",
		Cache:  axios4go.CacheEnabled(5 * time.Minute),
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}

	stats := client.CacheStats()
	fmt.Printf("After cached request - Size: %d\n", stats.Size)

	// Second request - cache explicitly disabled
	fmt.Println("Making request with cache explicitly disabled...")
	_, err = client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/microsoft",
		Cache:  axios4go.CacheDisabled(), // Explicitly disable cache
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}

	stats = client.CacheStats()
	fmt.Printf("After disabled-cache request - Size: %d (unchanged)\n", stats.Size)

	// Third request - no cache option (default is disabled)
	fmt.Println("Making request without cache option (default disabled)...")
	_, err = client.Request(&axios4go.RequestOptions{
		Method: "GET",
		URL:    "/users/apple",
		// No Cache option - caching is disabled by default (opt-in model)
	})
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return
	}

	stats = client.CacheStats()
	fmt.Printf("After no-cache-option request - Size: %d (unchanged)\n", stats.Size)
}
