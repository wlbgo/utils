package cachecfg

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var _ ValueFetcher[[]byte] = &RedisKeyValueFetcher{}
var _ DefaultValueFetcher[[]byte] = &RedisKeyValueFetcher{}

type RedisKeyValueFetcher struct {
	Rds *redis.Client
}

func (r *RedisKeyValueFetcher) FetchValue(ctx context.Context, args ...any) ([]byte, error) {
	key := r.Key(ctx, args...)
	s, err := r.Rds.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *RedisKeyValueFetcher) Key(ctx context.Context, args ...any) string {
	return args[0].(string)
}

func (r *RedisKeyValueFetcher) DefaultValue(ctx context.Context, _ ...any) ([]byte, error) {
	panic("implement me")
}
