package sdk_dto

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/auth"
)

type (
	RedisClientOptions struct {
		Network                      string
		Addr                         string
		ClientName                   string
		Dialer                       func(ctx context.Context, network, addr string) (net.Conn, error)
		OnConnect                    func(ctx context.Context, cn *redis.Conn) error
		Protocol                     int
		Username                     string
		Password                     string
		CredentialsProvider          func() (username string, password string)
		CredentialsProviderContext   func(ctx context.Context) (username string, password string, err error)
		StreamingCredentialsProvider auth.StreamingCredentialsProvider
		DB                           int
		MaxRetries                   int
		MinRetryBackoff              time.Duration
		MaxRetryBackoff              time.Duration
		DialTimeout                  time.Duration
		ReadTimeout                  time.Duration
		WriteTimeout                 time.Duration
		ContextTimeoutEnabled        bool
		PoolFIFO                     bool
		PoolSize                     int
		PoolTimeout                  time.Duration
		MinIdleConns                 int
		MaxIdleConns                 int
		MaxActiveConns               int
		ConnMaxIdleTime              time.Duration
		ConnMaxLifetime              time.Duration
		TLSConfig                    *tls.Config
		Limiter                      redis.Limiter
		readOnly                     bool
		DisableIndentity             bool
		DisableIdentity              bool
		IdentitySuffix               string
		UnstableResp3                bool
		MasterName                   string
		SentinelAddrs                []string
		SentinelUsername             string
		SentinelPassword             string
		RouteByLatency               bool
		RouteRandomly                bool
		ReplicaOnly                  bool
		UseDisconnectedReplicas      bool
	}
)
