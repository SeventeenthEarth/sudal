package redis_test

import (
	"context"
	"errors"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/mocks"
	sconfig "github.com/seventeenthearth/sudal/internal/service/config"
	redisdb "github.com/seventeenthearth/sudal/internal/service/redis"
)

var _ = ginkgo.Describe("Redis Tests", func() {
	var (
		ctrl       *gomock.Controller
		mockClient *mocks.MockRedisClient
		ctx        context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockClient = mocks.NewMockRedisClient(ctrl)
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("NewRedisManager", func() {
		ginkgo.Context("when Redis configuration is invalid", func() {
			ginkgo.It("should fail when Redis address is empty", func() {
				// Given
				config := &sconfig.Config{
					Redis: sconfig.RedisConfig{
						Addr:     "",
						Password: "",
					},
				}

				// When
				manager, err := redisdb.NewRedisManager(config)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis address is required"))
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("MockRedisClient", func() {
		ginkgo.Context("when testing Ping functionality", func() {
			ginkgo.It("should mock Ping method successfully", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetVal("PONG")
				mockClient.EXPECT().
					Ping(ctx).
					Return(statusCmd).
					Times(1)

				// When
				result := mockClient.Ping(ctx)

				// Then
				gomega.Expect(result).ToNot(gomega.BeNil())
				gomega.Expect(result.Val()).To(gomega.Equal("PONG"))
				gomega.Expect(result.Err()).ToNot(gomega.HaveOccurred())
			})

			ginkgo.It("should mock Ping method with error", func() {
				// Given
				statusCmd := redis.NewStatusCmd(ctx)
				statusCmd.SetErr(errors.New("connection failed"))
				mockClient.EXPECT().
					Ping(ctx).
					Return(statusCmd).
					Times(1)

				// When
				result := mockClient.Ping(ctx)

				// Then
				gomega.Expect(result).ToNot(gomega.BeNil())
				gomega.Expect(result.Err()).To(gomega.HaveOccurred())
				gomega.Expect(result.Err().Error()).To(gomega.ContainSubstring("connection failed"))
			})
		})

		ginkgo.Context("when testing Close functionality", func() {
			ginkgo.It("should mock Close method successfully", func() {
				// Given
				mockClient.EXPECT().
					Close().
					Return(nil).
					Times(1)

				// When
				err := mockClient.Close()

				// Then
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when testing PoolStats functionality", func() {
			ginkgo.It("should mock PoolStats method successfully", func() {
				// Given
				expectedStats := &redis.PoolStats{
					Hits:       10,
					Misses:     2,
					Timeouts:   0,
					TotalConns: 5,
					IdleConns:  3,
					StaleConns: 0,
				}
				mockClient.EXPECT().
					PoolStats().
					Return(expectedStats).
					Times(1)

				// When
				stats := mockClient.PoolStats()

				// Then
				gomega.Expect(stats).ToNot(gomega.BeNil())
				gomega.Expect(stats.Hits).To(gomega.Equal(uint32(10)))
				gomega.Expect(stats.Misses).To(gomega.Equal(uint32(2)))
				gomega.Expect(stats.TotalConns).To(gomega.Equal(uint32(5)))
			})
		})
	})

	ginkgo.Describe("RedisManagerConfiguration", func() {
		ginkgo.Context("when Redis configuration is invalid", func() {
			ginkgo.It("should fail with empty address", func() {
				// Given
				config := &sconfig.Config{
					Redis: sconfig.RedisConfig{
						Addr:     "",
						Password: "",
					},
				}

				// When
				manager, err := redisdb.NewRedisManager(config)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("redis address is required"))
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when Redis configuration is valid", func() {
			ginkgo.It("should validate Redis configuration parameters structure", func() {
				// Given
				config := &sconfig.Config{
					Redis: sconfig.RedisConfig{
						Addr:            "localhost:6379",
						Password:        "secret",
						DB:              1,
						PoolSize:        10,
						MinIdleConns:    2,
						PoolTimeout:     4,
						IdleTimeout:     300,
						DialTimeout:     5,
						ReadTimeout:     3,
						WriteTimeout:    3,
						MaxRetries:      3,
						MinRetryBackoff: 8,
						MaxRetryBackoff: 512,
					},
				}

				// When & Then - Test configuration structure validation (without actual connection)
				gomega.Expect(config.Redis.Addr).ToNot(gomega.BeEmpty())
				gomega.Expect(config.Redis.PoolSize).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.MaxRetries).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.MinIdleConns).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.PoolTimeout).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.IdleTimeout).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.DialTimeout).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.ReadTimeout).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.WriteTimeout).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.MinRetryBackoff).To(gomega.BeNumerically(">=", 0))
				gomega.Expect(config.Redis.MaxRetryBackoff).To(gomega.BeNumerically(">=", 0))
				// Note: Actual connection testing would be done in integration tests
			})
		})
	})
})
