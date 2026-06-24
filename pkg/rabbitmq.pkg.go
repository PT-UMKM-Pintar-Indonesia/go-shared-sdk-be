package sdk_pkg

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/lithammer/shortuuid"
	amqp "github.com/wagslane/go-rabbitmq"
)

type rabbitmq struct {
	rabbitmq       *amqp.Conn
	consumers      sync.Map
	delivery       sync.Map
	parser         sdk_inf.IParser
	publisher      *amqp.Publisher
	replyQueueName string
}

var (
	cachedCPUCount int32
	cpuCountOnce   sync.Once
)

func getCPUCount() int {
	cpuCountOnce.Do(func() {
		atomic.StoreInt32(&cachedCPUCount, int32(runtime.NumCPU()))
	})
	return int(atomic.LoadInt32(&cachedCPUCount))
}

func NewRabbitMQ(con *amqp.Conn) (sdk_inf.IRabbitMQ, error) {
	publisher, err := amqp.NewPublisher(con, amqp.WithPublisherOptionsLogging)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &rabbitmq{
		rabbitmq:  con,
		parser:    sdk_helper.NewParser(),
		publisher: publisher,
	}, nil
}

func (p *rabbitmq) Close() {
	p.consumers.Range(func(key, value any) bool {
		if consumer, ok := value.(*amqp.Consumer); ok {
			consumer.Close()
		}
		p.consumers.Delete(key)
		return true
	})

	if p.publisher != nil {
		p.publisher.Close()
	}
}

func (p *rabbitmq) Publisher(ctx context.Context, opt sdk_dto.Request[*sdk_dto.RabbitOptions]) error {
	return p.PublisherCtx(ctx, opt)
}

