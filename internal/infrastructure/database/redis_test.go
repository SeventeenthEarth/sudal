package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
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
				RedisAddr:     "",
				RedisPassword: "",
			},
			expectError: true,
			errorMsg:    "redis address is required",
		},
		{
			name: "should succeed with valid Redis configuration",
			config: &config.Config{
				RedisAddr:     "localhost:6379",
				RedisPassword: "",
			},
			expectError: false,
		},
		{
			name: "should succeed with Redis password",
			config: &config.Config{
				RedisAddr:     "localhost:6379",
				RedisPassword: "testpassword",
			},
			expectError: false,
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
				assert.NotNil(t, manager.GetClient())

				// Test ping functionality
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				pingErr := manager.Ping(ctx)
				assert.NoError(t, pingErr)

				// Clean up
				closeErr := manager.Close()
				assert.NoError(t, closeErr)
			}
		})
	}
}

func TestRedisManager_Ping(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
	}

	manager, err := database.NewRedisManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test ping functionality
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pingErr := manager.Ping(ctx)
	assert.NoError(t, pingErr)

	// Clean up
	closeErr := manager.Close()
	assert.NoError(t, closeErr)
}

func TestRedisManager_GetClient(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
	}

	manager, err := database.NewRedisManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	client := manager.GetClient()
	assert.NotNil(t, client)

	// Clean up
	closeErr := manager.Close()
	assert.NoError(t, closeErr)
}

func TestRedisManager_Close(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
	}

	manager, err := database.NewRedisManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test close functionality
	closeErr := manager.Close()
	assert.NoError(t, closeErr)
}

// TestRedisManagerConfiguration tests various configuration scenarios
func TestRedisManagerConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "localhost without password",
			addr:        "localhost:6379",
			password:    "",
			expectError: false,
		},
		{
			name:        "localhost with password",
			addr:        "localhost:6379",
			password:    "secret",
			expectError: false,
		},
		{
			name:        "empty address",
			addr:        "",
			password:    "",
			expectError: true,
			errorMsg:    "redis address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				RedisAddr:     tt.addr,
				RedisPassword: tt.password,
			}

			manager, err := database.NewRedisManager(cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.NotNil(t, manager.GetClient())

				// Clean up
				closeErr := manager.Close()
				assert.NoError(t, closeErr)
			}
		})
	}
}
