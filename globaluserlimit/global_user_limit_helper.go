package globaluserlimit

import (
	"context"
	"sync"
	"time"
)

type LimitElem interface {
	// InsertUser 插入用户，返回当前数量
	InsertUser(ctx context.Context, key, uid string) (int, error)

	IsLimited(ctx context.Context, key string, limit int) (bool, error)

	// TryInsert 尝试插入，如果插入成功返回true，否则返回false
	TryInsert(ctx context.Context, key string, limit int, uid string) (bool, error)

	// Key 一个实例可以支持多个key分别限量，可以用于多个道具分别限量
	Key(args ...any) string
}

type limitStateCache struct {
	limitState    bool
	lastLimitTime time.Time
}

type UserLimitHelper struct {
	LimitElem

	TTL            time.Duration // 过期时间
	LimitCacheTime time.Duration // 如果受限制，则多长时间不做查询

	// cache 只用于被限制住，不影响未限制
	limitStateCache map[string]*limitStateCache
	mutex           sync.Mutex
}

func (h *UserLimitHelper) TryInsert(ctx context.Context, limit int, uid string, args ...any) (bool, error) {
	key := h.LimitElem.Key(args...)
	if h.checkLimitedCache(key) {
		return false, nil
	}

	ok, err := h.LimitElem.TryInsert(ctx, key, limit, uid)

	if err == nil {
		if !ok {
			h.updateLimitedCache(key)
		}
		return ok, nil
	} else {
		return false, err
	}
}

func (h *UserLimitHelper) CheckUserLimit(ctx context.Context, limit int, args ...any) (bool, error) {
	key := h.LimitElem.Key(args...)

	retInCache := h.checkLimitedCache(key)
	if retInCache {
		return true, nil
	}

	state, err := h.LimitElem.IsLimited(ctx, key, limit)
	if err == nil && state {
		h.updateLimitedCache(key)
	}
	return state, err
}

func (h *UserLimitHelper) UpdateUser(ctx context.Context, uid string, args ...any) (int, error) {
	key := h.LimitElem.Key(args...)
	return h.LimitElem.InsertUser(ctx, key, uid)
}

func (h *UserLimitHelper) checkLimitedCache(key string) bool {
	if h.LimitCacheTime == 0 {
		return false
	}
	if h.limitStateCache == nil {
		h.mutex.Lock()
		defer h.mutex.Unlock()
		if h.limitStateCache == nil {
			h.limitStateCache = make(map[string]*limitStateCache)
		}
	}
	if c, ok := h.limitStateCache[key]; ok && c.limitState && time.Since(c.lastLimitTime) < h.LimitCacheTime {
		return true
	}
	return false
}

func (h *UserLimitHelper) updateLimitedCache(key string) {
	if h.LimitCacheTime == 0 {
		return
	}

	if h.limitStateCache == nil {
		h.mutex.Lock()
		defer h.mutex.Unlock()
		if h.limitStateCache == nil {
			h.limitStateCache = make(map[string]*limitStateCache)
		}
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.limitStateCache[key] = &limitStateCache{
		limitState:    true,
		lastLimitTime: time.Now(),
	}
}