func (p *rabbitmq) PublisherCtx(ctx context.Context, opt sdk_dto.Request[*sdk_dto.RabbitOptions]) error {
	payload := opt.Payload

	contentType := payload.ContentType
	if contentType == sdk_cons.EMPTY {
		contentType = "application/json"
	}

	appID := payload.AppID
	if appID == sdk_cons.EMPTY {
		appID = shortuuid.New()
	}

	timestamp := payload.Timestamp
	if time.Until(timestamp) < 1 {
		timestamp = time.Now().Local()
	}

	bodyByte, err := p.parser.Marshal(&payload.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	return p.publisher.PublishWithContext(ctx, bodyByte, []string{payload.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(payload.ExchangeName),
		amqp.WithPublishOptionsContentType(contentType),
		amqp.WithPublishOptionsTimestamp(timestamp),
		amqp.WithPublishOptionsHeaders(payload.Args),
		amqp.WithPublishOptionsAppID(appID),
	)
}

func (p *rabbitmq) Consumer(opt sdk_dto.Request[*sdk_dto.RabbitOptions], callback func(d amqp.Delivery) (action amqp.Action)) error {
	payload := opt.Payload

	consumerID := payload.ConsumerID
	if consumerID == sdk_cons.EMPTY {
		consumerID = shortuuid.New()
	}

	concurrency := payload.Concurrency
	if concurrency < 1 {
		concurrency = max(1, getCPUCount()/2)
	}

	prefetch := payload.Prefetch
	if prefetch < 1 {
		prefetch = 5
	}

	consumer, err := amqp.NewConsumer(p.rabbitmq, callback, payload.QueueName,
		amqp.WithConsumerOptionsExchangeName(payload.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(payload.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey:     payload.QueueName,
			BindingOptions: amqp.BindingOptions{Declare: true, NoWait: true, Args: payload.Args},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsExchangeDeclare,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(consumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(payload.Ack),
		amqp.WithConsumerOptionsConcurrency(concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(prefetch),
		amqp.WithConsumerOptionsQueueArgs(payload.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}

	p.consumers.Store(consumerID, consumer)
	return nil
}

func (p *rabbitmq) startRPCListener(replyQueueName string, opt *sdk_dto.RabbitOptions) error {
	if replyQueueName == sdk_cons.EMPTY {
		return errors.New("reply queue name cannot be empty")
	}

	p.replyQueueName = replyQueueName

	consumerID := shortuuid.New()
	concurrency := max(1, getCPUCount()/2)
	prefetch := 5

	_, err := amqp.NewConsumer(p.rabbitmq, func(d amqp.Delivery) (action amqp.Action) {
		ch, ok := p.delivery.Load(d.CorrelationId)
		if ok {
			select {
			case ch.(chan []byte) <- d.Body:
				p.delivery.Delete(d.CorrelationId)
			default:
				p.delivery.Delete(d.CorrelationId)
			}
		}
		return amqp.Ack
	}, replyQueueName,
		amqp.WithConsumerOptionsExchangeName(opt.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(opt.ExchangeType),
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(consumerID),
		amqp.WithConsumerOptionsConcurrency(concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(prefetch),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		return fmt.Errorf("failed to start RPC listener: %w", err)
	}

	return nil
}

func (p *rabbitmq) PublisherRPC(ctx context.Context, opt sdk_dto.Request[*sdk_dto.RabbitOptions]) ([]byte, error) {
	payload := opt.Payload

	if opt.Payload.ReplyTo == sdk_cons.EMPTY {
		return nil, errors.New("reply queue name cannot be empty")
	}

	if err := p.startRPCListener(payload.ReplyTo, payload); err != nil {
		return nil, fmt.Errorf("failed to start RPC listener: %w", err)
	}

	correlationID := payload.CorrelationID
	if correlationID == sdk_cons.EMPTY {
		correlationID = shortuuid.New()
	}

	res := make(chan []byte, 1)
	p.delivery.Store(correlationID, res)
	defer p.delivery.Delete(correlationID)

	bodyByte, err := p.parser.Marshal(&payload.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC body: %w", err)
	}

	err = p.publisher.PublishWithContext(ctx, bodyByte, []string{payload.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(payload.ExchangeName),
		amqp.WithPublishOptionsCorrelationID(correlationID),
		amqp.WithPublishOptionsReplyTo(p.replyQueueName),
		amqp.WithPublishOptionsAppID(payload.AppID),
		amqp.WithPublishOptionsUserID(payload.UserID),
		amqp.WithPublishOptionsContentType(payload.ContentType),
		amqp.WithPublishOptionsTimestamp(payload.Timestamp),
		amqp.WithPublishOptionsHeaders(payload.Args),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish RPC: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	select {
	case r := <-res:
		return r, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for RPC response: %w", ctx.Err())
	}
}

func (p *rabbitmq) ConsumerRPC(opt sdk_dto.Request[*sdk_dto.RabbitOptions], handler func(delivery amqp.Delivery) (action amqp.Action)) error {
	payload := opt.Payload

	consumerID := payload.ConsumerID
	if consumerID == sdk_cons.EMPTY {
		consumerID = shortuuid.New()
	}

	concurrency := payload.Concurrency
	if concurrency < 1 {
		concurrency = max(1, getCPUCount()/2)
	}

	prefetch := payload.Prefetch
	if prefetch < 1 {
		prefetch = 5
	}

	consumer, err := amqp.NewConsumer(p.rabbitmq, handler, payload.QueueName,
		amqp.WithConsumerOptionsExchangeName(payload.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(payload.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey:     payload.QueueName,
			BindingOptions: amqp.BindingOptions{Declare: true, NoWait: true, Args: payload.Args},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(consumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(payload.Ack),
		amqp.WithConsumerOptionsConcurrency(concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(prefetch),
		amqp.WithConsumerOptionsQueueArgs(payload.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		return fmt.Errorf("error creating RPC consumer: %w", err)
	}

	p.consumers.Store(consumerID, consumer)
	return nil
}

func (p *rabbitmq) ReplyToDeliveryPublisher(ctx context.Context, opt sdk_dto.Request[*sdk_dto.RabbitOptions]) error {
	payload := opt.Payload

	if payload.Delivery.Body == nil {
		return errors.New("delivery body cannot be nil")
	}

	bodyByte, err := p.parser.Marshal(&payload.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal reply body: %w", err)
	}

	return p.publisher.PublishWithContext(ctx, bodyByte, []string{payload.Delivery.ReplyTo},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsCorrelationID(payload.Delivery.CorrelationId),
		amqp.WithPublishOptionsAppID(payload.Delivery.AppId),
		amqp.WithPublishOptionsUserID(payload.Delivery.UserId),
		amqp.WithPublishOptionsContentType(payload.Delivery.ContentType),
		amqp.WithPublishOptionsTimestamp(payload.Delivery.Timestamp),
		amqp.WithPublishOptionsExpiration(payload.Expired),
		amqp.WithPublishOptionsHeaders(payload.Args),
	)
}
