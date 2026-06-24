package sdk_con

import (
	"errors"
	"fmt"
	"time"

	"github.com/wagslane/go-rabbitmq"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
)

const (
	DefaultFrameSize  = 131072
	DefaultChannelMax = 256
	DefaultHeartbeat  = 30 * time.Second
)

func RabbitConnection(url, vhost string) (*rabbitmq.Conn, error) {
	if url == sdk_cons.EMPTY {
		return nil, errors.New("rabbitmq url is required")
	}

	heartbeat := DefaultHeartbeat

	conn, err := rabbitmq.NewConn(url,
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(heartbeat),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Vhost:      vhost,
			FrameSize:  DefaultFrameSize,
			ChannelMax: DefaultChannelMax,
			Heartbeat:  heartbeat,
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	return conn, nil
}
