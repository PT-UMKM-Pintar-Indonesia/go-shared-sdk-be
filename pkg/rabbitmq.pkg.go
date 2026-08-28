package sdk_pkg

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	sdk_con "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/connections"
	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/lithammer/shortuuid"
	amqp "github.com/wagslane/go-rabbitmq"
)

const (
	defaultRPCTimeout       = 60 * time.Second
	maxRPCTimeout           = 5 * time.Minute
	replyQueueCleanupTicker = 1 * time.Minute
	publisherHealthInterval = 30 * time.Second
	publisherIdleTimeout    = 10 * time.Minute
)

type (
	rpcRequest struct {
		ch        chan []byte
		createdAt time.Time
		closed    atomic.Bool
	}

	publisherEntry struct {
		publisher   *amqp.Publisher
		exchange    string
		exchangeKey string
		createdAt   time.Time
		lastUsed    atomic.Int64
		healthy     atomic.Bool
	}

	rabbitmqClient struct {
		ctx            context.Context
		rabbitmq       *amqp.Conn
		replyQueueName string
		instanceID     string

		rpcConsumer       *amqp.Consumer
		startConsumerOnce sync.Once
		consumerErr       error
		consumerReady     atomic.Bool

		publisherPool  map[string]*publisherEntry
		publisherMutex sync.RWMutex

		requests     sync.Map
		cleanupDone  chan struct{}
		shutdownOnce sync.Once
	}
)

func NewRabbitMQ(opt *sdk_dto.RabbitClientOptions) (sdk_inf.IRabbitMQ, *amqp.Conn, error) {
	instanceID := shortuuid.New()

	var con *amqp.Conn
	var err error

	if opt.Cluster {
		con, err = sdk_con.RabbitConnectionCluster(opt)
	} else {
		con, err = sdk_con.RabbitConnection(opt)
	}

	if err != nil {
		return nil, nil, err
	}

	client := &rabbitmqClient{
		ctx:            opt.Ctx,
		rabbitmq:       con,
		instanceID:     instanceID,
		replyQueueName: shortuuid.New(),
		publisherPool:  make(map[string]*publisherEntry),
		cleanupDone:    make(chan struct{}),
	}

	go client.backgroundTasks()

	go func() {
		<-opt.Ctx.Done()
		client.Close()
	}()

	return client, con, nil
}

func (h *rabbitmqClient) Close() error {
	h.shutdown()
	return nil
}

func (h *rabbitmqClient) backgroundTasks() {
	cleanupTicker := time.NewTicker(replyQueueCleanupTicker)
	healthTicker := time.NewTicker(publisherHealthInterval)

	defer cleanupTicker.Stop()
	defer healthTicker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.cleanupDone:
			return
		case <-cleanupTicker.C:
			h.cleanupExpiredRequests()
		case <-healthTicker.C:
			h.checkPublishersHealth()
		}
	}
}

func (h *rabbitmqClient) cleanupExpiredRequests() {
	now := time.Now()

	h.requests.Range(func(key, value any) bool {
		req := value.(*rpcRequest)

		if now.Sub(req.createdAt) > maxRPCTimeout {
			h.requests.Delete(key)

			if req.closed.CompareAndSwap(false, true) {
				close(req.ch)
			}
		}

		return true
	})
}

