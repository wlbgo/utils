package cachecfg

import (
	"errors"
	"sync"
	"time"
)

var errUseOutdatedValue = errors.New("use outdated value")
var errDefaultUnimplemented = errors.New("default value unimplemented")

// UseDefaultValue is a custom error to indicate that DefaultValue should be used
var UseDefaultValue = errors.New("use default value")

// ValueFetcher defines the interface for fetching values
type ValueFetcher[T any] interface {
	// Key generates a key for the cache
	Key(args ...any) string

	// FetchValue fetches the value from the source, maybe multiple fetch inside
	FetchValue(args ...any) (T, error)
}

// DefaultValueFetcher defines the interface for fetching default values
type DefaultValueFetcher[T any] interface {
	// DefaultValue fetches the value default, if FetchValue failed
	DefaultValue(args ...any) (T, error)
}

// singleCache represents a single cached value
type singleCache[T any] struct {
	Value      T
	ExpireTime time.Time
}

// Config represents a cacheable configuration
type Config[T any] struct {
	ValueFetcher[T] // ValueFetcher interface
	TTL             time.Duration
	Cache           map[string]*singleCache[T]
	Mutex           sync.RWMutex
	ForceUpdate     bool

	// use for clean cache
	stopChan      chan struct{}
	cleanInterval time.Duration

	// tracks keys that have an in-flight async refresh to prevent duplicate goroutines
	updatingMu sync.Mutex
	updating   map[string]bool
}

// NewCacheCfg creates a new Config
func NewCacheCfg[T any](ttl time.Duration, forceUpdate bool) *Config[T] {
	return &Config[T]{
		TTL:         ttl,
		Cache:       make(map[string]*singleCache[T]),
		Mutex:       sync.RWMutex{},
		ForceUpdate: forceUpdate,
		updating:    make(map[string]bool),
	}
}

// NewCacheCfgWithAutoClean creates a new Config with automatic cache cleaning
func NewCacheCfgWithAutoClean[T any](ttl time.Duration, forceUpdate bool, cleanInterval time.Duration) *Config[T] {
	c := NewCacheCfg[T](ttl, forceUpdate)
	if cleanInterval <= 0 {
		panic("cleanInterval must be greater than 0")
	}
	c.cleanInterval = cleanInterval
	c.stopChan = make(chan struct{})
	go c.startCleaner()
	return c
}

// GetValue retrieves the value from the cache or fetches it if not present
func (c *Config[T]) GetValue(args ...any) (T, error) {
	c.Mutex.RLock()
	key := c.ValueFetcher.Key(args...)
	if v, ok := c.Cache[key]; ok && v.ExpireTime.After(time.Now()) {
		c.Mutex.RUnlock()
		return v.Value, nil
	}
	c.Mutex.RUnlock()

	value, err := c.FetchValue(args...)
	if err != nil && errors.Is(err, UseDefaultValue) {
		if defaultValueFetcher, ok := c.ValueFetcher.(DefaultValueFetcher[T]); ok {
			value, err = defaultValueFetcher.DefaultValue(args...)
		} else {
			err = errDefaultUnimplemented
		}
	}
	if err != nil {
		if c.ForceUpdate {
			c.Mutex.Lock()
			delete(c.Cache, key)
			c.Mutex.Unlock()
			return value, err
		}
		c.Mutex.RLock()
		if v, ok := c.Cache[key]; ok {
			c.Mutex.RUnlock()
			return v.Value, errUseOutdatedValue
		}
		c.Mutex.RUnlock()
		return value, err
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Cache[key] = &singleCache[T]{
		Value:      value,
		ExpireTime: time.Now().Add(c.TTL),
	}
	return value, nil
}

// AsyncGetValue returns the cached value immediately (even if expired) and triggers
// a background goroutine to refresh it when the cache is stale. Only one goroutine
// per key will be refreshing at a time. Falls back to synchronous GetValue when no
// cached value exists at all (e.g. first call).
func (c *Config[T]) AsyncGetValue(args ...any) (T, error) {
	key := c.ValueFetcher.Key(args...)

	c.Mutex.RLock()
	v, ok := c.Cache[key]
	c.Mutex.RUnlock()

	if ok {
		if v.ExpireTime.Before(time.Now()) {
			c.triggerAsyncUpdate(key, args...)
			return v.Value, errUseOutdatedValue
		}
		return v.Value, nil
	}

	// No cached value yet — block synchronously so the caller gets a real value.
	return c.GetValue(args...)
}

// triggerAsyncUpdate starts a background goroutine to refresh the cache for the
// given key. If a refresh is already in flight for that key, it is a no-op.
func (c *Config[T]) triggerAsyncUpdate(key string, args ...any) {
	c.updatingMu.Lock()
	if c.updating[key] {
		c.updatingMu.Unlock()
		return
	}
	c.updating[key] = true
	c.updatingMu.Unlock()

	go func() {
		defer func() {
			c.updatingMu.Lock()
			delete(c.updating, key)
			c.updatingMu.Unlock()
		}()
		c.GetValue(args...) //nolint:errcheck
	}()
}

// startCleaner starts a goroutine that periodically cleans up expired cache entries
func (c *Config[T]) startCleaner() {
	ticker := time.NewTicker(c.cleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpiredCache()
		case <-c.stopChan:
			return
		}
	}
}

// cleanExpiredCache removes expired cache entries
func (c *Config[T]) cleanExpiredCache() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	now := time.Now()
	for key, cache := range c.Cache {
		if cache.ExpireTime.Before(now) {
			delete(c.Cache, key)
		}
	}
}

// StopCleaner stops the cache cleaner goroutine
func (c *Config[T]) StopCleaner() {
	if c.stopChan != nil {
		close(c.stopChan)
	}
}
