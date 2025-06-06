package sdk_inf

import (
	"github.com/wagslane/go-rabbitmq"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

type IRabbitMQ interface {
	Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error
	Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d rabbitmq.Delivery) (action rabbitmq.Action))
	PublisherRPC(req sdk_dto.Request[sdk_dto.RabbitOptions]) ([]byte, error)
	ConsumerRPC(req sdk_dto.Request[sdk_dto.RabbitOptions], handler func(delivery rabbitmq.Delivery) (action rabbitmq.Action))
	ReplyToDeliveryPublisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error
}
