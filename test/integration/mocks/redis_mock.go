package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
)

// MockRedisManager is a mock implementation of the RedisManager
type MockRedisManager struct {
	GetClientFunc func() database.RedisClient
	PingFunc      func(ctx context.Context) error
	CloseFunc     func() error

	// Configuration for mock behavior
	ShouldFailPing  bool
	ShouldFailClose bool
	CustomError     error
	MockClient      *MockRedisClient
}

// NewMockRedisManager creates a new mock Redis manager
func NewMockRedisManager() *MockRedisManager {
	return &MockRedisManager{
		MockClient: NewMockRedisClient(),
	}
}

// NewMockRedisManagerWithError creates a mock that returns errors
func NewMockRedisManagerWithError(err error) *MockRedisManager {
	return &MockRedisManager{
		ShouldFailPing:  true,
		ShouldFailClose: true,
		CustomError:     err,
		MockClient:      NewMockRedisClientWithError(err),
	}
}

// GetClient returns the mock Redis client
func (m *MockRedisManager) GetClient() database.RedisClient {
	if m.GetClientFunc != nil {
		return m.GetClientFunc()
	}
	return m.MockClient
}

// Ping performs a mock ping operation
func (m *MockRedisManager) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}

	if m.ShouldFailPing {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock redis ping failed")
	}

	return nil
}

// Close performs a mock close operation
func (m *MockRedisManager) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}

	if m.ShouldFailClose {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock redis close failed")
	}

	return nil
}

// MockRedisClient is a mock implementation of the RedisClient interface
type MockRedisClient struct {
	// Storage for mock data
	data map[string]string
	ttls map[string]time.Time

	// Configuration for mock behavior
	ShouldFailPing  bool
	ShouldFailSet   bool
	ShouldFailGet   bool
	ShouldFailDel   bool
	ShouldFailKeys  bool
	ShouldFailClose bool
	CustomError     error
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
		ttls: make(map[string]time.Time),
	}
}

// NewMockRedisClientWithError creates a mock that returns errors
func NewMockRedisClientWithError(err error) *MockRedisClient {
	return &MockRedisClient{
		data:            make(map[string]string),
		ttls:            make(map[string]time.Time),
		ShouldFailPing:  true,
		ShouldFailSet:   true,
		ShouldFailGet:   true,
		ShouldFailDel:   true,
		ShouldFailKeys:  true,
		ShouldFailClose: true,
		CustomError:     err,
	}
}

// Ping returns a mock status command
func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)

	if m.ShouldFailPing {
		if m.CustomError != nil {
			cmd.SetErr(m.CustomError)
		} else {
			cmd.SetErr(fmt.Errorf("mock redis ping failed"))
		}
	} else {
		cmd.SetVal("PONG")
	}

	return cmd
}

// Close performs a mock close operation
func (m *MockRedisClient) Close() error {
	if m.ShouldFailClose {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock redis close failed")
	}
	return nil
}

// PoolStats returns mock pool statistics
func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	return &redis.PoolStats{
		Hits:       100,
		Misses:     10,
		Timeouts:   0,
		TotalConns: 5,
		IdleConns:  2,
		StaleConns: 0,
	}
}

// Set stores a key-value pair with TTL (mock implementation)
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)

	if m.ShouldFailSet {
		if m.CustomError != nil {
			cmd.SetErr(m.CustomError)
		} else {
			cmd.SetErr(fmt.Errorf("mock redis set failed"))
		}
		return cmd
	}

	// Store the value
	m.data[key] = fmt.Sprintf("%v", value)

	// Set TTL if specified
	if expiration > 0 {
		m.ttls[key] = time.Now().Add(expiration)
	}

	cmd.SetVal("OK")
	return cmd
}

