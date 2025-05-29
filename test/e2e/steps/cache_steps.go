package steps

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
)

// NewCacheTestContext creates a new cache test context
func NewCacheTestContext() *CacheTestContext {
	return &CacheTestContext{
		TestKeyPrefix: "e2e_test_cache_",
		CreatedKeys:   make([]string, 0),
		mu:            &sync.Mutex{},
	}
}

// AddTestKey adds a key to the list of created test keys for cleanup
func (ctx *CacheTestContext) AddTestKey(key string) {
	if mu, ok := ctx.mu.(*sync.Mutex); ok {
		mu.Lock()
		defer mu.Unlock()
	}
	ctx.CreatedKeys = append(ctx.CreatedKeys, key)
}

// GetTestKey returns a test key with the proper prefix
func (ctx *CacheTestContext) GetTestKey(suffix string) string {
	key := fmt.Sprintf("%s%s", ctx.TestKeyPrefix, suffix)
	ctx.AddTestKey(key)
	return key
}

// CleanupTestKeys removes all test keys created during the test
func (ctx *CacheTestContext) CleanupTestKeys() {
	if ctx.CacheUtil == nil {
		return
	}

	if mu, ok := ctx.mu.(*sync.Mutex); ok {
		mu.Lock()
		defer mu.Unlock()
	}

	if cacheUtil, ok := ctx.CacheUtil.(*cacheutil.CacheUtil); ok {
		for _, key := range ctx.CreatedKeys {
			_ = cacheUtil.Delete(key) // Ignore errors during cleanup
		}
	}
	ctx.CreatedKeys = ctx.CreatedKeys[:0] // Clear the slice
}

// Helper function to get cache utility with type assertion
func getCacheUtil(ctx *CacheTestContext) *cacheutil.CacheUtil {
	if cacheUtil, ok := ctx.CacheUtil.(*cacheutil.CacheUtil); ok {
		return cacheUtil
	}
	return nil
}

// GetCacheUtil is a public helper function to get cache utility from test context
func GetCacheUtil(ctx *CacheTestContext) *cacheutil.CacheUtil {
	return getCacheUtil(ctx)
}

// Cache BDD Step Functions

// GivenCacheUtilityIsAvailable initializes the cache utility for testing in BDD style
func GivenCacheUtilityIsAvailable(ctx *TestContext) {
	// Set test environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")

	// Load configuration first
	cfg, err := config.LoadConfig("")
	if err != nil {
		ctx.T.Errorf("Expected configuration to load successfully, but got error: %v", err)
		return
	}

	// Create Redis manager directly (bypassing DI for E2E tests)
	redisManager, err := database.NewRedisManager(cfg)
	if err != nil {
		ctx.T.Errorf("Expected Redis manager to be created successfully, but got error: %v", err)
		return
	}
	if redisManager == nil {
		ctx.T.Errorf("Expected Redis manager to exist, but it was nil")
		return
	}

	// Create cache utility directly
	cacheUtil := cacheutil.NewCacheUtil(redisManager)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be created successfully, but it was nil")
		return
	}

	// Create cache test context if it doesn't exist
	if ctx.CacheTestContext == nil {
		ctx.CacheTestContext = NewCacheTestContext()
	}
	ctx.CacheTestContext.CacheUtil = cacheUtil
}

// GivenCacheKeyDoesNotExist ensures a cache key does not exist in BDD style
func GivenCacheKeyDoesNotExist(ctx *TestContext, keySuffix string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	_ = cacheUtil.Delete(key) // Ensure key doesn't exist
}

// GivenCacheKeyExists sets a cache key with a value in BDD style
func GivenCacheKeyExists(ctx *TestContext, keySuffix, value string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	err := cacheUtil.Set(key, value, 0) // No TTL
	if err != nil {
		ctx.T.Errorf("Expected to set cache key for test setup, but got error: %v", err)
	}
}