func (h *rabbitmqClient) checkPublishersHealth() {
	h.publisherMutex.Lock()
	defer h.publisherMutex.Unlock()

	for key, entry := range h.publisherPool {
		lastUsed := time.Unix(0, entry.lastUsed.Load())

		if time.Since(lastUsed) > publisherIdleTimeout {
			entry.publisher.Close()
			delete(h.publisherPool, key)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := entry.publisher.PublishWithContext(ctx,
			[]byte(`{"_health":"check"}`),
			[]string{"__health_check__"},
			amqp.WithPublishOptionsExchange(entry.exchange),
		)
		cancel()

		if err != nil {
			entry.publisher.Close()
			delete(h.publisherPool, key)
		}
	}
}

func (h *rabbitmqClient) shutdown() {
	h.shutdownOnce.Do(func() {
		close(h.cleanupDone)

		h.publisherMutex.Lock()
		defer h.publisherMutex.Unlock()

		for key, entry := range h.publisherPool {
			entry.publisher.Close()
			delete(h.publisherPool, key)
		}

		if h.rpcConsumer != nil {
			h.rpcConsumer.Close()
		}

		h.requests.Range(func(key, value any) bool {
			req := value.(*rpcRequest)

			if req.closed.CompareAndSwap(false, true) {
				close(req.ch)
			}

			h.requests.Delete(key)
			return true
		})
	})
}

func (h *rabbitmqClient) getOrCreatePublisher(exchangeName, exchangeType string, args amqp.Table) (*amqp.Publisher, error) {
	poolKey := fmt.Sprintf("%s:%s", exchangeName, exchangeType)

	h.publisherMutex.RLock()
	entry, exists := h.publisherPool[poolKey]
	h.publisherMutex.RUnlock()

	if exists && entry.healthy.Load() {
		entry.lastUsed.Store(time.Now().UnixNano())
		return entry.publisher, nil
	}

	h.publisherMutex.Lock()
	defer h.publisherMutex.Unlock()

	if entry, exists = h.publisherPool[poolKey]; exists && entry.healthy.Load() {
		entry.lastUsed.Store(time.Now().UnixNano())
		return entry.publisher, nil
	}

	if exists {
		entry.publisher.Close()
		delete(h.publisherPool, poolKey)
	}

	publisher, err := amqp.NewPublisher(h.rabbitmq,
		amqp.WithPublisherOptionsExchangeName(exchangeName),
		amqp.WithPublisherOptionsExchangeKind(exchangeType),
		amqp.WithPublisherOptionsExchangeDeclare,
		amqp.WithPublisherOptionsExchangeDurable,
		amqp.WithPublisherOptionsExchangeNoWait,
		amqp.WithPublisherOptionsExchangeArgs(args),
		amqp.WithPublisherOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher for %s: %w", poolKey, err)
	}

	newEntry := &publisherEntry{
		publisher:   publisher,
		exchange:    exchangeName,
		exchangeKey: poolKey,
		createdAt:   time.Now(),
	}

	newEntry.healthy.Store(true)
	newEntry.lastUsed.Store(time.Now().UnixNano())

	h.publisherPool[poolKey] = newEntry
	return publisher, nil
}

func (h *rabbitmqClient) markPublisherUnhealthy(exchangeName, exchangeType string) {
	poolKey := fmt.Sprintf("%s:%s", exchangeName, exchangeType)

	h.publisherMutex.RLock()
	entry, exists := h.publisherPool[poolKey]
	h.publisherMutex.RUnlock()

	if exists {
		entry.healthy.Store(false)
	}
}

func (h *rabbitmqClient) ensureRPCReplyConsumer() error {
	h.startConsumerOnce.Do(func() {
		consumer, err := amqp.NewConsumer(h.rabbitmq, h.replyQueueName,
			amqp.WithConsumerOptionsConsumerExclusive,
			amqp.WithConsumerOptionsQueueAutoDelete,
			amqp.WithConsumerOptionsConsumerName(fmt.Sprintf("rpc-reply-%s", h.instanceID)),
			amqp.WithConsumerOptionsConcurrency(100),
			amqp.WithConsumerOptionsQOSPrefetch(100),
			amqp.WithConsumerOptionsLogging,
		)

		if err != nil {
			h.consumerErr = fmt.Errorf("failed to create RPC reply consumer: %w", err)
			return
		}

		err = consumer.Run(func(d amqp.Delivery) amqp.Action {
			if d.CorrelationId == sdk_cons.EMPTY {
				return amqp.NackDiscard
			}

			if val, ok := h.requests.Load(d.CorrelationId); ok {
				req := val.(*rpcRequest)

				if req.closed.Load() {
					return amqp.Ack
				}

				select {
				case req.ch <- d.Body:
					h.requests.Delete(d.CorrelationId)
					return amqp.Ack
				case <-time.After(1 * time.Second):
					return amqp.Ack
				}
			}

			return amqp.Ack
		})

		if err != nil {
			h.consumerErr = fmt.Errorf("failed to create RPC reply consumer: %w", err)
			return
		}

		h.rpcConsumer = consumer
		h.consumerReady.Store(true)
	})

	return h.consumerErr
}

func (h *rabbitmqClient) Publisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error {
	if req.Payload.ContentType == sdk_cons.EMPTY {
		req.Payload.ContentType = "application/json"
	}

	if req.Payload.AppID == sdk_cons.EMPTY {
		req.Payload.AppID = shortuuid.New()
	}

	if time.Until(req.Payload.Timestamp) < 1 {
		req.Payload.Timestamp = time.Now().Local()
	}

	publisher, err := h.getOrCreatePublisher(req.Payload.ExchangeName, req.Payload.ExchangeType, req.Payload.Args)
	if err != nil {
		return err
	}

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Payload.Body)
	if err != nil {
		return err
	}

	err = publisher.PublishWithContext(h.ctx, bodyByte, []string{req.Payload.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Payload.ExchangeName),
		amqp.WithPublishOptionsContentType(req.Payload.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Payload.Timestamp),
		amqp.WithPublishOptionsHeaders(req.Payload.Args),
	)

	if err != nil {
		h.markPublisherUnhealthy(req.Payload.ExchangeName, req.Payload.ExchangeType)
	}

	return err
}

