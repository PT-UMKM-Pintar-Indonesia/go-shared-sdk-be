package sdk_con

import (
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/wagslane/go-rabbitmq"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

const (
	DefaultFrameSize  = 131072
	DefaultChannelMax = 5000
	DefaultHeartbeat  = 10
	DefaultTimeout    = 10
	DefaultKeepAlive  = 30
)

func rabbitConnection(opt *sdk_dto.RabbitClientOptions) (*rabbitmq.Conn, error) {
	if opt.Url == sdk_cons.EMPTY {
		return nil, errors.New("rabbitmq url is required")
	}

	if opt.ChannelMax < 1 {
		opt.ChannelMax = DefaultChannelMax
	}

	if opt.FrameSize < 1 {
		opt.FrameSize = DefaultFrameSize
	}

	if opt.Heartbeat < 1 {
		opt.Heartbeat = DefaultHeartbeat
	}

	if opt.Timeout < 1 {
		opt.Timeout = DefaultTimeout
	}

	if opt.KeepAlive < 1 {
		opt.KeepAlive = DefaultKeepAlive
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(opt.Timeout) * time.Second,
		KeepAlive: time.Duration(opt.KeepAlive) * time.Second,
	}

	return rabbitmq.NewConn(opt.Url,
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Dial:       dialer.Dial,
			Vhost:      opt.Vhost,
			Heartbeat:  time.Duration(opt.Heartbeat) * time.Second,
			FrameSize:  opt.FrameSize,
			ChannelMax: opt.ChannelMax,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opt.Secure,
			},
		}),
	)
}

func rabbitConnectionCluster(opt *sdk_dto.RabbitClientOptions) (*rabbitmq.Conn, error) {
	if len(opt.Urls) < 1 {
		return nil, errors.New("rabbitmq urls is required")
	}

	if opt.ChannelMax < 1 {
		opt.ChannelMax = DefaultChannelMax
	}

	if opt.FrameSize < 1 {
		opt.FrameSize = DefaultFrameSize
	}

	if opt.Heartbeat < 1 {
		opt.Heartbeat = DefaultHeartbeat
	}

	if opt.Timeout < 1 {
		opt.Timeout = DefaultTimeout
	}

	if opt.KeepAlive < 1 {
		opt.KeepAlive = DefaultKeepAlive
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(opt.Timeout) * time.Second,
		KeepAlive: time.Duration(opt.KeepAlive) * time.Second,
	}

	return rabbitmq.NewClusterConn(rabbitmq.NewStaticResolver(opt.Urls, opt.Shuffle),
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Dial:       dialer.Dial,
			Vhost:      opt.Vhost,
			Heartbeat:  time.Duration(opt.Heartbeat) * time.Second,
			FrameSize:  opt.FrameSize,
			ChannelMax: opt.ChannelMax,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opt.Secure,
			},
		}),
	)
}

func RabbitMQConnection(opt *sdk_dto.RabbitClientOptions) (*rabbitmq.Conn, error) {
	if len(opt.Urls) > 0 {
		return rabbitConnectionCluster(opt)
	}

	return rabbitConnection(opt)
}
