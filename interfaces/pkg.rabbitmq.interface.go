package sdk_inf

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	"github.com/wagslane/go-rabbitmq"
)

type IRabbitMQ interface {
	Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error
	Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d rabbitmq.Delivery) (action rabbitmq.Action)) error
	PublisherRPC(req sdk_dto.Request[sdk_dto.RabbitOptions]) ([]byte, error)
	ConsumerRPC(req sdk_dto.Request[sdk_dto.RabbitOptions], handler func(delivery rabbitmq.Delivery) (action rabbitmq.Action)) error
	ReplyToDeliveryPublisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error
	Close() error
}
