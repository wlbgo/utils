package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	gul "github.com/wlbgo/utils/globaluserlimit"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "test",
	})
	hll, err := gul.NewRedisHLL(rds, 100*time.Second, 0, 0)
	if err != nil {
		panic(err)
	}

	ulh := gul.UserLimitHelper{
		LimitElem: &gul.RedisHLLSingleKeyImpl{
			RedisHLL: *hll,
		},
		TTL:            0,
		LimitCacheTime: 0,
	}

	wg := sync.WaitGroup{}
	aiOK := atomic.Int32{}
	aiFail := atomic.Int32{}

	rds.Del(context.Background(), "test_hll_limit")

	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			ok, err := ulh.TryInsert(context.Background(), 10, "uid"+strconv.Itoa(i), "test_hll_limit")
			if err != nil {
				fmt.Printf("error: %v\n", err)
			} else if ok {
				aiOK.Add(1)
			} else {
				aiFail.Add(1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("ok: %d, fail: %d, ttl: %v\n", aiOK.Load(), aiFail.Load(), rds.TTL(context.Background(), "test_hll_limit").Val())
}
