package pkg

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/lithammer/shortuuid"
	amqp "github.com/wagslane/go-rabbitmq"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type rabbitmq struct {
	ctx      context.Context
	rabbitmq *amqp.Conn
}

var delivery sync.Map = sync.Map{}

func NewRabbitMQ(ctx context.Context, con *amqp.Conn) sdk_inf.IRabbitMQ {
	return rabbitmq{ctx: ctx, rabbitmq: con}
}

func (p rabbitmq) Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error {
	if req.Option.ContentType == sdk_cons.EMPTY {
		req.Option.ContentType = "application/json"
	}

	if req.Option.AppID == sdk_cons.EMPTY {
		req.Option.AppID = shortuuid.New()
	}

	if time.Until(req.Option.Timestamp) < 1 {
		req.Option.Timestamp = time.Now().Local()
	}

	publisher, err := amqp.NewPublisher(p.rabbitmq,
		amqp.WithPublisherOptionsExchangeName(req.Option.ExchangeName),
		amqp.WithPublisherOptionsExchangeKind(req.Option.ExchangeType),
		amqp.WithPublisherOptionsExchangeDeclare,
		amqp.WithPublisherOptionsExchangeDurable,
		amqp.WithPublisherOptionsExchangeNoWait,
		// amqp.WithPublisherOptionsConfirm,
		amqp.WithPublisherOptionsExchangeArgs(req.Option.Args),
		amqp.WithPublisherOptionsLogging,
	)

	if err != nil {
		publisher.Close()
		return err
	}

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Option.Body)
	if err != nil {
		return err
	}

	// publisher.NotifyPublish(func(r amqp.Confirmation) {
	// 	if !r.Confirmation.Ack {
	// 		Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.QueueName)
	// 		return
	// 	}

	// 	Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.QueueName)
	// })

	err = publisher.PublishWithContext(p.ctx, bodyByte, []string{req.Option.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Option.ExchangeName),
		amqp.WithPublishOptionsContentType(req.Option.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Timestamp),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)
	defer publisher.Close()

	if err != nil {
		return err
	}

	return nil
}

