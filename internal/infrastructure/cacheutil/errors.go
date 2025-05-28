package cacheutil

import "errors"

// ErrCacheMiss is returned when a key is not found in the cache or has expired
var ErrCacheMiss = errors.New("cache miss: key not found or expired")