// Get retrieves a value by key (mock implementation)
func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)

	if m.ShouldFailGet {
		if m.CustomError != nil {
			cmd.SetErr(m.CustomError)
		} else {
			cmd.SetErr(fmt.Errorf("mock redis get failed"))
		}
		return cmd
	}

	// Check if key has expired
	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		cmd.SetErr(redis.Nil)
		return cmd
	}

	// Get the value
	if value, exists := m.data[key]; exists {
		cmd.SetVal(value)
	} else {
		cmd.SetErr(redis.Nil)
	}

	return cmd
}

// Del deletes keys (mock implementation)
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)

	if m.ShouldFailDel {
		if m.CustomError != nil {
			cmd.SetErr(m.CustomError)
		} else {
			cmd.SetErr(fmt.Errorf("mock redis del failed"))
		}
		return cmd
	}

	deletedCount := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			delete(m.ttls, key)
			deletedCount++
		}
	}

	cmd.SetVal(deletedCount)
	return cmd
}

// Keys returns keys matching a pattern (mock implementation)
func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)

	if m.ShouldFailKeys {
		if m.CustomError != nil {
			cmd.SetErr(m.CustomError)
		} else {
			cmd.SetErr(fmt.Errorf("mock redis keys failed"))
		}
		return cmd
	}

	var matchingKeys []string
	for key := range m.data {
		// Simple pattern matching for tests (supports * wildcard)
		if pattern == "*" || key == pattern {
			matchingKeys = append(matchingKeys, key)
		}
	}

	cmd.SetVal(matchingKeys)
	return cmd
}

// MockCacheUtil is a mock implementation of the CacheUtil
type MockCacheUtil struct {
	SetFunc             func(key string, value string, ttl time.Duration) error
	GetFunc             func(key string) (string, error)
	DeleteFunc          func(key string) error
	DeleteByPatternFunc func(pattern string) error

	// Configuration for mock behavior
	ShouldFailSet           bool
	ShouldFailGet           bool
	ShouldFailDelete        bool
	ShouldFailDeletePattern bool
	CustomError             error

	// Internal storage for mock
	data map[string]string
	ttls map[string]time.Time
}

// NewMockCacheUtil creates a new mock cache utility
func NewMockCacheUtil() *MockCacheUtil {
	return &MockCacheUtil{
		data: make(map[string]string),
		ttls: make(map[string]time.Time),
	}
}

// NewMockCacheUtilWithError creates a mock that returns errors
func NewMockCacheUtilWithError(err error) *MockCacheUtil {
	return &MockCacheUtil{
		data:                    make(map[string]string),
		ttls:                    make(map[string]time.Time),
		ShouldFailSet:           true,
		ShouldFailGet:           true,
		ShouldFailDelete:        true,
		ShouldFailDeletePattern: true,
		CustomError:             err,
	}
}

// Set stores a key-value pair with TTL
func (m *MockCacheUtil) Set(key string, value string, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}

	if m.ShouldFailSet {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache set failed")
	}

	m.data[key] = value
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(ttl)
	}

	return nil
}

// Get retrieves a value by key
func (m *MockCacheUtil) Get(key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}

	if m.ShouldFailGet {
		if m.CustomError != nil {
			return "", m.CustomError
		}
		return "", fmt.Errorf("mock cache get failed")
	}

	// Check if key has expired
	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		return "", cacheutil.ErrCacheMiss
	}

	if value, exists := m.data[key]; exists {
		return value, nil
	}

	return "", cacheutil.ErrCacheMiss
}

// Delete removes a key
func (m *MockCacheUtil) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}

	if m.ShouldFailDelete {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache delete failed")
	}

	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

// DeleteByPattern removes keys matching a pattern
func (m *MockCacheUtil) DeleteByPattern(pattern string) error {
	if m.DeleteByPatternFunc != nil {
		return m.DeleteByPatternFunc(pattern)
	}

	if m.ShouldFailDeletePattern {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache delete by pattern failed")
	}

	for key := range m.data {
		if pattern == "*" || key == pattern {
			delete(m.data, key)
			delete(m.ttls, key)
		}
	}

	return nil
}
