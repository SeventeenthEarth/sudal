package mocks

import (
	"context"
	"fmt"
	"sync"
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

	// Test helper state
	isHealthy          bool
	isUnavailable      bool
	hasConnectionError bool
	hasTimeoutError    bool
}

// NewMockRedisManager creates a new mock Redis manager
func NewMockRedisManager() *MockRedisManager {
	return &MockRedisManager{
		MockClient: NewMockRedisClient(),
		isHealthy:  true, // Default to healthy state
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

// Test helper methods for MockRedisManager

// SetHealthyStatus configures the mock to be in healthy state
func (m *MockRedisManager) SetHealthyStatus() {
	m.isHealthy = true
	m.isUnavailable = false
	m.hasConnectionError = false
	m.hasTimeoutError = false
	m.ShouldFailPing = false
	m.ShouldFailClose = false
	m.CustomError = nil
	if m.MockClient != nil {
		m.MockClient.ShouldFailPing = false
		m.MockClient.ShouldFailSet = false
		m.MockClient.ShouldFailGet = false
		m.MockClient.ShouldFailDel = false
		m.MockClient.ShouldFailKeys = false
		m.MockClient.ShouldFailClose = false
		m.MockClient.CustomError = nil
	}
}

// SetUnavailableStatus configures the mock to simulate unavailable Redis
func (m *MockRedisManager) SetUnavailableStatus() {
	m.isHealthy = false
	m.isUnavailable = true
	m.hasConnectionError = false
	m.hasTimeoutError = false
	m.ShouldFailPing = true
	m.CustomError = fmt.Errorf("redis client is not available")
}

// SetConnectionError configures the mock to simulate connection errors
func (m *MockRedisManager) SetConnectionError() {
	m.isHealthy = false
	m.isUnavailable = false
	m.hasConnectionError = true
	m.hasTimeoutError = false
	m.ShouldFailPing = true
	m.ShouldFailClose = true
	m.CustomError = fmt.Errorf("connection error")
	if m.MockClient != nil {
		m.MockClient.ShouldFailPing = true
		m.MockClient.ShouldFailSet = true
		m.MockClient.ShouldFailGet = true
		m.MockClient.ShouldFailDel = true
		m.MockClient.ShouldFailKeys = true
		m.MockClient.ShouldFailClose = true
		m.MockClient.CustomError = fmt.Errorf("connection error")
	}
}

// SetTimeoutError configures the mock to simulate timeout errors
func (m *MockRedisManager) SetTimeoutError() {
	m.isHealthy = false
	m.isUnavailable = false
	m.hasConnectionError = false
	m.hasTimeoutError = true
	m.ShouldFailPing = true
	m.CustomError = fmt.Errorf("timeout error")
	if m.MockClient != nil {
		m.MockClient.ShouldFailPing = true
		m.MockClient.ShouldFailSet = true
		m.MockClient.ShouldFailGet = true
		m.MockClient.ShouldFailDel = true
		m.MockClient.ShouldFailKeys = true
		m.MockClient.CustomError = fmt.Errorf("timeout error")
	}
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

	// Internal storage for mock (with mutex for thread safety)
	mu   sync.RWMutex
	data map[string]string
	ttls map[string]time.Time

	// Redis manager reference for checking availability
	redisManager *MockRedisManager
}

// NewMockCacheUtil creates a new mock cache utility
func NewMockCacheUtil() *MockCacheUtil {
	return &MockCacheUtil{
		data: make(map[string]string),
		ttls: make(map[string]time.Time),
	}
}

// NewMockCacheUtilWithRedis creates a new mock cache utility with Redis manager
func NewMockCacheUtilWithRedis(redisManager *MockRedisManager) *MockCacheUtil {
	return &MockCacheUtil{
		data:         make(map[string]string),
		ttls:         make(map[string]time.Time),
		redisManager: redisManager,
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

	// Check Redis manager error states
	if m.redisManager != nil {
		if m.redisManager.hasConnectionError {
			return fmt.Errorf("connection error")
		}
		if m.redisManager.hasTimeoutError {
			return fmt.Errorf("timeout error")
		}
		if m.redisManager.isUnavailable {
			return fmt.Errorf("redis client is not available")
		}
	}

	if m.ShouldFailSet {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache set failed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

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

	// Validate key
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	// Check Redis manager error states
	if m.redisManager != nil {
		if m.redisManager.hasConnectionError {
			return "", fmt.Errorf("connection error")
		}
		if m.redisManager.hasTimeoutError {
			return "", fmt.Errorf("timeout error")
		}
		if m.redisManager.isUnavailable {
			return "", fmt.Errorf("redis client is not available")
		}
	}

	if m.ShouldFailGet {
		if m.CustomError != nil {
			return "", m.CustomError
		}
		return "", fmt.Errorf("mock cache get failed")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if key has expired
	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		// Need to upgrade to write lock for deletion
		m.mu.RUnlock()
		m.mu.Lock()
		// Double-check after acquiring write lock
		if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
			delete(m.data, key)
			delete(m.ttls, key)
		}
		m.mu.Unlock()
		m.mu.RLock()
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

	// Validate key
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Check Redis manager error states
	if m.redisManager != nil {
		if m.redisManager.hasConnectionError {
			return fmt.Errorf("connection error")
		}
		if m.redisManager.hasTimeoutError {
			return fmt.Errorf("timeout error")
		}
		if m.redisManager.isUnavailable {
			return fmt.Errorf("redis client is not available")
		}
	}

	if m.ShouldFailDelete {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache delete failed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

// DeleteByPattern removes keys matching a pattern
func (m *MockCacheUtil) DeleteByPattern(pattern string) error {
	if m.DeleteByPatternFunc != nil {
		return m.DeleteByPatternFunc(pattern)
	}

	// Validate pattern
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Check Redis availability
	if m.redisManager != nil && m.redisManager.isUnavailable {
		return fmt.Errorf("redis client is not available")
	}

	if m.ShouldFailDeletePattern {
		if m.CustomError != nil {
			return m.CustomError
		}
		return fmt.Errorf("mock cache delete by pattern failed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key := range m.data {
		if pattern == "*" || key == pattern {
			delete(m.data, key)
			delete(m.ttls, key)
		}
	}

	return nil
}

// Additional helper methods for MockRedisManager

// SetMockValue sets a value in the mock Redis client for testing
func (m *MockRedisManager) SetMockValue(key, value string) {
	if m.MockClient != nil {
		m.MockClient.data[key] = value
	}
}

// SetMockValueInCache sets a value directly in the cache util for testing
func (m *MockRedisManager) SetMockValueInCache(cache *MockCacheUtil, key, value string) {
	if cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		cache.data[key] = value
	}
}

// SetCacheMiss configures the mock to return cache miss for a specific key
func (m *MockRedisManager) SetCacheMiss(key string) {
	if m.MockClient != nil {
		delete(m.MockClient.data, key)
	}
}

// SetPatternKeys configures the mock to return specific keys for a pattern
func (m *MockRedisManager) SetPatternKeys(pattern string, keys []string) {
	if m.MockClient != nil {
		// Store the pattern and keys for later retrieval
		// For simplicity, we'll just ensure the keys exist in data
		for _, key := range keys {
			m.MockClient.data[key] = "pattern_value"
		}
	}
}
