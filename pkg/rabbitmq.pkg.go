package pkg

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/lithammer/shortuuid"
	amqp "github.com/wagslane/go-rabbitmq"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type rabbitmq struct {
	ctx      context.Context
	user     string
	rabbitmq *amqp.Conn
}

var (
	rabbitmqOptions []sdk_dto.RabbitOptions = []sdk_dto.RabbitOptions{}
	delivery        chan map[string][]byte  = make(chan map[string][]byte, 1)
)

func NewRabbitMQ(ctx context.Context, con *amqp.Conn) sdk_inf.IRabbitMQ {
	return rabbitmq{ctx: ctx, rabbitmq: con}
}

func (p rabbitmq) Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error {
	if req.Option.ContentType == "" {
		req.Option.ContentType = "application/json"
	}

	if req.Option.AppID == "" {
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
		amqp.WithPublisherOptionsExchangeArgs(req.Option.Args),
		amqp.WithPublisherOptionsLogging,
	)

	defer p.closeConnection(publisher, nil, p.rabbitmq)
	if err != nil {
		return err
	}

	parser := sdk_helper.NewParser()
	bodyByte, err := parser.Marshal(&req.Option.Body)
	if err != nil {
		return err
	}

	publisher.NotifyPublish(func(r amqp.Confirmation) {
		if !r.Confirmation.Ack {
			Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.QueueName)
			return
		}

		Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.QueueName)
	})

	err = publisher.PublishWithContext(p.ctx, bodyByte, []string{req.Option.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Option.ExchangeName),
		amqp.WithPublishOptionsContentType(req.Option.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Timestamp),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)

	if err != nil {
		return err
	}

	return nil
}

func (p rabbitmq) Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d amqp.Delivery) (action amqp.Action)) {
	if req.Option.ConsumerID == "" {
		req.Option.ConsumerID = shortuuid.New()
	}

	if req.Option.Concurrency < 1 {
		req.Option.Concurrency = runtime.NumCPU() / 2
	}

	if req.Option.Prefetch < 1 {
		req.Option.Prefetch = 5
	}

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
	defer p.closeConnection(nil, consumer, p.rabbitmq)

	if err != nil {
		Logrus(sdk_cons.ERROR, err)
		return
	}
}

func (p rabbitmq) listeningConsumerRPC(mutex *sync.RWMutex, req sdk_dto.RabbitOptions) error {
	if req.ConsumerID == "" {
		req.ConsumerID = shortuuid.New()
	}

	if req.Concurrency < 1 {
		req.Concurrency = runtime.NumCPU() / 2
	}

	if req.Prefetch < 1 {
		req.Prefetch = 5
	}

	consumer, err := amqp.NewConsumer(p.rabbitmq, func(d amqp.Delivery) (action amqp.Action) {
		if d.CorrelationId == "" {
			return amqp.NackDiscard
		}

		for _, opt := range rabbitmqOptions {
			if opt.CorrelationID != d.CorrelationId {
				return amqp.NackDiscard
			}
		}

		mutex.Lock()
		defer mutex.Unlock()

		delivery <- map[string][]byte{d.CorrelationId: d.Body}
		return amqp.Ack

	}, req.ReplyTo,
		amqp.WithConsumerOptionsExchangeName(req.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.ExchangeType),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsExchangeDeclare,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsQueueAutoDelete,
		amqp.WithConsumerOptionsConsumerName(req.CorrelationID),
		amqp.WithConsumerOptionsConsumerName(req.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Args),
		amqp.WithConsumerOptionsLogging,
	)
	defer p.closeConnection(nil, consumer, p.rabbitmq)

	if err != nil {
		return err
	}

	return nil
}

