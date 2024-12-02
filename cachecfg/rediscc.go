package cachecfg

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

var _ ValueFetcher[[]byte] = &RedisKeyValueFetcher{}

type RedisKeyValueFetcher struct {
	Rds *redis.Client

	// 配置
	EmptyArrayAsNil bool
}

func (r *RedisKeyValueFetcher) FetchValue(args ...any) ([]byte, error) {
	ctx := args[0].(context.Context)
	key := r.Key(args...)
	s, err := r.Rds.Get(ctx, key).Bytes()
	/* If taking redis.Nil as a common blank value, you can return nil as error.
	   This will not cause querying redis every time when calling `GetValue`.
	*/
	if r.EmptyArrayAsNil && errors.Is(err, redis.Nil) {
		return make([]byte, 0), nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *RedisKeyValueFetcher) Key(args ...any) string {
	// args[0] is context.Context
	return args[1].(string)
}
