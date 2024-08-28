package globaluserlimit

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gotest.tools/assert"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	key = "test_hll_limit_test_key"
)

func TestRedisHLL_Expire(t *testing.T) {

	rds := newTestRedis()
	hll, err := NewRedisHLL(rds, 200*time.Second, 0, 0)
	if err != nil {
		panic(err)
	}

	ulh := UserLimitHelper{
		LimitElem: &RedisHLLSingleKeyImpl{
			RedisHLL: *hll,
		},
		TTL:            0,
		LimitCacheTime: 0,
	}

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	rds.Del(ctx, key)

	ulh.InsertUser(context.Background(), key, "uid1")
	ttl, _ := rds.TTL(ctx, key).Result()
	t.Logf("ttl: %v, %v\n", ttl, hll.TTL)
	assert.Assert(t, ttl > 0 && ttl <= hll.TTL)

}

func TestRedisHLL_TryInsert(t *testing.T) {

	rds := newTestRedis()
	hll, err := NewRedisHLL(rds, 0, 0, 0)
	if err != nil {
		panic(err)
	}

	ulh := UserLimitHelper{
		LimitElem: &RedisHLLSingleKeyImpl{
			RedisHLL: *hll,
		},
		TTL:            0,
		LimitCacheTime: 0,
	}

	wg := sync.WaitGroup{}
	aiOK := atomic.Int32{}
	aiFail := atomic.Int32{}

	rds.Del(context.Background(), key)
	defer rds.Del(context.Background(), key)

	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			ok, err := ulh.TryInsert(context.Background(), 10, "uid"+strconv.Itoa(i), key)
			if err != nil {
				t.Errorf("error: %v\n", err)
			} else if ok {
				aiOK.Add(1)
			} else {
				aiFail.Add(1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Assert(t, aiOK.Load() == 10 && aiFail.Load() == 90)
}

func TestRedisHLL_IsLimited1(t *testing.T) {

	rds := newTestRedis()
	hll, err := NewRedisHLL(rds, 0, 0, 0)
	if err != nil {
		panic(err)
	}

	ulh := UserLimitHelper{
		LimitElem: &RedisHLLSingleKeyImpl{
			RedisHLL: *hll,
		},
		TTL:            0,
		LimitCacheTime: 0,
	}

	rds.Del(context.Background(), key)
	defer rds.Del(context.Background(), key)
	ctx := context.Background()

	limit := rand.Int() % 100
	for i := 0; i < 100; i++ {
		ulh.InsertUser(ctx, key, "uid"+strconv.Itoa(i))
		limited, err := ulh.IsLimited(ctx, key, limit)
		assert.NilError(t, err)
		if i < limit {
			assert.Assert(t, !limited)
		} else {
			assert.Assert(t, limited)
		}

	}
}

func newTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "test",
	})

}
