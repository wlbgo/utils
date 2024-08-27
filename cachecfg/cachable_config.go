package rank

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var errUseOutdatedValue = errors.New("use outdated value")

// ValueFetcher defines the interface for fetching values
type ValueFetcher[T any] interface {
	Key(args ...any) string
	FetchValue(args ...any) (T, error) // 考虑到有默认值的使用
}

// singleCache represents a single cached value
type singleCache[T any] struct {
	Value      T
	ExpireTime time.Time
}

// CachableConfig represents a cacheable configuration
type CachableConfig[T any] struct {
	ValueFetcher[T] // ValueFetcher interface
	TTL            time.Duration
	Cache          map[string]*singleCache[T]
	Mutex          sync.RWMutex
	ForceUpdate    bool
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
	defer fmt.Printf("GetValue() key: %s, value: %v, err: %v\n", key, value, err)
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