func (h *rabbitmqClient) Consumer(req sdk_dto.Request[sdk_dto.RabbitOptions], callback func(d amqp.Delivery) (action amqp.Action)) error {
	if req.Payload.ConsumerID == sdk_cons.EMPTY {
		req.Payload.ConsumerID = shortuuid.New()
	}

	if req.Payload.Concurrency < 1 {
		req.Payload.Concurrency = int(float64(runtime.NumCPU()) * 0.75)

		if req.Payload.Concurrency < 1 {
			req.Payload.Concurrency = 1
		}
	}

	if req.Payload.Prefetch < 1 {
		req.Payload.Prefetch = 5
	}

	consumer, err := amqp.NewConsumer(h.rabbitmq, req.Payload.QueueName,
		amqp.WithConsumerOptionsExchangeName(req.Payload.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.Payload.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey: req.Payload.QueueName,
			BindingOptions: amqp.BindingOptions{
				Declare: true,
				NoWait:  true,
				Args:    req.Payload.Args,
			},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsExchangeDeclare,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(req.Payload.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Payload.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Payload.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Payload.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Payload.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		return err
	}

	return consumer.Run(callback)
}

func (h *rabbitmqClient) PublisherRPC(req sdk_dto.Request[sdk_dto.RabbitOptions]) ([]byte, error) {
	if err := h.ensureRPCReplyConsumer(); err != nil {
		return nil, fmt.Errorf("RPC consumer not ready: %w", err)
	}

	deadline := time.Now().Add(5 * time.Second)

	for !h.consumerReady.Load() {
		if time.Now().After(deadline) {
			return nil, errors.New("RPC consumer not ready after 5s")
		}

		time.Sleep(50 * time.Millisecond)
	}

	publisher, err := h.getOrCreatePublisher(req.Payload.ExchangeName, req.Payload.ExchangeType, req.Payload.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher: %w", err)
	}

	if req.Payload.ContentType == sdk_cons.EMPTY {
		req.Payload.ContentType = "application/json"
	}

	if req.Payload.AppID == sdk_cons.EMPTY {
		req.Payload.AppID = shortuuid.New()
	}

	if req.Payload.CorrelationID == sdk_cons.EMPTY {
		req.Payload.CorrelationID = shortuuid.New()
	}

	if time.Until(req.Payload.Timestamp) < 1 {
		req.Payload.Timestamp = time.Now().Local()
	}

	rpcReq := &rpcRequest{
		ch:        make(chan []byte, 1),
		createdAt: time.Now(),
	}

	h.requests.Store(req.Payload.CorrelationID, rpcReq)

	defer func() {
		h.requests.Delete(req.Payload.CorrelationID)
		select {
		case <-rpcReq.ch:
		default:
		}
	}()

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Payload.Body)
	if err != nil {
		return nil, err
	}

	err = publisher.PublishWithContext(h.ctx, bodyByte, []string{req.Payload.QueueName},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(req.Payload.ExchangeName),
		amqp.WithPublishOptionsCorrelationID(req.Payload.CorrelationID),
		amqp.WithPublishOptionsReplyTo(h.replyQueueName),
		amqp.WithPublishOptionsAppID(req.Payload.AppID),
		amqp.WithPublishOptionsUserID(req.Payload.UserID),
		amqp.WithPublishOptionsContentType(req.Payload.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Payload.Timestamp),
		amqp.WithPublishOptionsHeaders(req.Payload.Args),
	)

	if err != nil {
		h.markPublisherUnhealthy(req.Payload.ExchangeName, req.Payload.ExchangeType)
		return nil, fmt.Errorf("failed to publish RPC request: %w", err)
	}

	timeoutDuration := req.Payload.Timeout
	if timeoutDuration <= 0 {
		timeoutDuration = defaultRPCTimeout
	}
	if timeoutDuration > maxRPCTimeout {
		timeoutDuration = maxRPCTimeout
	}

	ctx, cancel := context.WithTimeout(h.ctx, timeoutDuration)
	defer cancel()

	select {
	case r, ok := <-rpcReq.ch:
		if !ok {
			return nil, errors.New("RPC channel closed before response received")
		}

		return r, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("RPC timeout after %v for correlation_id: %s", timeoutDuration, req.Payload.CorrelationID)
		}

		return nil, fmt.Errorf("RPC cancelled: %w", ctx.Err())
	}
}

