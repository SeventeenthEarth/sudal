package cacheutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrCacheMiss tests that the ErrCacheMiss error is properly defined
func TestErrCacheMiss(t *testing.T) {
	assert.NotNil(t, ErrCacheMiss)
	assert.Equal(t, "cache miss: key not found or expired", ErrCacheMiss.Error())
}

// TestNewCacheUtil tests that NewCacheUtil creates a proper instance
func TestNewCacheUtil(t *testing.T) {
	// This test only verifies the constructor without requiring Redis
	cacheUtil := NewCacheUtil(nil)
	assert.NotNil(t, cacheUtil)
	assert.Nil(t, cacheUtil.redisManager) // Should be nil when passed nil
	assert.NotNil(t, cacheUtil.logger)    // Logger should be initialized
}
