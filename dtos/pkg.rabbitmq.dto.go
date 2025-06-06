package sdk_dto

import (
	"time"

	"github.com/wagslane/go-rabbitmq"
)

type (
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
	}

	PublisherOptions struct {
		CorrelationID string
		ReplyTo       string
		ContentType   string
		Timestamp     time.Time
	}
)
