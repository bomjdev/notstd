package notstd

import (
	"errors"
	"sync"
	"time"
)

// CacheValue is a thread-safe cache for a single value with optional timeout and default function
type CacheValue[T any] struct {
	mu        sync.RWMutex
	value     T
	expiresAt time.Time
	defaultFn func() (T, error)
	timeout   time.Duration
	hasValue  bool
}

// NewCacheValue creates a new CacheValue instance
// timeout: cache entry expiration time (0 = no expiration)
// defaultFn: function to generate value if not found (nil = no default)
func NewCacheValue[T any](timeout time.Duration, defaultFn func() (T, error)) *CacheValue[T] {
	return &CacheValue[T]{
		timeout:   timeout,
		defaultFn: defaultFn,
	}
}

// isExpired checks if the value has expired
func (cv *CacheValue[T]) isExpired() bool {
	if cv.timeout == 0 || !cv.hasValue {
		return false
	}
	return time.Now().After(cv.expiresAt)
}

// GetNoDefault retrieves the value without using defaultFn
// Returns (value, true) if found and not expired, (zero, false) otherwise
func (cv *CacheValue[T]) GetNoDefault() (T, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if !cv.hasValue || cv.isExpired() {
		var zero T
		return zero, false
	}

	return cv.value, true
}

// GetDefault retrieves the value or uses defaultFn if not found
// Returns (value, true, nil) if found actual value
// Returns (value, false, nil) if defaultFn was used successfully
// Returns (zero, false, error) if defaultFn failed
func (cv *CacheValue[T]) GetDefault() (T, bool, error) {
	// Try to get cached value first
	val, ok := cv.GetNoDefault()
	if ok {
		return val, true, nil
	}

	// No defaultFn configured
	if cv.defaultFn == nil {
		var zero T
		return zero, false, nil
	}

	// Call defaultFn without holding the lock
	val, err := cv.defaultFn()
	if err != nil {
		var zero T
		return zero, false, err
	}

	// Store the value
	cv.Set(val)
	return val, false, nil
}

// Get is a helper that calls GetDefault if defaultFn is set, otherwise GetNoDefault
// Returns (value, ok, nil) if found or defaultFn was used successfully
// Returns (zero, false, error) if defaultFn failed
func (cv *CacheValue[T]) Get() (T, bool, error) {
	if cv.defaultFn != nil {
		val, ok, err := cv.GetDefault()
		if err != nil {
			return val, false, err
		}
		return val, ok || err == nil, nil
	}

	val, ok := cv.GetNoDefault()
	return val, ok, nil
}

// Set stores a value in the cache
// Returns true if an actual (non-expired) value was overwritten
func (cv *CacheValue[T]) Set(value T) bool {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	wasActual := cv.hasValue && !cv.isExpired()

	cv.value = value
	cv.hasValue = true
	if cv.timeout > 0 {
		cv.expiresAt = time.Now().Add(cv.timeout)
	}

	return wasActual
}

// Delete removes the value from the cache
// Returns true if an actual (non-expired) value was deleted
func (cv *CacheValue[T]) Delete() bool {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	wasActual := cv.hasValue && !cv.isExpired()

	var zero T
	cv.value = zero
	cv.hasValue = false
	cv.expiresAt = time.Time{}

	return wasActual
}

// Has checks if a value exists and is not expired
func (cv *CacheValue[T]) Has() bool {
	_, ok := cv.GetNoDefault()
	return ok
}

// Clear removes the value from the cache
func (cv *CacheValue[T]) Clear() {
	cv.Delete()
}

// Update refreshes the cached value by calling defaultFn
// Returns error if defaultFn is not set or if defaultFn fails
func (cv *CacheValue[T]) Update() error {
	if cv.defaultFn == nil {
		return errors.New("defaultFn is not set")
	}

	// Call defaultFn without holding the lock
	val, err := cv.defaultFn()
	if err != nil {
		return err
	}

	cv.Set(val)
	return nil
}

// Cache is a generic thread-safe in-memory cache service
type Cache[K comparable, V any] struct {
	mu        sync.RWMutex
	storage   map[K]*CacheValue[V]
	keyFn     KeyFn[V, K]
	defaultFn func(K) (V, error)
	timeout   time.Duration
}

// NewCache creates a new Cache instance
// timeout: cache entry expiration time (0 = no expiration)
// capacity: initial map capacity (0 = default)
// keyFn: function to extract key from value (nil = SetValue will not work)
// defaultFn: function to generate value if not found (nil = no default)
func NewCache[K comparable, V any](
	timeout time.Duration,
	capacity int,
	keyFn KeyFn[V, K],
	defaultFn func(K) (V, error),
) *Cache[K, V] {
	storage := make(map[K]*CacheValue[V])
	if capacity > 0 {
		storage = make(map[K]*CacheValue[V], capacity)
	}

	return &Cache[K, V]{
		storage:   storage,
		timeout:   timeout,
		keyFn:     keyFn,
		defaultFn: defaultFn,
	}
}