func (p rabbitmq) PublisherRPC(req sdk_dto.Request[sdk_dto.RabbitOptions]) ([]byte, error) {
	if req.Option.ContentType == "" {
		req.Option.ContentType = "application/json"
	}

	if req.Option.AppID == "" {
		req.Option.AppID = shortuuid.New()
	}

	if req.Option.CorrelationID == "" {
		req.Option.CorrelationID = shortuuid.New()
	}

	if req.Option.ReplyTo == "" {
		req.Option.ReplyTo = req.Option.CorrelationID
	}

	if time.Until(req.Option.Timestamp) < 1 {
		req.Option.Timestamp = time.Now().Local()
	}

	if req.Option.Expired == "" {
		req.Option.Expired = "60"
	}

	if len(rabbitmqOptions) > 0 {
		rabbitmqOptions = []sdk_dto.RabbitOptions{}
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

	defer p.closeConnection(publisher, nil, p.rabbitmq)
	if err != nil {
		return nil, err
	}

	parser := sdk_helper.NewParser()
	bodyByte, err := parser.Marshal(&req.Option.Body)
	if err != nil {
		return nil, err
	}

	publisher.NotifyPublish(func(r amqp.Confirmation) {
		if !r.Confirmation.Ack {
			Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.QueueName)
			return
		}

		Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.QueueName)
	})

	err = publisher.PublishWithContext(p.ctx, bodyByte, []string{req.Option.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Option.ExchangeName),
		amqp.WithPublishOptionsCorrelationID(req.Option.CorrelationID),
		amqp.WithPublishOptionsReplyTo(req.Option.ReplyTo),
		amqp.WithPublishOptionsAppID(req.Option.AppID),
		amqp.WithPublishOptionsUserID(req.Option.UserID),
		amqp.WithPublishOptionsContentType(req.Option.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Timestamp),
		amqp.WithPublishOptionsExpiration(req.Option.Expired),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)
	defer p.closeConnection(publisher, nil, p.rabbitmq)

	if err != nil {
		return nil, err
	}

	rabbitmqOptions = append(rabbitmqOptions, req.Option)
	mutex := new(sync.RWMutex)

	if err := p.listeningConsumerRPC(mutex, req.Option); err != nil {
		return nil, err
	}

	select {
	case d := <-delivery:
		return d[req.Option.CorrelationID], nil

	case <-time.After(time.Second * time.Duration(10)):
		return nil, errors.New("Timeout waiting for RPC response")
	}
}

func (p rabbitmq) ConsumerRPC(req sdk_dto.Request[sdk_dto.RabbitOptions], handler func(delivery amqp.Delivery) (action amqp.Action)) {
	if req.Option.ConsumerID == "" {
		req.Option.ConsumerID = shortuuid.New()
	}

	if req.Option.Concurrency < 1 {
		req.Option.Concurrency = runtime.NumCPU() / 2
	}

	if req.Option.Prefetch < 1 {
		req.Option.Prefetch = 5
	}

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
	defer p.closeConnection(nil, consumer, p.rabbitmq)

	if err != nil {
		Logrus(sdk_cons.ERROR, err)
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

	defer p.closeConnection(publisher, nil, p.rabbitmq)
	if err != nil {
		return err
	}

	publisher.NotifyPublish(func(r amqp.Confirmation) {
		if !r.Confirmation.Ack {
			Logrus(sdk_cons.ERROR, "Failed message delivery to: %s", req.Option.Delivery.ReplyTo)
			return
		}

		Logrus(sdk_cons.INFO, "Success message delivery to: %s", req.Option.Delivery.ReplyTo)
	})

	err = publisher.Publish(req.Option.Delivery.Body, []string{req.Option.Delivery.ReplyTo},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsCorrelationID(req.Option.Delivery.CorrelationId),
		amqp.WithPublishOptionsAppID(req.Option.Delivery.AppId),
		amqp.WithPublishOptionsUserID(req.Option.Delivery.UserId),
		amqp.WithPublishOptionsContentType(req.Option.Delivery.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Option.Delivery.Timestamp),
		amqp.WithPublishOptionsExpiration(req.Option.Expired),
		amqp.WithPublishOptionsHeaders(req.Option.Args),
	)

	if err != nil {
		return err
	}

	return nil
}

func (p rabbitmq) closeConnection(publisher *amqp.Publisher, consumer *amqp.Consumer, connection *amqp.Conn) {
	defer p.recovery()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM, syscall.SIGABRT, syscall.SIGUSR1)

	go func() {
		<-closeChan

		if consumer != nil {
			consumer.Close()
		}

		if publisher != nil {
			publisher.Close()
		}

		if connection != nil {
			connection.Close()
		}
	}()
}

func (p rabbitmq) recovery() {
	if err := recover(); err != nil {
		return
	}
}
