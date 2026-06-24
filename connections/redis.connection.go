package sdk_con

import (
	"context"
	"errors"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	"github.com/redis/go-redis/v9"
)

type Option func(*redis.Options)

func newRedisClient(url string, opts ...Option) (*redis.Client, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	options.PoolSize = 20
	options.MinIdleConns = 5
	options.PoolFIFO = false
	options.PoolTimeout = 30 * time.Second
	options.MaxRetries = 3
	options.ReadTimeout = 10 * time.Second
	options.WriteTimeout = 10 * time.Second
	options.DialTimeout = 10 * time.Second

	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func RedisConnection(url string) (*redis.Client, error) {
	if url == sdk_cons.EMPTY {
		return nil, errors.New("redis url is required")
	}

	return newRedisClient(url)
}

func RedisConnectionWithPool(url string, poolSize, minIdle int) (*redis.Client, error) {
	if url == sdk_cons.EMPTY {
		return nil, errors.New("redis url is required")
	}

	return newRedisClient(url, func(o *redis.Options) {
		o.PoolSize = poolSize
		o.MinIdleConns = minIdle
	})
}