// getCacheValue retrieves or creates a CacheValue for a key
func (c *Cache[K, V]) getCacheValue(key K) *CacheValue[V] {
	c.mu.RLock()
	cv, ok := c.storage[key]
	c.mu.RUnlock()

	if ok {
		return cv
	}

	// Create new CacheValue with timeout if not exists
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	cv, ok = c.storage[key]
	if ok {
		return cv
	}

	cv = NewCacheValue[V](c.timeout, nil)
	c.storage[key] = cv
	return cv
}

// GetNoDefault retrieves a value from the cache by key without using defaultFn
// Returns (value, true) if found and not expired, (zero, false) otherwise
func (c *Cache[K, V]) GetNoDefault(key K) (V, bool) {
	c.mu.RLock()
	cv, ok := c.storage[key]
	c.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	return cv.GetNoDefault()
}

// GetDefault retrieves a value from the cache by key, or uses defaultFn if not found
// Returns (value, true, nil) if found actual value in cache
// Returns (value, false, nil) if defaultFn was used successfully
// Returns (zero, false, error) if defaultFn failed
func (c *Cache[K, V]) GetDefault(key K) (V, bool, error) {
	cv := c.getCacheValue(key)

	// If no defaultFn for the cache, just use CacheValue's GetNoDefault
	if c.defaultFn == nil {
		val, ok := cv.GetNoDefault()
		return val, ok, nil
	}

	// Try to get from CacheValue first
	val, ok := cv.GetNoDefault()
	if ok {
		return val, true, nil
	}

	// Call defaultFn without holding the lock
	val, err := c.defaultFn(key)
	if err != nil {
		var zero V
		return zero, false, err
	}

	// Store the value
	cv.Set(val)
	return val, false, nil
}

// Get is a helper that calls GetDefault if defaultFn is set, otherwise GetNoDefault
// Returns (value, ok, nil) if found in cache or defaultFn was used successfully
// Returns (zero, false, error) if defaultFn failed
func (c *Cache[K, V]) Get(key K) (V, bool, error) {
	if c.defaultFn != nil {
		val, ok, err := c.GetDefault(key)
		if err != nil {
			return val, false, err
		}
		return val, ok || err == nil, nil
	}

	val, ok := c.GetNoDefault(key)
	return val, ok, nil
}

// Set stores a value in the cache
// Returns true if an actual (non-expired) value was overwritten
func (c *Cache[K, V]) Set(key K, value V) bool {
	cv := c.getCacheValue(key)
	return cv.Set(value)
}

// SetValue stores a value in the cache using keyFn to extract the key
// Returns true if an actual (non-expired) value was overwritten
func (c *Cache[K, V]) SetValue(value V) bool {
	if c.keyFn == nil {
		return false
	}

	key := c.keyFn(value)
	return c.Set(key, value)
}

// Delete removes a value from the cache by key
// Returns true if an actual (non-expired) value was deleted
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	cv, existed := c.storage[key]
	if !existed {
		return false
	}

	wasActual := cv.Has()
	delete(c.storage, key)
	return wasActual
}

// GetKeyOf extracts the key from a value using keyFn
func (c *Cache[K, V]) GetKeyOf(value V) K {
	if c.keyFn != nil {
		return c.keyFn(value)
	}
	var zero K
	return zero
}

// Has checks if a key exists in the cache and is not expired
func (c *Cache[K, V]) Has(key K) bool {
	_, ok := c.GetNoDefault(key)
	return ok
}

// Clear removes all entries from the cache
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage = make(map[K]*CacheValue[V])
}

// Len returns the number of non-expired entries in the cache
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, cv := range c.storage {
		if cv.Has() {
			count++
		}
	}
	return count
}

// Keys returns all non-expired keys in the cache
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.storage))
	for k, cv := range c.storage {
		if cv.Has() {
			keys = append(keys, k)
		}
	}
	return keys
}

// Values returns all non-expired values in the cache
func (c *Cache[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.storage))
	for _, cv := range c.storage {
		if val, ok := cv.GetNoDefault(); ok {
			values = append(values, val)
		}
	}
	return values
}

// Range iterates over all non-expired key-value pairs in the cache
// If the callback returns false, iteration stops
func (c *Cache[K, V]) Range(fn func(key K, value V) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, cv := range c.storage {
		if val, ok := cv.GetNoDefault(); ok {
			if !fn(k, val) {
				break
			}
		}
	}
}

// SetMany stores multiple values in the cache
// Returns the count of actual values that were overwritten
func (c *Cache[K, V]) SetMany(items map[K]V) int {
	count := 0
	for k, v := range items {
		if c.Set(k, v) {
			count++
		}
	}
	return count
}

// DeleteMany removes multiple keys from the cache
// Returns the count of actual values that were deleted
func (c *Cache[K, V]) DeleteMany(keys []K) int {
	count := 0
	for _, k := range keys {
		if c.Delete(k) {
			count++
		}
	}
	return count
}

// Update refreshes the cached value for a key by calling defaultFn
// Returns error if defaultFn is not set or if defaultFn fails
func (c *Cache[K, V]) Update(key K) error {
	if c.defaultFn == nil {
		return errors.New("defaultFn is not set")
	}

	// Call defaultFn without holding the lock
	val, err := c.defaultFn(key)
	if err != nil {
		return err
	}

	c.Set(key, val)
	return nil
}