func (p rabbitmq) Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d amqp.Delivery) (action amqp.Action)) {
	if req.Option.ConsumerID == sdk_cons.EMPTY {
		req.Option.ConsumerID = shortuuid.New()
	}

	if req.Option.Concurrency < 1 {
		req.Option.Concurrency = runtime.NumCPU() / 2
	}

	if req.Option.Prefetch < 1 {
		req.Option.Prefetch = 5
	}

	// if req.Option.Args == nil {
	// 	req.Option.Args = amqp.Table{
	// 		"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.Option.QueueName),
	// 		"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.Option.QueueName),
	// 	}
	// } else {
	// 	err := mergo.Merge(&req.Option.Args, amqp.Table{
	// 		"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.Option.QueueName),
	// 		"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.Option.QueueName),
	// 	})
	// 	Logrus(sdk_cons.ERROR, err)
	// }

	consumer, err := amqp.NewConsumer(p.rabbitmq, callback, req.Option.QueueName,
		amqp.WithConsumerOptionsExchangeName(req.Option.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.Option.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey: req.Option.QueueName,
			BindingOptions: amqp.BindingOptions{
				Declare: true,
				NoWait:  true,
				Args:    req.Option.Args,
			},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsExchangeDeclare,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(req.Option.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Option.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Option.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Option.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Option.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		Logrus(sdk_cons.ERROR, err)
		consumer.Close()
		return
	}
}

func (p rabbitmq) listeningConsumerRPC(req sdk_dto.RabbitOptions) (*amqp.Consumer, error) {
	if req.ConsumerID == sdk_cons.EMPTY {
		req.ConsumerID = shortuuid.New()
	}

	if req.Concurrency < 1 {
		req.Concurrency = runtime.NumCPU() / 2
	}

	if req.Prefetch < 1 {
		req.Prefetch = 5
	}

	if req.Args == nil {
		req.Args = amqp.Table{
			"x-max-length":              1000000,
			"x-message-ttl":             3600 * 1000,
			"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.ReplyTo),
			"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.ReplyTo),
		}
	} else {
		err := mergo.Merge(&req.Args, amqp.Table{
			"x-max-length":              1000000,
			"x-message-ttl":             3600 * 1000,
			"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.ReplyTo),
			"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.ReplyTo),
		})
		Logrus(sdk_cons.ERROR, err)
	}

	consumer, err := amqp.NewConsumer(p.rabbitmq, func(d amqp.Delivery) (action amqp.Action) {
		if d.CorrelationId == sdk_cons.EMPTY {
			return amqp.NackDiscard
		}

		if ch, ok := delivery.Load(d.CorrelationId); ok {
			ch.(chan []byte) <- d.Body
			delivery.Delete(d.CorrelationId)
		}

		return amqp.Ack
	}, req.ReplyTo,
		amqp.WithConsumerOptionsExchangeName(req.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.ExchangeType),
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerExclusive,
		amqp.WithConsumerOptionsQueueAutoDelete,
		amqp.WithConsumerOptionsConsumerName(req.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		consumer.Close()
		return nil, err
	}

	return consumer, nil
}

func (p rabbitmq) PublisherRPC(req sdk_dto.Request[sdk_dto.RabbitOptions]) ([]byte, error) {
	if req.Option.ContentType == sdk_cons.EMPTY {
		req.Option.ContentType = "application/json"
	}

	if req.Option.AppID == sdk_cons.EMPTY {
		req.Option.AppID = shortuuid.New()
	}

	if req.Option.CorrelationID == sdk_cons.EMPTY {
		req.Option.CorrelationID = shortuuid.New()
	}

	if req.Option.ReplyTo == sdk_cons.EMPTY {
		req.Option.ReplyTo = req.Option.CorrelationID
	}

	if time.Until(req.Option.Timestamp) < 1 {
		req.Option.Timestamp = time.Now().Local()
	}

	res := make(chan []byte, 1000000)
	delivery.Store(req.Option.CorrelationID, res)

	defer func() {
		delivery.Delete(req.Option.CorrelationID)
		close(res)
	}()

	consumer, err := p.listeningConsumerRPC(req.Option)
	if err != nil {
		consumer.Close()
		return nil, err
	}
	defer consumer.Close()

	publisher, err := amqp.NewPublisher(p.rabbitmq,
		amqp.WithPublisherOptionsExchangeName(req.Option.ExchangeName),
		amqp.WithPublisherOptionsExchangeKind(req.Option.ExchangeType),
		amqp.WithPublisherOptionsExchangeDeclare,
		amqp.WithPublisherOptionsExchangeDurable,
		amqp.WithPublisherOptionsExchangeNoWait,
		// amqp.WithPublisherOptionsConfirm,
		amqp.WithPublisherOptionsExchangeArgs(req.Option.Args),
		amqp.WithPublisherOptionsLogging,
	)

	if err != nil {
		publisher.Close()
		return nil, err
	}
	defer publisher.Close()

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Option.Body)
	if err != nil {
		return nil, err
	}

	// publisher.NotifyPublish(func(r amqp.Confirmation) {
	// 	if !r.Confirmation.Ack {
	// 		Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.QueueName)
	// 		return
	// 	}
	// 	Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.QueueName)
	// })

	err = publisher.PublishWithContext(p.ctx, bodyByte, []string{req.Option.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Option.ExchangeName),
		amqp.WithPublishOptionsCorrelationID(req.Option.CorrelationID),
		amqp.WithPublishOptionsReplyTo(req.Option.ReplyTo),
		amqp.WithPublishOptionsAppID(req.Option.AppID),
		amqp.WithPublishOptionsUserID(req.Option.UserID),
		amqp.WithPublishOptionsContentType(req.Option.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Timestamp),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(p.ctx, time.Second*time.Duration(3600000))
	defer cancel()

	select {
	case r := <-res:
		return r, nil

	case <-ctx.Done():
		return nil, errors.New("Timeout waiting for RPC response")
	}
}

func (p rabbitmq) ConsumerRPC(req sdk_dto.Request[sdk_dto.RabbitOptions], handler func(delivery amqp.Delivery) (action amqp.Action)) {
	if req.Option.ConsumerID == sdk_cons.EMPTY {
		req.Option.ConsumerID = shortuuid.New()
	}

	if req.Option.Concurrency < 1 {
		req.Option.Concurrency = runtime.NumCPU() / 2
	}

	if req.Option.Prefetch < 1 {
		req.Option.Prefetch = 5
	}

	// if req.Option.Args == nil {
	// 	req.Option.Args = amqp.Table{
	// 		"x-max-length":              1000000,
	// 		"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.Option.QueueName),
	// 		"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.Option.QueueName),
	// 	}
	// } else {
	// 	err := mergo.Merge(&req.Option.Args, amqp.Table{
	// 		"x-max-length":              1000000,
	// 		"x-dead-letter-exchange":    fmt.Sprintf("%s.dlx", req.Option.QueueName),
	// 		"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", req.Option.QueueName),
	// 	})
	// 	Logrus(sdk_cons.ERROR, err)
	// }

	consumer, err := amqp.NewConsumer(p.rabbitmq, handler, req.Option.QueueName,
		amqp.WithConsumerOptionsExchangeName(req.Option.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.Option.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey: req.Option.QueueName,
			BindingOptions: amqp.BindingOptions{
				Declare: true,
				NoWait:  true,
				Args:    req.Option.Args,
			},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(req.Option.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Option.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Option.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Option.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Option.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		Logrus(sdk_cons.ERROR, err)
		consumer.Close()
		return
	}
}

func (p rabbitmq) ReplyToDeliveryPublisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error {
	if req.Option.Delivery.Body == nil {
		return errors.New("Delivery body cannot be nil")
	}

	publisher, err := amqp.NewPublisher(p.rabbitmq,
		amqp.WithPublisherOptionsExchangeName(req.Option.ExchangeName),
		amqp.WithPublisherOptionsExchangeKind(req.Option.ExchangeType),
		amqp.WithPublisherOptionsExchangeDeclare,
		amqp.WithPublisherOptionsExchangeDurable,
		amqp.WithPublisherOptionsExchangeNoWait,
		amqp.WithPublisherOptionsExchangeArgs(req.Option.Args),
		amqp.WithPublisherOptionsLogging,
	)

	if err != nil {
		return err
	}

	// publisher.NotifyPublish(func(r amqp.Confirmation) {
	// 	if !r.Confirmation.Ack {
	// 		Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.Delivery.ReplyTo)
	// 		return
	// 	}

	// 	Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.Delivery.ReplyTo)
	// })

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Option.Body)
	if err != nil {
		return err
	}

	err = publisher.Publish(bodyByte, []string{req.Option.Delivery.ReplyTo},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsCorrelationID(req.Option.Delivery.CorrelationId),
		amqp.WithPublishOptionsAppID(req.Option.Delivery.AppId),
		amqp.WithPublishOptionsUserID(req.Option.Delivery.UserId),
		amqp.WithPublishOptionsContentType(req.Option.Delivery.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Delivery.Timestamp),
		amqp.WithPublishOptionsExpiration(req.Option.Expired),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)
	defer publisher.Close()

	if err != nil {
		return err
	}

	return nil
}
