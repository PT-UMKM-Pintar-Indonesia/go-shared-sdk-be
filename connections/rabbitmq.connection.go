package sdk_con

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/wagslane/go-rabbitmq"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func RabbitConnection[T any](req sdk_dto.Request[sdk_dto.Environtment[T]]) (*rabbitmq.Conn, error) {
	interval := time.Duration(time.Second * 5)

	return rabbitmq.NewConn(req.Config.RABBITMQ.URL,
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(interval),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Vhost:           req.Config.RABBITMQ.VSN,
			FrameSize:       http.DefaultMaxHeaderBytes * 5,
			Heartbeat:       interval,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}))
}
