package sdk_inf

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type IRedis interface {
	Client(ctx context.Context) *redis.Client
	Set(ctx context.Context, key string, value any) error
	SetEx(ctx context.Context, key string, expiration time.Duration, value any) error
	SetNX(ctx context.Context, key string, expiration time.Duration, value any) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, key string) (int64, error)
	MSet(ctx context.Context, values ...any) (string, error)
	MSetNX(ctx context.Context, values ...any) (bool, error)
	MGet(ctx context.Context, key string) ([]any, error)
	HSet(ctx context.Context, key string, values ...any) error
	HSetEx(ctx context.Context, key string, expiration time.Duration, values ...any) error
	HSetNX(ctx context.Context, key string, field string, value any) error
	HGet(ctx context.Context, key string, field string) ([]byte, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HExists(ctx context.Context, key string, field string) (bool, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error)
	LPush(ctx context.Context, key string, values ...any) ([]string, error)
	SAdd(ctx context.Context, key string, members ...any) (int64, error)
	SIsMember(ctx context.Context, key string, member any) (bool, error)
	SRem(ctx context.Context, key string, member any) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error)
	ZRange(ctx context.Context, key string) ([]string, error)
	IncrBy(ctx context.Context, key string, value int) (int, error)
	TTL(ctx context.Context, key string) (int, error)
	Unlink(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
}
