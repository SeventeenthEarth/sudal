package e2e

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// Cache utility tests don't need a running server, but we define serverURL for consistency
const cacheServerURL = ""

// TestCacheUtility tests the Cache Utility functionality
func TestCacheUtility(t *testing.T) {
	// BDD Scenarios for Cache Utility Basic CRUD Operations
	basicCRUDScenarios := []steps.BDDScenario{
		{
			Name:        "Set cache key without TTL",
			Description: "Should successfully set a key-value pair without TTL",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenISetCacheKey(ctx, "basic_set_test", "test_value")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheOperationShouldSucceed(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Get existing cache key",
			Description: "Should successfully retrieve an existing key-value pair",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyExists(ctx, "basic_get_test", "expected_value")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIGetCacheKey(ctx, "basic_get_test")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheValueShouldBe(ctx, "expected_value")
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Get non-existent cache key",
			Description: "Should return ErrCacheMiss for non-existent key",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyDoesNotExist(ctx, "nonexistent_key")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIGetCacheKey(ctx, "nonexistent_key")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheKeyShouldNotExist(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Delete existing cache key",
			Description: "Should successfully delete an existing key",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyExists(ctx, "delete_test", "value_to_delete")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIDeleteCacheKey(ctx, "delete_test")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheOperationShouldSucceed(ctx)
				// Verify key is actually deleted
				steps.WhenIGetCacheKey(ctx, "delete_test")
				steps.ThenCacheKeyShouldNotExist(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Delete non-existent cache key",
			Description: "Should succeed when deleting non-existent key",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyDoesNotExist(ctx, "nonexistent_delete")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIDeleteCacheKey(ctx, "nonexistent_delete")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheOperationShouldSucceed(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
	}

	// Run basic CRUD scenarios
	steps.RunBDDScenarios(t, cacheServerURL, basicCRUDScenarios)
}

// TestCacheUtilityTTL tests the Cache Utility TTL functionality
func TestCacheUtilityTTL(t *testing.T) {
	// BDD Scenarios for Cache Utility TTL Operations
	ttlScenarios := []steps.BDDScenario{
		{
			Name:        "Set cache key with TTL",
			Description: "Should successfully set a key-value pair with TTL",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenISetCacheKeyWithTTL(ctx, "ttl_set_test", "ttl_value", 5*time.Second)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheOperationShouldSucceed(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Get cache key before TTL expires",
			Description: "Should successfully retrieve key before TTL expiration",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyExistsWithTTL(ctx, "ttl_before_test", "ttl_before_value", 3*time.Second)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIGetCacheKey(ctx, "ttl_before_test")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheValueShouldBe(ctx, "ttl_before_value")
				steps.CleanupCacheTestKeys(ctx)
			},
		},
		{
			Name:        "Get cache key after TTL expires",
			Description: "Should return ErrCacheMiss after TTL expiration",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
				steps.GivenCacheKeyExistsWithTTL(ctx, "ttl_after_test", "ttl_after_value", 1*time.Second)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIWaitForDuration(ctx, 2*time.Second) // Wait for TTL to expire
				steps.WhenIGetCacheKey(ctx, "ttl_after_test")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenCacheKeyShouldNotExist(ctx)
				steps.CleanupCacheTestKeys(ctx)
			},
		},
	}

	// Run TTL scenarios
	steps.RunBDDScenarios(t, cacheServerURL, ttlScenarios)
}

// TestCacheUtilityConcurrency tests the Cache Utility concurrent operations
func TestCacheUtilityConcurrency(t *testing.T) {
	// BDD Scenarios for Cache Utility Concurrent Operations
	concurrencyScenarios := []steps.BDDScenario{
		{
			Name:        "Concurrent cache operations",
			Description: "Should handle concurrent Set, Get, and Delete operations safely",
			Given: func(ctx *steps.TestContext) {
				steps.GivenCacheUtilityIsAvailable(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// Perform concurrent operations with distinct keys per goroutine
				var wg sync.WaitGroup
				var errorsMutex sync.Mutex
				var errors []string
				numGoroutines := 10
				operationsPerGoroutine := 5

				// Get the cache utility from the main context
				cacheUtil := steps.GetCacheUtil(ctx.CacheTestContext)
				if cacheUtil == nil {
					ctx.T.Fatal("Cache utility is not available")
					return
				}

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()

						for j := 0; j < operationsPerGoroutine; j++ {
							keySuffix := fmt.Sprintf("concurrent_g%d_op%d", goroutineID, j)
							value := fmt.Sprintf("value_g%d_op%d", goroutineID, j)
							key := fmt.Sprintf("e2e_test_cache_%s", keySuffix)

							// Set operation
							err := cacheUtil.Set(key, value, 0)
							if err != nil {
								errorsMutex.Lock()
								errors = append(errors, fmt.Sprintf("Goroutine %d operation %d: Set failed: %v", goroutineID, j, err))
								errorsMutex.Unlock()
								return
							}

							// Get operation
							retrievedValue, err := cacheUtil.Get(key)
							if err != nil {
								errorsMutex.Lock()
								errors = append(errors, fmt.Sprintf("Goroutine %d operation %d: Get failed: %v", goroutineID, j, err))
								errorsMutex.Unlock()
								return
							}
							if retrievedValue != value {
								errorsMutex.Lock()
								errors = append(errors, fmt.Sprintf("Goroutine %d operation %d: Expected value %s, got %s", goroutineID, j, value, retrievedValue))
								errorsMutex.Unlock()
								return
							}

							// Delete operation
							err = cacheUtil.Delete(key)
							if err != nil {
								errorsMutex.Lock()
								errors = append(errors, fmt.Sprintf("Goroutine %d operation %d: Delete failed: %v", goroutineID, j, err))
								errorsMutex.Unlock()
								return
							}
						}
					}(i)
				}

				wg.Wait()

				// Check for any errors that occurred during concurrent operations
				if len(errors) > 0 {
					for _, err := range errors {
						ctx.T.Error(err)
					}
				}
			},
			Then: func(ctx *steps.TestContext) {
				// All operations should have completed without errors
				// The test will fail if any goroutine reports an error
				steps.CleanupCacheTestKeys(ctx)
			},
		},
	}

	// Run concurrency scenarios
	steps.RunBDDScenarios(t, cacheServerURL, concurrencyScenarios)
}
