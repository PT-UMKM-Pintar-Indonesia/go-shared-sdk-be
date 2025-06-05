package sdk_inf

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/dtos"
	"github.com/wagslane/go-rabbitmq"
)

type IRabbitMQ interface {
	Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error
	Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d rabbitmq.Delivery) (action rabbitmq.Action))
}
