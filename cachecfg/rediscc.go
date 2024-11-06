package cachecfg

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisKeyValueFetcher struct {
	Rds *redis.Client
}

func (r *RedisKeyValueFetcher) FetchValue(args ...any) ([]byte, error) {
	ctx := args[0].(context.Context)
	key := r.Key(args...)
	s, err := r.Rds.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *RedisKeyValueFetcher) Key(args ...any) string {
	// args[0] is context.Context
	return args[1].(string)
}

func (r *RedisKeyValueFetcher) DefaultValue(_ ...any) ([]byte, error) {
	panic("implement me")
}
