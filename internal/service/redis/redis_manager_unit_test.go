package redis_test

import (
	"context"
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/mocks"
	sconfig "github.com/seventeenthearth/sudal/internal/service/config"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	redisdb "github.com/seventeenthearth/sudal/internal/service/redis"
)

var _ = ginkgo.Describe("RedisManager Unit Tests", func() {
	var (
		ctrl         *gomock.Controller
		mockClient   *mocks.MockRedisClient
		redisManager redisdb.RedisManager
		ctx          context.Context
		cfg          *sconfig.Config
	)

	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockClient = mocks.NewMockRedisClient(ctrl)
		ctx = context.Background()

		cfg = &sconfig.Config{
			Redis: sconfig.RedisConfig{
				Addr:            "localhost:6379",
				Password:        "",
				DB:              0,
				PoolSize:        10,
				MinIdleConns:    5,
				MaxRetries:      3,
				MinRetryBackoff: 100,
				MaxRetryBackoff: 1000,
				DialTimeout:     5000,
				ReadTimeout:     3000,
				WriteTimeout:    3000,
				PoolTimeout:     4000,
				IdleTimeout:     300000,
			},
		}

		// Create RedisManager with mock client using the new constructor
		redisManager = redisdb.NewRedisManagerWithClient(mockClient, cfg)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("RedisManager Ping", func() {

		ginkgo.Context("when testing Ping functionality", func() {
			ginkgo.It("should return nil when Redis is healthy", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetVal("PONG")
				mockClient.EXPECT().Ping(ctx).Return(statusCmd)

				// When
				err := redisManager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should handle Redis connection error", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetErr(errors.New("connection refused"))
				mockClient.EXPECT().Ping(ctx).Return(statusCmd).Times(4) // MaxRetries + 1

				// When
				err := redisManager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis health check failed"))
			})

			ginkgo.It("should handle context timeout", func() {
				// Given
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				statusCmd := redis.NewStatusCmd(timeoutCtx)
				statusCmd.SetErr(context.DeadlineExceeded)
				mockClient.EXPECT().Ping(timeoutCtx).Return(statusCmd)

				// When
				err := redisManager.Ping(timeoutCtx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis health check failed"))
			})

			ginkgo.It("should succeed after retry", func() {
				// Given - First call fails, second succeeds
				failCmd := redis.NewStatusCmd(ctx)
				failCmd.SetErr(errors.New("connection refused"))
				successCmd := redis.NewStatusCmd(ctx)
				successCmd.SetVal("PONG")

				gomock.InOrder(
					mockClient.EXPECT().Ping(ctx).Return(failCmd),
					mockClient.EXPECT().Ping(ctx).Return(successCmd),
				)

				// When
				err := redisManager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when testing Close functionality", func() {
			ginkgo.It("should close successfully", func() {
				// Given
				mockClient.EXPECT().Close().Return(nil)

				// When
				err := redisManager.Close()

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return error when close fails", func() {
				// Given
				expectedError := errors.New("failed to close Redis connection")
				mockClient.EXPECT().Close().Return(expectedError)

				// When
				err := redisManager.Close()

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err).To(gomega.Equal(expectedError))
			})
		})

		ginkgo.Context("when testing PoolStats functionality", func() {
			ginkgo.It("should return pool statistics", func() {
				// Given
				expectedStats := &redis.PoolStats{
					Hits:       100,
					Misses:     10,
					Timeouts:   0,
					TotalConns: 5,
					IdleConns:  3,
					StaleConns: 0,
				}
				mockClient.EXPECT().PoolStats().Return(expectedStats)

				// When
				stats := mockClient.PoolStats()

				// Then
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
				gomega.Expect(stats.Hits).To(gomega.Equal(uint32(100)))
				gomega.Expect(stats.TotalConns).To(gomega.Equal(uint32(5)))
				gomega.Expect(stats.IdleConns).To(gomega.Equal(uint32(3)))
			})

			ginkgo.It("should return nil when pool stats unavailable", func() {
				// Given
				mockClient.EXPECT().PoolStats().Return(nil)

				// When
				stats := redisManager.GetConnectionPoolStats()

				// Then
				gomega.Expect(stats).To(gomega.BeNil())
			})

			ginkgo.It("should log pool statistics", func() {
				// Given
				expectedStats := &redis.PoolStats{
					Hits:       50,
					Misses:     5,
					TotalConns: 3,
					IdleConns:  2,
				}
				mockClient.EXPECT().PoolStats().Return(expectedStats)

				// When - This should not panic and should log the stats
				redisManager.LogConnectionPoolStats()

				// Then - No assertion needed as this is a logging function
			})
		})

		ginkgo.Context("when testing multiple operations", func() {
			ginkgo.It("should handle ping followed by pool stats", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetVal("PONG")
				mockClient.EXPECT().Ping(ctx).Return(statusCmd)

				expectedStats := &redis.PoolStats{
					Hits:       50,
					Misses:     5,
					TotalConns: 3,
					IdleConns:  2,
				}
				mockClient.EXPECT().PoolStats().Return(expectedStats)

				// When
				pingErr := redisManager.Ping(ctx)
				stats := redisManager.GetConnectionPoolStats()

				// Then
				gomega.Expect(pingErr).To(gomega.BeNil())
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
			})

			ginkgo.It("should handle failed ping followed by close", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetErr(errors.New("Redis server down"))
				mockClient.EXPECT().Ping(ctx).Return(statusCmd).Times(4) // MaxRetries + 1
				mockClient.EXPECT().Close().Return(nil)

				// When
				pingErr := redisManager.Ping(ctx)
				closeErr := redisManager.Close()

				// Then
				gomega.Expect(pingErr).To(gomega.HaveOccurred())
				gomega.Expect(pingErr.Error()).To(gomega.ContainSubstring("redis health check failed"))
				gomega.Expect(closeErr).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when testing error scenarios", func() {
			ginkgo.It("should handle network timeout errors with retry", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetErr(errors.New("i/o timeout"))
				mockClient.EXPECT().Ping(ctx).Return(statusCmd).Times(4) // MaxRetries + 1

				// When
				err := redisManager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis health check failed"))
			})

			ginkgo.It("should handle authentication errors without retry", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetErr(errors.New("NOAUTH Authentication required"))
				mockClient.EXPECT().Ping(ctx).Return(statusCmd).Times(1) // No retry for auth errors

				// When
				err := redisManager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis health check failed"))
			})

			ginkgo.It("should handle connection pool exhaustion", func() {
				// Given
				expectedStats := &redis.PoolStats{
					Hits:       1000,
					Misses:     100,
					Timeouts:   50, // High timeout count indicates pool exhaustion
					TotalConns: 10,
					IdleConns:  0, // No idle connections
					StaleConns: 2,
				}
				mockClient.EXPECT().PoolStats().Return(expectedStats)

				// When
				stats := redisManager.GetConnectionPoolStats()

				// Then
				gomega.Expect(stats.Timeouts).To(gomega.BeNumerically(">", 0))
				gomega.Expect(stats.IdleConns).To(gomega.Equal(uint32(0)))
				gomega.Expect(stats.TotalConns).To(gomega.Equal(uint32(10)))
			})

			ginkgo.It("should test ExecuteWithRetry method", func() {
				// Given
				operation := "test_operation"
				expectedError := errors.New("connection refused")
				callCount := 0

				// When
				err := redisManager.ExecuteWithRetry(ctx, operation, func() error {
					callCount++
					if callCount <= 2 {
						return expectedError
					}
					return nil
				})

				// Then
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(callCount).To(gomega.Equal(3)) // Failed twice, succeeded on third
			})
		})
	})
})
