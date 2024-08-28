package globaluserlimit

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisHash struct {
	Rds *redis.Client
	TTL time.Duration

	// 用于非精确计数，以下两个有一个为0则立即更新，只有两个都非0才会Cache
	batchUpdateSize int
	batchUpdateDur  time.Duration
}

func NewRedisHash(rds *redis.Client, ttl time.Duration, size, dur int) (*RedisHash, error) {
	if size < 0 || dur < 0 {
		return nil, ErrBadConfig
	}

	return &RedisHash{
		Rds:             rds,
		TTL:             ttl,
		batchUpdateSize: size,
		batchUpdateDur:  time.Duration(dur) * time.Second,
	}, nil
}

func (s *RedisHash) TryInsert(ctx context.Context, key string, limit int, uid string) (bool, error) {
	if limit <= 0 {
		return false, nil
	}

	var luaRet interface{}
	var err error
	if s.TTL == 0 {
		luaScript := "local i = redis.call('HLEN', KEYS[1]); if i > tonumber(ARGV[1]) then return 1; end; " +
			"redis.call('HSET', KEYS[1], ARGV[2], ARGV[3]); return 0;"
		luaRet, err = s.Rds.Eval(ctx, luaScript, []string{key}, limit, uid).Result()
	} else {
		luaScript := "local i = redis.call('HLEN', KEYS[1]); if i > tonumber(ARGV[1]) then return 1; end; " +
			"redis.call('HSET', KEYS[1], ARGV[2], ARGV[3]); redis.call('EXPIRE', KEYS[1], ARGV[4]); return 0;"
		luaRet, err = s.Rds.Eval(ctx, luaScript, []string{key}, limit, uid, s.TTL).Result()
	}
	if err != nil {
		return false, err
	}

	return luaRet.(int64) == 0, nil
}

func (s *RedisHash) InsertUser(ctx context.Context, key, uid string) (int, error) {
	i, err := s.Rds.PFAdd(ctx, key, uid).Result()
	if err != nil {
		return 0, err
	}
	if s.TTL == 0 {
		return int(i), nil
	}
	_, err = s.Expire(ctx, key, s.TTL)
	return int(i), err
}

func (s *RedisHash) IsLimited(ctx context.Context, key string, limit int) (bool, error) {
	i, err := s.Rds.PFCount(ctx, key).Result()
	return int(i) > limit, err
}

func (s *RedisHash) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.Rds.Expire(ctx, key, ttl).Result()
}
