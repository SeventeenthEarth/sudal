# Cache Utility Documentation

## Overview

The Cache Utility provides a simple key-value caching interface using Redis as the backend. It supports basic CRUD operations with optional Time-To-Live (TTL) functionality and is designed to be thread-safe for concurrent operations.

## Features

- **Basic CRUD Operations**: Set, Get, Delete operations for string key-value pairs
- **TTL Support**: Optional expiration time for cache entries
- **Error Handling**: Distinguishable error types, including `ErrCacheMiss` for missing/expired keys
- **Thread Safety**: Safe for concurrent access across multiple goroutines
- **Dependency Injection**: Integrated with the application's DI system
- **Structured Logging**: Comprehensive logging for debugging and monitoring

## API Reference

### Core Functions

#### `Set(key string, value string, ttl time.Duration) error`

Stores a key-value pair with an optional TTL.

- If `ttl` is zero or negative, the key persists indefinitely
- Returns an error if the operation fails

#### `Get(key string) (string, error)`

Retrieves the value for a given key.

- Returns `ErrCacheMiss` if the key is not found or has expired
- Returns the value and nil error if successful

#### `Delete(key string) error`

Removes a key-value pair from the cache.

- Returns nil even if the key doesn't exist
- Returns an error only if the operation fails

#### `DeleteByPattern(pattern string) error`

Deletes all keys matching a pattern (useful for test cleanup).

- Uses Redis KEYS command with the provided pattern
- Returns an error if the operation fails

### Error Types

#### `ErrCacheMiss`

A sentinel error returned when a key is not found or has expired.

```go
if errors.Is(err, cacheutil.ErrCacheMiss) {
    // Handle cache miss
}
```

## Usage Examples

### Basic Operations

```go
// Initialize cache utility through dependency injection
cacheUtil, err := di.InitializeCacheUtil()
if err != nil {
    log.Fatal("Failed to initialize cache utility:", err)
}

// Set a key without TTL
err = cacheUtil.Set("user:123", "john_doe", 0)
if err != nil {
    log.Error("Failed to set cache key:", err)
}

// Get a key
value, err := cacheUtil.Get("user:123")
if errors.Is(err, cacheutil.ErrCacheMiss) {
    log.Info("Key not found in cache")
} else if err != nil {
    log.Error("Failed to get cache key:", err)
} else {
    log.Info("Retrieved value:", value)
}

// Delete a key
err = cacheUtil.Delete("user:123")
if err != nil {
    log.Error("Failed to delete cache key:", err)
}
```

### TTL Operations

```go
// Set a key with 5-minute TTL
err = cacheUtil.Set("session:abc123", "user_data", 5*time.Minute)
if err != nil {
    log.Error("Failed to set cache key with TTL:", err)
}

// The key will automatically expire after 5 minutes
```

### Test Cleanup

```go
// Clean up test keys (useful in tests)
err = cacheUtil.DeleteByPattern("test_*")
if err != nil {
    log.Error("Failed to cleanup test keys:", err)
}
```

## Configuration

The cache utility uses the existing Redis configuration from the application's config system:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=2
REDIS_POOL_TIMEOUT=4
REDIS_IDLE_TIMEOUT=300
REDIS_DIAL_TIMEOUT=5
REDIS_READ_TIMEOUT=3
REDIS_WRITE_TIMEOUT=3
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF=8
REDIS_MAX_RETRY_BACKOFF=512
```

## Testing

### E2E Tests

The cache utility includes comprehensive End-to-End tests that verify:

1. **Basic CRUD Operations**
   - Setting keys with and without TTL
   - Getting existing and non-existent keys
   - Deleting existing and non-existent keys

2. **TTL Functionality**
   - Keys accessible before TTL expiration
   - Keys return `ErrCacheMiss` after TTL expiration

3. **Concurrent Operations**
   - Multiple goroutines performing operations simultaneously
   - Each goroutine uses distinct keys to avoid conflicts
   - Verifies thread safety and data integrity

### Running E2E Tests

```bash
# Start Redis (required for E2E tests)
docker-compose up -d redis

# Run cache utility E2E tests
make test.e2e

# Or run specific cache tests
go test -v ./test/e2e/ -run TestCacheUtility
```

### Test Key Management

E2E tests use a consistent prefix (`e2e_test_cache_`) for all test keys to:

- Avoid conflicts with production data
- Enable easy cleanup after tests
- Maintain test isolation

## Architecture

### Dependency Injection

The cache utility is integrated with the application's Wire-based dependency injection system:

```go
// Wire provider sets
var CacheSet = wire.NewSet(
    ProvideConfig,
    ProvideRedisManager,
    ProvideCacheUtil,
)

// Initialization
func InitializeCacheUtil() (*cacheutil.CacheUtil, error) {
    wire.Build(CacheSet)
    return nil, nil // Wire will fill this in
}
```

### Error Handling

The utility follows Go best practices for error handling:

- Sentinel errors for specific conditions (`ErrCacheMiss`)
- Wrapped errors with context for debugging
- Structured logging for operational visibility

### Thread Safety

The cache utility is thread-safe because:

- The underlying Redis client (`go-redis`) is thread-safe
- No shared mutable state in the cache utility itself
- Each operation is atomic at the Redis level

## Best Practices

1. **Error Handling**: Always check for `ErrCacheMiss` specifically when handling cache misses
2. **TTL Usage**: Use appropriate TTL values based on data freshness requirements
3. **Key Naming**: Use consistent, hierarchical key naming conventions (e.g., `user:123`, `session:abc`)
4. **Test Isolation**: Use prefixed keys in tests and clean up after test completion
5. **Monitoring**: Monitor cache hit/miss ratios and operation latencies in production

## Future Enhancements

Potential future improvements:

- Support for complex data types (JSON serialization)
- Batch operations (MGET, MSET)
- Cache statistics and metrics
- Distributed locking primitives
- Pub/Sub functionality
