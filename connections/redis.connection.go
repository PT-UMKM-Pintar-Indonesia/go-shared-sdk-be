package sdk_con

import (
	"context"
	"errors"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	"github.com/redis/go-redis/v9"
)

type Option func(*redis.Options)

func newRedisClient(opt *sdk_dto.RedisClientOptions) (*redis.Client, error) {
	options, err := redis.ParseURL(opt.Addr)
	if err != nil {
		return nil, err
	}

	if opt.PoolSize < 1 {
		options.PoolSize = 20
	}

	if opt.MinIdleConns < 1 {
		options.MinIdleConns = 5
	}

	if opt.PoolFIFO {
		options.PoolFIFO = false
	}

	if opt.PoolTimeout < 1 {
		options.PoolTimeout = 30 * time.Second
	}

	if opt.MaxRetries < 1 {
		options.MaxRetries = 3
	}

	if opt.ReadTimeout < 1 {
		options.ReadTimeout = 10 * time.Second
	}

	if opt.WriteTimeout < 1 {
		options.WriteTimeout = 10 * time.Second
	}

	if opt.DialTimeout < 1 {
		options.DialTimeout = 10 * time.Second
	}

	if opt.MinRetryBackoff < 1 {
		options.MinRetryBackoff = 8 * time.Millisecond
	}

	if opt.MaxRetryBackoff < 1 {
		options.MaxRetryBackoff = 512 * time.Millisecond
	}

	if opt.ConnMaxIdleTime < 1 {
		options.ConnMaxIdleTime = 5 * time.Minute
	}

	if opt.ConnMaxLifetime < 1 {
		options.ConnMaxLifetime = 30 * time.Minute
	}

	if err := sdk_helper.NewTransform().SrcToDest(opt, options); err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	if err := ping(client, opt.Ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func newRedisSentinelClient(opt *sdk_dto.RedisClientOptions) (*redis.Client, error) {
	transform := sdk_helper.NewTransform()

	options := &redis.FailoverOptions{
		MasterName:       opt.MasterName,
		SentinelAddrs:    opt.SentinelAddrs,
		Username:         opt.Username,
		Password:         opt.Password,
		SentinelUsername: opt.SentinelUsername,
		SentinelPassword: opt.SentinelPassword,
	}

	if opt.PoolSize < 1 {
		options.PoolSize = 20
	}

	if opt.MinIdleConns < 1 {
		options.MinIdleConns = 5
	}

	if opt.PoolFIFO {
		options.PoolFIFO = false
	}

	if opt.PoolTimeout < 1 {
		options.PoolTimeout = 30 * time.Second
	}

	if opt.MaxRetries < 1 {
		options.MaxRetries = 3
	}

	if opt.ReadTimeout < 1 {
		options.ReadTimeout = 10 * time.Second
	}

	if opt.WriteTimeout < 1 {
		options.WriteTimeout = 10 * time.Second
	}

	if opt.DialTimeout < 1 {
		options.DialTimeout = 10 * time.Second
	}

	if opt.MinRetryBackoff < 1 {
		options.MinRetryBackoff = 8 * time.Millisecond
	}

	if opt.MaxRetryBackoff < 1 {
		options.MaxRetryBackoff = 512 * time.Millisecond
	}

	if opt.ConnMaxIdleTime < 1 {
		options.ConnMaxIdleTime = 5 * time.Minute
	}

	if opt.ConnMaxLifetime < 1 {
		options.ConnMaxLifetime = 30 * time.Minute
	}

	if err := transform.SrcToDest(options, opt); err != nil {
		return nil, err
	}

	client := redis.NewFailoverClient(options)

	if err := ping(client, opt.Ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func ping(c *redis.Client, ctx context.Context) error {
	return c.Ping(ctx).Err()
}

func RedisConnection(opt *sdk_dto.RedisClientOptions) (*redis.Client, error) {
	if opt.Addr == sdk_cons.EMPTY {
		return nil, errors.New("redis url is required")
	}

	if opt.Cluster {
		return newRedisSentinelClient(opt)
	}

	return newRedisClient(opt)
}
