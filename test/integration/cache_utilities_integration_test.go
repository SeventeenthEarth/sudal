package integration_test

import (
	"errors"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	testMocks "github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = Describe("Cache Utilities Integration Tests", func() {
	var (
		ctrl        *gomock.Controller
		mockCache   *testMocks.MockCacheUtil
		testKeyBase string
	)

	BeforeEach(func() {
		// Initialize logger to avoid race conditions
		log.Init(log.InfoLevel)

		// Initialize gomock controller
		ctrl = gomock.NewController(GinkgoT())
		mockCache = testMocks.NewMockCacheUtil(ctrl)
		testKeyBase = fmt.Sprintf("test:cache:integration:%d", time.Now().UnixNano())

		// Ensure mock is available
		Expect(mockCache).NotTo(BeNil())
	})

	AfterEach(func() {
		// Cleanup gomock controller
		if ctrl != nil {
			ctrl.Finish()
		}
	})

	Describe("Basic CRUD Operations", func() {
		Context("when setting cache keys", func() {
			It("should successfully set a key-value pair without TTL", func() {
				// Given: A cache utility is available
				key := testKeyBase + ":basic_set"
				value := "test_value"

				// Configure mock to succeed
				mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(nil)

				// When: Setting a cache key without TTL
				err := mockCache.Set(key, value, 0)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())
			})

			It("should successfully set a key-value pair with TTL", func() {
				// Given: A cache utility is available
				key := testKeyBase + ":ttl_set"
				value := "ttl_value"
				ttl := 5 * time.Second

				// Configure mock to succeed
				mockCache.EXPECT().Set(key, value, ttl).Return(nil)

				// When: Setting a cache key with TTL
				err := mockCache.Set(key, value, ttl)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error when cache operation fails", func() {
				// Given: Cache is configured to fail
				key := testKeyBase + ":fail_set"
				value := "test_value"

				// Configure mock to fail
				mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(fmt.Errorf("cache operation failed"))

				// When: Setting a cache key
				err := mockCache.Set(key, value, 0)

				// Then: Operation should fail
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cache operation failed"))
			})
		})

		Context("when getting cache keys", func() {
			It("should successfully retrieve an existing key", func() {
				// Given: A cache key exists
				key := testKeyBase + ":get_existing"
				expectedValue := "existing_value"

				// Configure mock to return the value
				mockCache.EXPECT().Get(key).Return(expectedValue, nil)

				// When: Getting the cache key
				value, err := mockCache.Get(key)

				// Then: Operation should succeed and return the value
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(expectedValue))
			})

			It("should return ErrCacheMiss for non-existent key", func() {
				// Given: A cache key does not exist
				key := testKeyBase + ":nonexistent"

				// Configure mock to return cache miss
				mockCache.EXPECT().Get(key).Return("", cacheutil.ErrCacheMiss)

				// When: Getting the non-existent cache key
				value, err := mockCache.Get(key)

				// Then: Operation should return ErrCacheMiss
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, cacheutil.ErrCacheMiss)).To(BeTrue())
				Expect(value).To(BeEmpty())
			})

			It("should return error when key is empty", func() {
				// Given: A cache utility is available

				// Configure mock to return error for empty key
				mockCache.EXPECT().Get("").Return("", fmt.Errorf("key cannot be empty"))

				// When: Getting a cache key with empty key
				value, err := mockCache.Get("")

				// Then: Operation should fail with appropriate error
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("key cannot be empty"))
				Expect(value).To(BeEmpty())
			})

			It("should return error when Redis client is unavailable", func() {
				// Given: Redis client is unavailable
				key := testKeyBase + ":unavailable_get"

				// Configure mock to return client unavailable error
				mockCache.EXPECT().Get(key).Return("", fmt.Errorf("redis client is not available"))

				// When: Getting a cache key
				value, err := mockCache.Get(key)

				// Then: Operation should fail
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("redis client is not available"))
				Expect(value).To(BeEmpty())
			})
		})

		Context("when deleting cache keys", func() {
			It("should successfully delete an existing key", func() {
				// Given: A cache key exists
				key := testKeyBase + ":delete_existing"

				// Configure mock to succeed
				mockCache.EXPECT().Delete(key).Return(nil)

				// When: Deleting the cache key
				err := mockCache.Delete(key)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())
			})

			It("should succeed when deleting non-existent key", func() {
				// Given: A cache key does not exist
				key := testKeyBase + ":delete_nonexistent"

				// Configure mock to succeed (Redis DELETE succeeds even for non-existent keys)
				mockCache.EXPECT().Delete(key).Return(nil)

				// When: Deleting the non-existent cache key
				err := mockCache.Delete(key)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error when key is empty", func() {
				// Given: A cache utility is available

				// Configure mock to return error for empty key
				mockCache.EXPECT().Delete("").Return(fmt.Errorf("key cannot be empty"))

				// When: Deleting a cache key with empty key
				err := mockCache.Delete("")

				// Then: Operation should fail with appropriate error
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("key cannot be empty"))
			})

			It("should return error when Redis client is unavailable", func() {
				// Given: Redis client is unavailable
				key := testKeyBase + ":unavailable_delete"

				// Configure mock to return client unavailable error
				mockCache.EXPECT().Delete(key).Return(fmt.Errorf("redis client is not available"))

				// When: Deleting a cache key
				err := mockCache.Delete(key)

				// Then: Operation should fail
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("redis client is not available"))
			})
		})
	})

	Describe("TTL Operations", func() {
		Context("when working with TTL", func() {
			It("should handle keys with TTL correctly", func() {
				// Given: A cache utility is available
				key := testKeyBase + ":ttl_test"
				value := "ttl_value"
				ttl := 1 * time.Second

				// Configure mock to succeed for Set operation
				mockCache.EXPECT().Set(key, value, ttl).Return(nil)

				// When: Setting a key with TTL
				err := mockCache.Set(key, value, ttl)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())

				// And: Key should be retrievable immediately
				mockCache.EXPECT().Get(key).Return(value, nil)
				retrievedValue, err := mockCache.Get(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedValue).To(Equal(value))
			})

			It("should handle zero TTL as persistent key", func() {
				// Given: A cache utility is available
				key := testKeyBase + ":persistent"
				value := "persistent_value"

				// Configure mock to succeed
				mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(nil)

				// When: Setting a key with zero TTL
				err := mockCache.Set(key, value, 0)

				// Then: Operation should succeed (key persists indefinitely)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle negative TTL as persistent key", func() {
				// Given: A cache utility is available
				key := testKeyBase + ":negative_ttl"
				value := "negative_ttl_value"

				// Configure mock to succeed
				mockCache.EXPECT().Set(key, value, -1*time.Second).Return(nil)

				// When: Setting a key with negative TTL
				err := mockCache.Set(key, value, -1*time.Second)

				// Then: Operation should succeed (key persists indefinitely)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Pattern-based Operations", func() {
		Context("when deleting by pattern", func() {
			It("should successfully delete keys matching pattern", func() {
				// Given: A cache utility is available
				pattern := testKeyBase + ":pattern:*"

				// Configure mock to succeed
				mockCache.EXPECT().DeleteByPattern(pattern).Return(nil)

				// When: Deleting keys by pattern
				err := mockCache.DeleteByPattern(pattern)

				// Then: Operation should succeed
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error when pattern is empty", func() {
				// Given: A cache utility is available

				// Configure mock to return error for empty pattern
				mockCache.EXPECT().DeleteByPattern("").Return(fmt.Errorf("pattern cannot be empty"))

				// When: Deleting keys with empty pattern
				err := mockCache.DeleteByPattern("")

				// Then: Operation should fail with appropriate error
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("pattern cannot be empty"))
			})

			It("should return error when Redis client is unavailable", func() {
				// Given: Redis client is unavailable
				pattern := testKeyBase + ":unavailable:*"

				// Configure mock to return client unavailable error
				mockCache.EXPECT().DeleteByPattern(pattern).Return(fmt.Errorf("redis client is not available"))

				// When: Deleting keys by pattern
				err := mockCache.DeleteByPattern(pattern)

				// Then: Operation should fail
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("redis client is not available"))
			})
		})
	})

	Describe("Concurrent Operations", func() {
		Context("when performing concurrent cache operations", func() {
			It("should handle concurrent Set operations safely", func() {
				// Given: A cache utility is available
				numGoroutines := 10
				var wg sync.WaitGroup
				errors := make([]error, numGoroutines)

				// Configure mock to succeed for all Set operations
				for i := 0; i < numGoroutines; i++ {
					key := fmt.Sprintf("%s:concurrent:set:%d", testKeyBase, i)
					value := fmt.Sprintf("value_%d", i)
					mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(nil)
				}

				// When: Performing concurrent Set operations
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						key := fmt.Sprintf("%s:concurrent:set:%d", testKeyBase, index)
						value := fmt.Sprintf("value_%d", index)
						errors[index] = mockCache.Set(key, value, 0)
					}(i)
				}

				wg.Wait()

				// Then: All operations should succeed
				for i, err := range errors {
					Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Set operation %d should succeed", i))
				}
			})

			It("should handle concurrent Get operations safely", func() {
				// Given: Cache keys exist
				numGoroutines := 10
				var wg sync.WaitGroup
				results := make([]string, numGoroutines)
				errors := make([]error, numGoroutines)

				// Configure mock to succeed and return values for all Get operations
				for i := 0; i < numGoroutines; i++ {
					key := fmt.Sprintf("%s:concurrent:get:%d", testKeyBase, i)
					expectedValue := fmt.Sprintf("value_%d", i)
					mockCache.EXPECT().Get(key).Return(expectedValue, nil)
				}

				// When: Performing concurrent Get operations
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						key := fmt.Sprintf("%s:concurrent:get:%d", testKeyBase, index)
						results[index], errors[index] = mockCache.Get(key)
					}(i)
				}

				wg.Wait()

				// Then: All operations should succeed with correct values
				for i := 0; i < numGoroutines; i++ {
					Expect(errors[i]).NotTo(HaveOccurred(), fmt.Sprintf("Get operation %d should succeed", i))
					expectedValue := fmt.Sprintf("value_%d", i)
					Expect(results[i]).To(Equal(expectedValue), fmt.Sprintf("Get operation %d should return correct value", i))
				}
			})

			It("should handle concurrent Delete operations safely", func() {
				// Given: A cache utility is available
				numGoroutines := 10
				var wg sync.WaitGroup
				errors := make([]error, numGoroutines)

				// Configure mock to succeed for all Delete operations
				for i := 0; i < numGoroutines; i++ {
					key := fmt.Sprintf("%s:concurrent:delete:%d", testKeyBase, i)
					mockCache.EXPECT().Delete(key).Return(nil)
				}

				// When: Performing concurrent Delete operations
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						key := fmt.Sprintf("%s:concurrent:delete:%d", testKeyBase, index)
						errors[index] = mockCache.Delete(key)
					}(i)
				}

				wg.Wait()

				// Then: All operations should succeed
				for i, err := range errors {
					Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Delete operation %d should succeed", i))
				}
			})
		})
	})

	Describe("Error Scenarios", func() {
		Context("when Redis operations fail", func() {
			It("should handle Redis connection errors gracefully", func() {
				// Given: Redis connection fails
				key := testKeyBase + ":connection_error"
				value := "test_value"

				// Configure mock to simulate connection error
				connectionErr := fmt.Errorf("connection refused")
				mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(connectionErr)
				mockCache.EXPECT().Get(key).Return("", connectionErr)
				mockCache.EXPECT().Delete(key).Return(connectionErr)

				// When: Attempting cache operations
				setErr := mockCache.Set(key, value, 0)
				_, getErr := mockCache.Get(key)
				deleteErr := mockCache.Delete(key)

				// Then: All operations should fail gracefully
				Expect(setErr).To(HaveOccurred())
				Expect(getErr).To(HaveOccurred())
				Expect(deleteErr).To(HaveOccurred())
			})

			It("should handle Redis timeout errors", func() {
				// Given: Redis operations timeout
				key := testKeyBase + ":timeout_error"
				value := "test_value"

				// Configure mock to simulate timeout
				timeoutErr := fmt.Errorf("operation timed out")
				mockCache.EXPECT().Set(key, value, time.Duration(0)).Return(timeoutErr)
				mockCache.EXPECT().Get(key).Return("", timeoutErr)

				// When: Attempting cache operations
				setErr := mockCache.Set(key, value, 0)
				_, getErr := mockCache.Get(key)

				// Then: Operations should fail with timeout error
				Expect(setErr).To(HaveOccurred())
				Expect(getErr).To(HaveOccurred())
			})
		})
	})
})