// GivenCacheKeyExistsWithTTL sets a cache key with a value and TTL in BDD style
func GivenCacheKeyExistsWithTTL(ctx *TestContext, keySuffix, value string, ttl time.Duration) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	err := cacheUtil.Set(key, value, ttl)
	if err != nil {
		ctx.T.Errorf("Expected to set cache key with TTL for test setup, but got error: %v", err)
	}
}

// WhenISetCacheKey sets a cache key with a value in BDD style
func WhenISetCacheKey(ctx *TestContext, keySuffix, value string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Set(key, value, 0)
}

// WhenISetCacheKeyWithTTL sets a cache key with a value and TTL in BDD style
func WhenISetCacheKeyWithTTL(ctx *TestContext, keySuffix, value string, ttl time.Duration) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Set(key, value, ttl)
}

// WhenIGetCacheKey retrieves a cache key in BDD style
func WhenIGetCacheKey(ctx *TestContext, keySuffix string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastValue, ctx.CacheTestContext.LastError = cacheUtil.Get(key)
}

// WhenIDeleteCacheKey deletes a cache key in BDD style
func WhenIDeleteCacheKey(ctx *TestContext, keySuffix string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Delete(key)
}

// WhenIWaitForDuration waits for a specified duration
func WhenIWaitForDuration(ctx *TestContext, duration time.Duration) {
	time.Sleep(duration)
}

// ThenCacheOperationShouldSucceed verifies that the last cache operation succeeded in BDD style
func ThenCacheOperationShouldSucceed(ctx *TestContext) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	if ctx.CacheTestContext.LastError != nil {
		ctx.T.Errorf("Expected cache operation to succeed, but got error: %v", ctx.CacheTestContext.LastError)
	}
}

// ThenCacheOperationShouldFail verifies that the last cache operation failed in BDD style
func ThenCacheOperationShouldFail(ctx *TestContext) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	if ctx.CacheTestContext.LastError == nil {
		ctx.T.Errorf("Expected cache operation to fail, but it succeeded")
	}
}

// ThenCacheValueShouldBe verifies that the retrieved cache value matches expected value in BDD style
func ThenCacheValueShouldBe(ctx *TestContext, expectedValue string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	if ctx.CacheTestContext.LastError != nil {
		ctx.T.Errorf("Expected to be able to retrieve cache value, but got error: %v", ctx.CacheTestContext.LastError)
		return
	}
	if ctx.CacheTestContext.LastValue != expectedValue {
		ctx.T.Errorf("Expected cache value to be '%s', but got '%s'", expectedValue, ctx.CacheTestContext.LastValue)
	}
}

// ThenCacheKeyShouldNotExist verifies that a cache key does not exist (returns ErrCacheMiss) in BDD style
func ThenCacheKeyShouldNotExist(ctx *TestContext) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	if ctx.CacheTestContext.LastError == nil {
		ctx.T.Errorf("Expected to get an error when key doesn't exist, but operation succeeded")
		return
	}
	if !errors.Is(ctx.CacheTestContext.LastError, cacheutil.ErrCacheMiss) {
		ctx.T.Errorf("Expected error to be ErrCacheMiss, but got: %v", ctx.CacheTestContext.LastError)
	}
}

// ThenCacheKeyShouldExist verifies that a cache key exists and can be retrieved in BDD style
func ThenCacheKeyShouldExist(ctx *TestContext, keySuffix string) {
	if ctx.CacheTestContext == nil {
		ctx.T.Errorf("Expected cache test context to be initialized, but it was nil")
		return
	}
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	if cacheUtil == nil {
		ctx.T.Errorf("Expected cache utility to be available, but it was nil")
		return
	}

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	_, err := cacheUtil.Get(key)
	if err != nil {
		ctx.T.Errorf("Expected cache key '%s' to exist and be retrievable, but got error: %v", key, err)
	}
}

// CleanupCacheTestKeys cleans up all test keys created during the test
func CleanupCacheTestKeys(ctx *TestContext) {
	if ctx.CacheTestContext != nil {
		ctx.CacheTestContext.CleanupTestKeys()
	}
}
