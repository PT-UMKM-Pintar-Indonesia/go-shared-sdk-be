package sdk_pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk_con "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/connections"
	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	goredis "github.com/redis/go-redis/v9"
)

type redis struct {
	redis *goredis.Client
}

func NewRedis(ctx context.Context, opt *sdk_dto.RedisClientOptions) (sdk_inf.IRedis, *goredis.Client, error) {
	con, err := sdk_con.RedisConnection(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	return &redis{redis: con}, con, nil
}

func (p *redis) Client(ctx context.Context) *goredis.Client {
	return p.redis
}

func (p *redis) Set(ctx context.Context, key string, value any) error {
	if err := p.redis.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("redis set failed for key %s: %w", key, err)
	}
	return nil
}

func (p *redis) SetEx(ctx context.Context, key string, expiration time.Duration, value any) error {
	if err := p.redis.SetEx(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis setex failed for key %s: %w", key, err)
	}
	return nil
}

func (p *redis) SetNX(ctx context.Context, key string, expiration time.Duration, value any) error {
	if err := p.redis.SetNX(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis setnx failed for key %s: %w", key, err)
	}
	return nil
}

func (p *redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := p.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("redis get failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) Del(ctx context.Context, key string) (int64, error) {
	val, err := p.redis.Del(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis del failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) Exists(ctx context.Context, key string) (int64, error) {
	val, err := p.redis.Exists(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) MSet(ctx context.Context, values ...any) (string, error) {
	val, err := p.redis.MSet(ctx, values...).Result()
	if err != nil {
		return sdk_cons.EMPTY, fmt.Errorf("redis mset failed: %w", err)
	}
	return val, nil
}

func (p *redis) MSetNX(ctx context.Context, values ...any) (bool, error) {
	val, err := p.redis.MSetNX(ctx, values...).Result()
	if err != nil {
		return false, fmt.Errorf("redis msetnx failed: %w", err)
	}
	return val, nil
}

func (p *redis) MGet(ctx context.Context, key string) ([]any, error) {
	val, err := p.redis.MGet(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) HSet(ctx context.Context, key string, values ...any) error {
	if err := p.redis.HSet(ctx, key, values...).Err(); err != nil {
		return fmt.Errorf("redis hset failed for key %s: %w", key, err)
	}
	return nil
}

func (p *redis) HSetEx(ctx context.Context, key string, expiration time.Duration, values ...any) error {
	if len(values) < 1 {
		return errors.New("Invalid field and values")
	}

	if err := p.HSet(ctx, key, values); err != nil {
		return err
	}

	if err := p.redis.HExpire(ctx, key, expiration, values[0].(string)); err.Err() != nil {
		return err.Err()
	}

	return nil
}

func (p *redis) HSetNX(ctx context.Context, key string, field string, value any) error {
	if err := p.redis.HSetNX(ctx, key, field, value); err.Err() != nil {
		return err.Err()
	}

	return nil
}

func (p *redis) HGet(ctx context.Context, key, field string) ([]byte, error) {
	val, err := p.redis.HGet(ctx, key, field).Bytes()
	if err != nil {
		return nil, fmt.Errorf("redis hget failed for key %s field %s: %w", key, field, err)
	}
	return val, nil
}

func (p *redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	val, err := p.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis hgetall failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) HExists(ctx context.Context, key, field string) (bool, error) {
	val, err := p.redis.HExists(ctx, key, field).Result()
	if err != nil {
		return false, fmt.Errorf("redis hexists failed for key %s field %s: %w", key, field, err)
	}
	return val, nil
}

func (p *redis) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	val, err := p.redis.HDel(ctx, key, fields...).Result()
	if err != nil {
		return -1, fmt.Errorf("redis hdel failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	val, err := p.redis.HIncrByFloat(ctx, key, field, incr).Result()
	if err != nil {
		return -1, fmt.Errorf("redis hincrbyfloat failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) LPush(ctx context.Context, key string, values ...any) ([]string, error) {
	length, err := p.redis.LPush(ctx, key, values).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lpush failed for key %s: %w", key, err)
	}

	res, err := p.redis.LRange(ctx, key, 0, length-1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lrange failed for key %s: %w", key, err)
	}
	return res, nil
}

func (p *redis) SAdd(ctx context.Context, key string, members ...any) (int64, error) {
	val, err := p.redis.SAdd(ctx, key, members...).Result()
	if err != nil {
		return -1, fmt.Errorf("redis sadd failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	val, err := p.redis.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("redis sismember failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) SRem(ctx context.Context, key string, member any) (int64, error) {
	val, err := p.redis.SRem(ctx, key, member).Result()
	if err != nil {
		return -1, fmt.Errorf("redis srem failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) SMembers(ctx context.Context, key string) ([]string, error) {
	val, err := p.redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) ZAdd(ctx context.Context, key string, members ...goredis.Z) (int64, error) {
	val, err := p.redis.ZAdd(ctx, key, members...).Result()
	if err != nil {
		return -1, fmt.Errorf("redis zadd failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) ZRange(ctx context.Context, key string) ([]string, error) {
	val, err := p.redis.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zrange failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) IncrBy(ctx context.Context, key string, value int) (int, error) {
	val, err := p.redis.IncrBy(ctx, key, int64(value)).Result()
	if err != nil {
		return -1, fmt.Errorf("redis incrby failed for key %s: %w", key, err)
	}
	return int(val), nil
}

func (p *redis) TTL(ctx context.Context, key string) (int, error) {
	val, err := p.redis.TTL(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("redis ttl failed for key %s: %w", key, err)
	}
	return int(val.Seconds()), nil
}

func (p *redis) Unlink(ctx context.Context, key string) (int64, error) {
	val, err := p.redis.Unlink(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("redis unlink failed for key %s: %w", key, err)
	}
	return val, nil
}

func (p *redis) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	val, err := p.redis.Expire(ctx, key, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis expire failed for key %s: %w", key, err)
	}
	return val, nil
}