func (h *rabbitmqClient) ConsumerRPC(req sdk_dto.Request[sdk_dto.RabbitOptions], handler func(delivery amqp.Delivery) (action amqp.Action)) error {
	if req.Payload.ConsumerID == sdk_cons.EMPTY {
		req.Payload.ConsumerID = shortuuid.New()
	}

	if req.Payload.Concurrency < 1 {
		req.Payload.Concurrency = int(float64(runtime.NumCPU()) * 0.75)

		if req.Payload.Concurrency < 1 {
			req.Payload.Concurrency = 1
		}
	}

	if req.Payload.Prefetch < 1 {
		req.Payload.Prefetch = 5
	}

	consumer, err := amqp.NewConsumer(h.rabbitmq, req.Payload.QueueName,
		amqp.WithConsumerOptionsExchangeName(req.Payload.ExchangeName),
		amqp.WithConsumerOptionsExchangeKind(req.Payload.ExchangeType),
		amqp.WithConsumerOptionsBinding(amqp.Binding{
			RoutingKey: req.Payload.QueueName,
			BindingOptions: amqp.BindingOptions{
				Declare: true,
				NoWait:  true,
				Args:    req.Payload.Args,
			},
		}),
		amqp.WithConsumerOptionsExchangeDurable,
		amqp.WithConsumerOptionsQueueDurable,
		amqp.WithConsumerOptionsConsumerName(req.Payload.ConsumerID),
		amqp.WithConsumerOptionsConsumerAutoAck(req.Payload.Ack),
		amqp.WithConsumerOptionsConcurrency(req.Payload.Concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(req.Payload.Prefetch),
		amqp.WithConsumerOptionsQueueArgs(req.Payload.Args),
		amqp.WithConsumerOptionsLogging,
	)

	if err != nil {
		return err
	}

	return consumer.Run(handler)
}

func (h *rabbitmqClient) ReplyToDeliveryPublisher(req sdk_dto.Request[sdk_dto.RabbitOptions]) error {
	if req.Payload.Delivery.Body == nil {
		return errors.New("delivery body cannot be nil")
	}

	if req.Payload.Delivery.ReplyTo == sdk_cons.EMPTY || req.Payload.Delivery.CorrelationId == sdk_cons.EMPTY {
		return errors.New("invalid delivery: ReplyTo or CorrelationId is empty")
	}

	publisher, err := h.getOrCreatePublisher(req.Payload.ExchangeName, req.Payload.ExchangeType, req.Payload.Args)
	if err != nil {
		return err
	}

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Payload.Body)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = publisher.PublishWithContext(ctx, bodyByte, []string{req.Payload.Delivery.ReplyTo},
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsCorrelationID(req.Payload.Delivery.CorrelationId),
		amqp.WithPublishOptionsAppID(req.Payload.Delivery.AppId),
		amqp.WithPublishOptionsUserID(req.Payload.Delivery.UserId),
		amqp.WithPublishOptionsContentType(req.Payload.Delivery.ContentType),
		amqp.WithPublishOptionsTimestamp(req.Payload.Delivery.Timestamp),
		amqp.WithPublishOptionsExpiration(req.Payload.Expired),
		amqp.WithPublishOptionsHeaders(req.Payload.Args),
	)

	if err != nil {
		h.markPublisherUnhealthy(req.Payload.ExchangeName, req.Payload.ExchangeType)
	}

	return err
}
