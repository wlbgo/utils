package globaluserlimit

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	ErrBadConfig = errors.New("bad config")
)

type RedisHLL struct {
	Rds *redis.Client
	TTL time.Duration

	// 用于非精确计数，以下两个有一个为0则立即更新，只有两个都非0才会Cache
	batchUpdateSize int
	batchUpdateDur  time.Duration
}

type RedisHLLSingleKeyImpl struct {
	RedisHLL
}

func (r *RedisHLLSingleKeyImpl) Key(args ...any) string {
	return args[0].(string)
}

func NewRedisHLL(rds *redis.Client, ttl time.Duration, size, dur int) (*RedisHLL, error) {
	if size < 0 || dur < 0 {
		return nil, ErrBadConfig
	}

	return &RedisHLL{
		Rds:             rds,
		TTL:             ttl,
		batchUpdateSize: size,
		batchUpdateDur:  time.Duration(dur) * time.Second,
	}, nil
}

func (s *RedisHLL) TryInsert(ctx context.Context, key string, limit int, uid string) (bool, error) {
	if limit <= 0 {
		return false, nil
	}

	var luaRet interface{}
	var err error

	if s.TTL == 0 {
		luaScript := "local i = redis.call('PFCOUNT', KEYS[1]); if i >= tonumber(ARGV[1]) then return 1; end;" +
			" redis.call('PFADD', KEYS[1], ARGV[2]); return 0;"
		luaRet, err = s.Rds.Eval(ctx, luaScript, []string{key}, limit, uid).Result()
	} else {
		luaScript := "local i = redis.call('PFCOUNT', KEYS[1]); if i >= tonumber(ARGV[1]) then return 1; end; " +
			"redis.call('PFADD', KEYS[1], ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]); return 0;"
		luaRet, err = s.Rds.Eval(ctx, luaScript, []string{key}, limit, uid, s.TTL.Seconds()).Result()
	}
	if err != nil {
		return false, err
	}

	return luaRet.(int64) == 0, nil
}

func (s *RedisHLL) InsertUser(ctx context.Context, key, uid string) (int, error) {

	var lua string
	var ret interface{}
	var err error

	if s.TTL == 0 {
		lua = "local i = redis.call('PFADD', KEYS[1], ARGV[1]); return redis.call('PFCOUNT', KEYS[1]);"
		ret, err = s.Rds.Eval(ctx, lua, []string{key}, uid).Result()
	} else {
		lua = "local i = redis.call('PFADD', KEYS[1], ARGV[1]); redis.call('EXPIRE', KEYS[1], ARGV[2]); return redis.call('PFCOUNT', KEYS[1]);"
		ret, err = s.Rds.Eval(ctx, lua, []string{key}, uid, s.TTL.Seconds()).Result()
	}
	if err != nil {
		return 0, err
	}
	return int(ret.(int64)), nil
}

func (s *RedisHLL) IsLimited(ctx context.Context, key string, limit int) (bool, error) {
	i, err := s.Rds.PFCount(ctx, key).Result()
	return int(i) > limit, err
}

func (s *RedisHLL) expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.Rds.Expire(ctx, key, ttl).Result()
}
