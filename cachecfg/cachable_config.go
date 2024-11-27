package cachecfg

import (
	"errors"
	"sync"
	"time"
)

var errUseOutdatedValue = errors.New("use outdated value")
var errDefaultUnimplemented = errors.New("default value unimplemented")

// DefaultValueError is a custom error type to indicate that DefaultValue should be used
type DefaultValueError struct {
	Err error
}

func (e *DefaultValueError) Error() string {
	return e.Err.Error()
}

func (e *DefaultValueError) Unwrap() error {
	return e.Err
}

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

// CachableConfig represents a cacheable configuration
type CachableConfig[T any] struct {
	ValueFetcher[T] // ValueFetcher interface
	TTL             time.Duration
	Cache           map[string]*singleCache[T]
	Mutex           sync.RWMutex
	ForceUpdate     bool
}

func NewCacheCfg[T any](ttl time.Duration, forceUpdate bool) *CachableConfig[T] {
	return &CachableConfig[T]{
		TTL:         ttl,
		Cache:       make(map[string]*singleCache[T]),
		Mutex:       sync.RWMutex{},
		ForceUpdate: forceUpdate,
	}
}

// GetValue retrieves the value from the cache or fetches it if not present
func (c *CachableConfig[T]) GetValue(args ...any) (T, error) {
	c.Mutex.RLock()
	key := c.ValueFetcher.Key(args...)
	if v, ok := c.Cache[key]; ok && v.ExpireTime.After(time.Now()) {
		c.Mutex.RUnlock()
		return v.Value, nil
	}
	c.Mutex.RUnlock()

	value, err := c.FetchValue(args...)
	if err != nil {
		var defaultValueErr *DefaultValueError
		if errors.As(err, &defaultValueErr) {
			if defaultValueFetcher, ok := c.ValueFetcher.(DefaultValueFetcher[T]); ok {
				value, err = defaultValueFetcher.DefaultValue(args...)
			} else {
				err = errDefaultUnimplemented
			}
		}
	}
	if err != nil {
		if c.ForceUpdate {
			c.Mutex.Lock()
			delete(c.Cache, key)
			c.Mutex.Unlock()
			return value, err
		}
		if v, ok := c.Cache[key]; ok {
			return v.Value, errUseOutdatedValue
		} else {
			return value, err
		}
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Cache[key] = &singleCache[T]{
		Value:      value,
		ExpireTime: time.Now().Add(c.TTL),
	}
	return value, nil
}
