package cachecfg

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisKeyValueFetcher struct {
	Rds *redis.Client
}

func (r *RedisKeyValueFetcher) FetchValue(args ...any) ([]byte, error) {
	key := r.Key(args)
	ctx := context.Background()
	s, err := r.Rds.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *RedisKeyValueFetcher) Key(args ...any) string {
	return args[0].(string)
}

func (r *RedisKeyValueFetcher) DefaultValue(_ ...any) (string, error) {
	panic("implement me")
}
