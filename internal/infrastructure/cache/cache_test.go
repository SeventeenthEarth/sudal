package cache

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("CacheUtil", func() {
	var (
		cacheUtil CacheUtil
	)

	ginkgo.Describe("NewCacheUtil", func() {
		ginkgo.Context("when creating a new cache utility", func() {
			ginkgo.It("should create a cache utility with nil redis client (KV)", func() {
				// When
				cacheUtil = NewCacheUtil(nil)

				// Then
				gomega.Expect(cacheUtil).NotTo(gomega.BeNil())

				// Type assert to access internal fields
				if impl, ok := cacheUtil.(*CacheUtilImpl); ok {
					gomega.Expect(impl.kv).To(gomega.BeNil())
					gomega.Expect(impl.logger).NotTo(gomega.BeNil())
				} else {
					ginkgo.Fail("Expected CacheUtilImpl implementation")
				}
			})
		})
	})

	ginkgo.Describe("Set", func() {
		ginkgo.BeforeEach(func() {
			cacheUtil = NewCacheUtil(nil)
		})

		ginkgo.Context("when setting a cache key", func() {
			ginkgo.It("should return error for empty key", func() {
				// When
				err := cacheUtil.Set("", "value", 0)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("key cannot be empty"))
			})

			ginkgo.It("should return error when redis client is not available", func() {
				// When
				err := cacheUtil.Set("test-key", "test-value", 0)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis client is not available"))
			})
		})
	})

	ginkgo.Describe("Get", func() {
		ginkgo.BeforeEach(func() {
			cacheUtil = NewCacheUtil(nil)
		})

		ginkgo.Context("when getting a cache key", func() {
			ginkgo.It("should return error for empty key", func() {
				// When
				_, err := cacheUtil.Get("")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("key cannot be empty"))
			})

			ginkgo.It("should return error when redis client is not available", func() {
				// When
				_, err := cacheUtil.Get("test-key")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis client is not available"))
			})
		})
	})

	ginkgo.Describe("Delete", func() {
		ginkgo.BeforeEach(func() {
			cacheUtil = NewCacheUtil(nil)
		})

		ginkgo.Context("when deleting a cache key", func() {
			ginkgo.It("should return error for empty key", func() {
				// When
				err := cacheUtil.Delete("")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("key cannot be empty"))
			})

			ginkgo.It("should return error when redis client is not available", func() {
				// When
				err := cacheUtil.Delete("test-key")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis client is not available"))
			})
		})
	})

	ginkgo.Describe("DeleteByPattern", func() {
		ginkgo.BeforeEach(func() {
			cacheUtil = NewCacheUtil(nil)
		})

		ginkgo.Context("when deleting cache keys by pattern", func() {
			ginkgo.It("should return error for empty pattern", func() {
				// When
				err := cacheUtil.DeleteByPattern("")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("pattern cannot be empty"))
			})

			ginkgo.It("should return error when redis client is not available", func() {
				// When
				err := cacheUtil.DeleteByPattern("test-*")

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis client is not available"))
			})
		})
	})

	ginkgo.Describe("ErrCacheMiss", func() {
		ginkgo.Context("when checking cache miss error", func() {
			ginkgo.It("should have correct error message", func() {
				// Then
				gomega.Expect(ErrCacheMiss).NotTo(gomega.BeNil())
				gomega.Expect(ErrCacheMiss.Error()).To(gomega.Equal("cache miss: key not found or expired"))
			})
		})
	})
})
