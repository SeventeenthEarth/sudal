package steps

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// GivenCacheUtilityIsAvailable initializes the cache utility for testing
func GivenCacheUtilityIsAvailable(ctx *TestContext) {
	// Set test environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")

	// Load configuration first
	cfg, err := config.LoadConfig("")
	require.NoError(ctx.T, err, "Failed to load configuration")

	// Create Redis manager directly (bypassing DI for E2E tests)
	redisManager, err := database.NewRedisManager(cfg)
	require.NoError(ctx.T, err, "Failed to create Redis manager")
	require.NotNil(ctx.T, redisManager, "Redis manager should not be nil")

	// Create cache utility directly
	cacheUtil := cacheutil.NewCacheUtil(redisManager)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should not be nil")

	// Create cache test context if it doesn't exist
	if ctx.CacheTestContext == nil {
		ctx.CacheTestContext = NewCacheTestContext()
	}
	ctx.CacheTestContext.CacheUtil = cacheUtil
}

// GivenCacheKeyDoesNotExist ensures a cache key does not exist
func GivenCacheKeyDoesNotExist(ctx *TestContext, keySuffix string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	_ = cacheUtil.Delete(key) // Ensure key doesn't exist
}

// GivenCacheKeyExists sets a cache key with a value
func GivenCacheKeyExists(ctx *TestContext, keySuffix, value string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	err := cacheUtil.Set(key, value, 0) // No TTL
	require.NoError(ctx.T, err, "Failed to set cache key for test setup")
}

// GivenCacheKeyExistsWithTTL sets a cache key with a value and TTL
func GivenCacheKeyExistsWithTTL(ctx *TestContext, keySuffix, value string, ttl time.Duration) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	err := cacheUtil.Set(key, value, ttl)
	require.NoError(ctx.T, err, "Failed to set cache key with TTL for test setup")
}

// WhenISetCacheKey sets a cache key with a value
func WhenISetCacheKey(ctx *TestContext, keySuffix, value string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Set(key, value, 0)
}

// WhenISetCacheKeyWithTTL sets a cache key with a value and TTL
func WhenISetCacheKeyWithTTL(ctx *TestContext, keySuffix, value string, ttl time.Duration) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Set(key, value, ttl)
}

// WhenIGetCacheKey retrieves a cache key
func WhenIGetCacheKey(ctx *TestContext, keySuffix string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastValue, ctx.CacheTestContext.LastError = cacheUtil.Get(key)
}

// WhenIDeleteCacheKey deletes a cache key
func WhenIDeleteCacheKey(ctx *TestContext, keySuffix string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	ctx.CacheTestContext.LastError = cacheUtil.Delete(key)
}

// WhenIWaitForDuration waits for a specified duration
func WhenIWaitForDuration(ctx *TestContext, duration time.Duration) {
	time.Sleep(duration)
}

// ThenCacheOperationShouldSucceed verifies that the last cache operation succeeded
func ThenCacheOperationShouldSucceed(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	assert.NoError(ctx.T, ctx.CacheTestContext.LastError, "Cache operation should succeed")
}

// ThenCacheOperationShouldFail verifies that the last cache operation failed
func ThenCacheOperationShouldFail(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	assert.Error(ctx.T, ctx.CacheTestContext.LastError, "Cache operation should fail")
}

// ThenCacheValueShouldBe verifies that the retrieved cache value matches expected value
func ThenCacheValueShouldBe(ctx *TestContext, expectedValue string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	assert.NoError(ctx.T, ctx.CacheTestContext.LastError, "Should be able to retrieve cache value")
	assert.Equal(ctx.T, expectedValue, ctx.CacheTestContext.LastValue, "Cache value should match expected value")
}

// ThenCacheKeyShouldNotExist verifies that a cache key does not exist (returns ErrCacheMiss)
func ThenCacheKeyShouldNotExist(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	assert.Error(ctx.T, ctx.CacheTestContext.LastError, "Should get an error when key doesn't exist")
	assert.True(ctx.T, errors.Is(ctx.CacheTestContext.LastError, cacheutil.ErrCacheMiss),
		"Error should be ErrCacheMiss, got: %v", ctx.CacheTestContext.LastError)
}

// ThenCacheKeyShouldExist verifies that a cache key exists and can be retrieved
func ThenCacheKeyShouldExist(ctx *TestContext, keySuffix string) {
	require.NotNil(ctx.T, ctx.CacheTestContext, "Cache test context should be initialized")
	cacheUtil := getCacheUtil(ctx.CacheTestContext)
	require.NotNil(ctx.T, cacheUtil, "Cache utility should be available")

	key := ctx.CacheTestContext.GetTestKey(keySuffix)
	_, err := cacheUtil.Get(key)
	assert.NoError(ctx.T, err, "Cache key should exist and be retrievable")
}

// CleanupCacheTestKeys cleans up all test keys created during the test
func CleanupCacheTestKeys(ctx *TestContext) {
	if ctx.CacheTestContext != nil {
		ctx.CacheTestContext.CleanupTestKeys()
	}
}
