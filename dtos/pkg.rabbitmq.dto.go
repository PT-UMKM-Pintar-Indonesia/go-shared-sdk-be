package sdk_dto

import (
	"context"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

type (
	RabbitClientOptions struct {
		Ctx        context.Context
		Url        string
		Urls       []string
		Vhost      string
		Heartbeat  time.Duration
		FrameSize  int
		ChannelMax uint16
		Secure     bool
		Timeout    int
		KeepAlive  int
		Cluster    bool
		Shuffle    bool
	}

	RabbitOptions struct {
		AppID         string
		UserID        string
		ExchangeName  string
		ExchangeType  string
		QueueName     string
		Ack           bool
		Concurrency   int
		ConsumerID    string
		Args          rabbitmq.Table
		Body          any
		ContentType   string
		Timestamp     time.Time
		Prefetch      int
		CorrelationID string
		ReplyTo       string
		Delivery      rabbitmq.Delivery
		Expired       string
		Timeout       time.Duration
	}

	PublisherOptions struct {
		CorrelationID string
		ReplyTo       string
		ContentType   string
		Timestamp     time.Time
	}
)
