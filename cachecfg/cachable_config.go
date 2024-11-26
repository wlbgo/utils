package cachecfg

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errUseOutdatedValue = errors.New("use outdated value")
var errDefaultUnimplemented = errors.New("default value unimplemented")

// TODO remove long time cache

// ValueFetcher defines the interface for fetching values
type ValueFetcher[T any] interface {
	// Key generates a key for the cache
	Key(ctx context.Context, args ...any) string

	// FetchValue fetches the value from the source, maybe multiple fetch inside
	FetchValue(ctx context.Context, args ...any) (T, error) // 考虑到有默认值的使用
}

type DefaultValueFetcher[T any] interface {
	// DefaultValue fetches the value default, if FetchValue failed
	DefaultValue(ctx context.Context, args ...any) (T, error) // 考虑到有默认值的使用
}

// singleCache represents a single cached value
type singleCache[T any] struct {
	Value      T
	ExpireTime time.Time
}

// CachableConfig represents a cacheable configuration
type CachableConfig[T any] struct {
	fetcher     ValueFetcher[T] // ValueFetcher interface
	TTL         time.Duration
	Cache       map[string]*singleCache[T]
	Mutex       sync.RWMutex
	ForceUpdate bool
}

func NewCacheCfg[T any](ttl time.Duration, forceUpdate bool, fetcher ValueFetcher[T]) *CachableConfig[T] {
	return &CachableConfig[T]{
		fetcher:     fetcher,
		TTL:         ttl,
		Cache:       make(map[string]*singleCache[T]),
		Mutex:       sync.RWMutex{},
		ForceUpdate: forceUpdate,
	}
}

// GetValue retrieves the value from the cache or fetches it if not present
func (c *CachableConfig[T]) GetValue(ctx context.Context, args ...any) (T, error) {
	c.Mutex.RLock()
	key := c.fetcher.Key(ctx, args...)
	if v, ok := c.Cache[key]; ok && v.ExpireTime.After(time.Now()) {
		c.Mutex.RUnlock()
		return v.Value, nil
	}
	c.Mutex.RUnlock()

	value, err := c.fetcher.FetchValue(ctx, args...)
	if err != nil {
		if f, ok := c.fetcher.(DefaultValueFetcher[T]); ok {
			value, err = f.DefaultValue(ctx, args...)
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
