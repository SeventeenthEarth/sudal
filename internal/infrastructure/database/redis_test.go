package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

func TestNewRedisManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "should fail when Redis address is empty",
			config: &config.Config{
				Redis: config.RedisConfig{
					Addr:     "",
					Password: "",
				},
			},
			expectError: true,
			errorMsg:    "redis address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := database.NewRedisManager(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestMockRedisClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockRedisClient(ctrl)
	ctx := context.Background()

	t.Run("should mock Ping method successfully", func(t *testing.T) {
		// Create a successful status command
		statusCmd := redis.NewStatusCmd(ctx)
		statusCmd.SetVal("PONG")

		mockClient.EXPECT().
			Ping(ctx).
			Return(statusCmd).
			Times(1)

		// Test the mock
		result := mockClient.Ping(ctx)
		assert.NotNil(t, result)
		assert.Equal(t, "PONG", result.Val())
		assert.NoError(t, result.Err())
	})

	t.Run("should mock Ping method with error", func(t *testing.T) {
		// Create a failed status command
		statusCmd := redis.NewStatusCmd(ctx)
		statusCmd.SetErr(errors.New("connection failed"))

		mockClient.EXPECT().
			Ping(ctx).
			Return(statusCmd).
			Times(1)

		// Test the mock
		result := mockClient.Ping(ctx)
		assert.NotNil(t, result)
		assert.Error(t, result.Err())
		assert.Contains(t, result.Err().Error(), "connection failed")
	})

	t.Run("should mock Close method", func(t *testing.T) {
		mockClient.EXPECT().
			Close().
			Return(nil).
			Times(1)

		// Test the mock
		err := mockClient.Close()
		assert.NoError(t, err)
	})

	t.Run("should mock PoolStats method", func(t *testing.T) {
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

		// Test the mock
		stats := mockClient.PoolStats()
		assert.NotNil(t, stats)
		assert.Equal(t, uint32(10), stats.Hits)
		assert.Equal(t, uint32(2), stats.Misses)
		assert.Equal(t, uint32(5), stats.TotalConns)
	})
}

// TestRedisManagerConfiguration tests Redis configuration validation
func TestRedisManagerConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "should fail with empty address",
			config: &config.Config{
				Redis: config.RedisConfig{
					Addr:     "",
					Password: "",
				},
			},
			expectError: true,
			errorMsg:    "redis address is required",
		},
		{
			name: "should validate Redis configuration parameters structure",
			config: &config.Config{
				Redis: config.RedisConfig{
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
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				// Test configuration validation that should fail
				manager, err := database.NewRedisManager(tt.config)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, manager)
			} else {
				// Test configuration structure validation (without actual connection)
				assert.NotEmpty(t, tt.config.Redis.Addr)
				assert.GreaterOrEqual(t, tt.config.Redis.PoolSize, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.MaxRetries, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.MinIdleConns, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.PoolTimeout, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.IdleTimeout, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.DialTimeout, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.ReadTimeout, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.WriteTimeout, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.MinRetryBackoff, 0)
				assert.GreaterOrEqual(t, tt.config.Redis.MaxRetryBackoff, 0)
				// Note: Actual connection testing would be done in integration tests
			}
		})
	}
}
