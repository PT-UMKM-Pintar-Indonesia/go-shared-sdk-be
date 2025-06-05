package pkg

import (
	"context"
	"time"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/interfaces"
	goredis "github.com/redis/go-redis/v9"
)

type redis struct {
	redis *goredis.Client
	ctx   context.Context
}

func NewRedis(ctx context.Context, con *goredis.Client) (sdk_inf.IRedis, error) {
	return &redis{redis: con, ctx: ctx}, nil
}

func (p redis) SetEx(key string, expiration time.Duration, value any) error {
	cmd := p.redis.SetEx(p.ctx, key, value, expiration)

	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

func (p redis) Get(key string) ([]byte, error) {
	cmd := p.redis.Get(p.ctx, key)

	if err := cmd.Err(); err != nil {
		return nil, err
	}

	res := cmd.Val()
	return []byte(res), nil
}

func (p redis) Del(key string) (int64, error) {
	cmd := p.redis.Del(p.ctx, key)

	if err := cmd.Err(); err != nil {
		return 0, err
	}

	return cmd.Val(), nil
}

func (p redis) Exists(key string) (int64, error) {
	cmd := p.redis.Exists(p.ctx, key)

	if err := cmd.Err(); err != nil {
		return 0, err
	}

	return cmd.Val(), nil
}

func (p redis) HSet(key string, values ...any) error {
	cmd := p.redis.HSet(p.ctx, key, values...)

	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

func (p redis) HSetEx(key string, expiration time.Duration, values ...any) error {
	cmd := p.redis.HSet(p.ctx, key, values)
	p.redis.Expire(p.ctx, key, expiration)

	if err := cmd.Err(); err != nil {
		return err
	}

	return nil
}

func (p redis) HGet(key, field string) ([]byte, error) {
	cmd := p.redis.HGet(p.ctx, key, field)

	if err := cmd.Err(); err != nil {
		return nil, err
	}

	res := cmd.Val()
	return []byte(res), nil
}

func (p redis) HGetAll(key string) (map[string]string, error) {
	cmd := p.redis.HGetAll(p.ctx, key)

	if err := cmd.Err(); err != nil {
		return nil, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) HExists(key, field string) (bool, error) {
	cmd := p.redis.HExists(p.ctx, key, field)

	if err := cmd.Err(); err != nil {
		return false, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) HDel(key string, fields ...string) (int64, error) {
	cmd := p.redis.HDel(p.ctx, key, fields...)

	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) HIncrByFloat(key, field string, incr float64) (float64, error) {
	cmd := p.redis.HIncrByFloat(p.ctx, key, field, incr)

	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) LPush(key string, values ...any) ([]string, error) {
	cmd := p.redis.LPush(p.ctx, key, values)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	cmdl := p.redis.LRange(p.ctx, key, 0, cmd.Val())
	if err := cmdl.Err(); err != nil {
		return nil, err
	}

	res := cmdl.Val()
	return res, nil
}

func (p redis) SAdd(key string, members ...any) (int64, error) {
	cmd := p.redis.SAdd(p.ctx, key, members...)
	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) SIsMember(key string, member any) (bool, error) {
	cmdl := p.redis.SIsMember(p.ctx, key, member)
	if err := cmdl.Err(); err != nil {
		return false, err
	}

	res := cmdl.Val()
	return res, nil
}

func (p redis) SRem(key string, member any) (int64, error) {
	cmdl := p.redis.SRem(p.ctx, key, member)
	if err := cmdl.Err(); err != nil {
		return -1, err
	}

	res := cmdl.Val()
	return res, nil
}

func (p redis) SMembers(key string) ([]string, error) {
	cmdl := p.redis.SMembers(p.ctx, key)
	if err := cmdl.Err(); err != nil {
		return nil, err
	}

	res := cmdl.Val()
	return res, nil
}

func (p redis) ZAdd(key string, members ...goredis.Z) (int64, error) {
	cmd := p.redis.ZAdd(p.ctx, key)
	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return res, nil
}

func (p redis) ZRange(key string) ([]string, error) {
	cmdl := p.redis.ZRange(p.ctx, key, 0, -1)
	if err := cmdl.Err(); err != nil {
		return nil, err
	}

	res := cmdl.Val()
	return res, nil
}

func (p redis) IncrBy(key string, value int) (int, error) {
	cmd := p.redis.IncrBy(p.ctx, key, int64(value))

	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return int(res), nil
}

func (p redis) TTL(key string) (int, error) {
	cmd := p.redis.TTL(p.ctx, key)

	if err := cmd.Err(); err != nil {
		return -1, err
	}

	res := cmd.Val()
	return int(res.Seconds()), nil
}
