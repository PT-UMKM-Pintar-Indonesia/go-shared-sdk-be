package sdk_con

import (
	"time"

	"github.com/redis/go-redis/v9"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func RedisConnection(req sdk_dto.Request[sdk_dto.Environtment]) (*redis.Client, error) {
	options, err := redis.ParseURL(req.Config.REDIS.URL)
	if err != nil {
		return nil, err
	}

	options.MaxRetries = 10
	options.PoolSize = 20
	options.PoolFIFO = true
	options.ReadTimeout = time.Duration(time.Second * 30)
	options.WriteTimeout = time.Duration(time.Second * 30)
	options.DialTimeout = time.Duration(time.Second * 60)
	options.MinRetryBackoff = time.Duration(time.Second * 60)
	options.MaxRetryBackoff = time.Duration(time.Second * 120)

	return redis.NewClient(options), nil
}
