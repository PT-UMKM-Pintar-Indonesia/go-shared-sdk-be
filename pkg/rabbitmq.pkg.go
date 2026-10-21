package sdk_pkg

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/sirupsen/logrus"
	amqp "github.com/wagslane/go-rabbitmq"

	sdk_con "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/connections"
	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

const (
	defaultRPCTimeout = 60 * time.Second
	minRPCTimeout     = 500 * time.Millisecond
	maxRPCTimeout     = 5 * time.Minute

	defaultPublishTimeout = 10 * time.Second
	minPublishTimeout     = 1 * time.Second
	maxPublishTimeout     = 30 * time.Second

	defaultPrefetch    = 10
	defaultConcurrency = 4

	defaultRPCReplyPrefetch    = 100
	defaultRPCReplyConcurrency = 20

	replyQueueCleanupTicker = 30 * time.Second
	replyProbeTries         = 3
	replyProbeTimeout       = 3 * time.Second
	replyProbeRetryDelay    = 250 * time.Millisecond

	consumerRetryMinDelay = 500 * time.Millisecond
	consumerRetryMaxDelay = 30 * time.Second
	consumerCloseTimeout  = 5 * time.Second

	publisherIdleTimeout = 30 * time.Minute
	maxPublishFailures   = 5

	requestCleanupGrace = 30 * time.Second
)

var errClientClosed = errors.New("rabbitmq client is closed")

type rabbitLogger struct{}

func (rabbitLogger) Fatalf(format string, v ...any) {
	logrus.Fatalf("LOG_RABBITMQ: "+format, v...)
}

func (rabbitLogger) Errorf(format string, v ...any) {
	logrus.Errorf("LOG_RABBITMQ: "+format, v...)
}

func (rabbitLogger) Infof(format string, v ...any) {
	logrus.Debugf("LOG_RABBITMQ: "+format, v...)
}

func (rabbitLogger) Warnf(format string, v ...any) {
	logrus.Warnf("LOG_RABBITMQ: "+format, v...)
}

func (rabbitLogger) Debugf(format string, v ...any) {
	logrus.Debugf("LOG_RABBITMQ: "+format, v...)
}

type (
	rpcRequest struct {
		ch        chan []byte
		createdAt time.Time
		deadline  time.Time
		settled   atomic.Bool
	}

	publisherEntry struct {
		publisher *amqp.Publisher
		exchange  string
		kind      string
		createdAt time.Time
		lastUsed  atomic.Int64
		failures  atomic.Int64
		retired   atomic.Bool
		refs      atomic.Int64
		closeOnce sync.Once
	}

	consumerHandle struct {
		consumer *amqp.Consumer
		once     sync.Once
	}

	consumerSpec struct {
		queueName       string
		exchangeName    string
		exchangeType    string
		bindingKey      string
		consumerName    string
		queueArgs       amqp.Table
		exchangeArgs    amqp.Table
		concurrency     int
		prefetch        int
		autoAck         bool
		queueDurable    bool
		queueExclusive  bool
		queueAutoDelete bool
	}

	rabbitmqClient struct {
		ctx      context.Context
		cancel   context.CancelFunc
		rabbitmq *amqp.Conn
		connUser string

		replyQueueName   string
		instanceID       string
		replyPrefetch    int
		replyConcurrency int
		publishTimeout   time.Duration

		replyMutex    sync.Mutex
		replyConsumer *consumerHandle
		replyReady    bool
		replyGen      int

		publisherPool  map[string]*publisherEntry
		publisherLocks sync.Map
		publisherMutex sync.RWMutex

		requests sync.Map

		consumersMutex sync.Mutex
		consumers      map[int]*consumerHandle
		consumerSeq    int

		cleanupDone  chan struct{}
		closedCh     chan struct{}
		closed       atomic.Bool
		shutdownOnce sync.Once
	}
)

func NewRabbitMQ(ctx context.Context, opt *sdk_dto.RabbitClientOptions) (sdk_inf.IRabbitMQ, *amqp.Conn, error) {
	if opt == nil {
		return nil, nil, errors.New("rabbitmq options cannot be nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	con, err := sdk_con.RabbitMQConnection(opt)
	if err != nil {
		return nil, nil, err
	}

	clientCtx, cancel := context.WithCancel(ctx)
	instanceID := shortuuid.New()

	client := &rabbitmqClient{
		ctx:              clientCtx,
		cancel:           cancel,
		rabbitmq:         con,
		connUser:         usernameFromURL(opt.Url),
		instanceID:       instanceID,
		replyQueueName:   fmt.Sprintf("rpc.reply.%s", instanceID),
		replyPrefetch:    opt.RPCReplyPrefetch,
		replyConcurrency: opt.RPCReplyConcurrency,
		publishTimeout:   opt.PublishTimeout,
		publisherPool:    make(map[string]*publisherEntry),
		consumers:        make(map[int]*consumerHandle),
		cleanupDone:      make(chan struct{}),
		closedCh:         make(chan struct{}),
	}

	if client.replyPrefetch < 1 {
		client.replyPrefetch = defaultRPCReplyPrefetch
	}

	if client.replyConcurrency < 1 {
		client.replyConcurrency = defaultRPCReplyConcurrency
	}

	client.publishTimeout = publishTimeoutOrDefault(opt.PublishTimeout)

	go client.backgroundTasks()

	go func() {
		select {
		case <-ctx.Done():
			client.Close()
		case <-client.closedCh:
		}
	}()

	return client, con, nil
}

func publishTimeoutOrDefault(value time.Duration) time.Duration {
	if value < minPublishTimeout {
		return defaultPublishTimeout
	}

	if value > maxPublishTimeout {
		return maxPublishTimeout
	}

	return value
}

func usernameFromURL(raw string) string {
	if raw == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.User == nil {
		return sdk_cons.EMPTY
	}

	return parsed.User.Username()
}

func (h *rabbitmqClient) Close() error {
	h.shutdown()
	return nil
}

func (h *rabbitmqClient) shutdown() {
	h.shutdownOnce.Do(func() {
		h.closed.Store(true)
		close(h.closedCh)
		close(h.cleanupDone)

		if h.cancel != nil {
			h.cancel()
		}

		h.replyMutex.Lock()
		replyConsumer := h.replyConsumer
		h.replyConsumer = nil
		h.replyReady = false
		h.replyGen++
		h.replyMutex.Unlock()

		if replyConsumer != nil {
			h.closeConsumer(replyConsumer)
		}

		h.consumersMutex.Lock()
		consumers := make([]*consumerHandle, 0, len(h.consumers))
		for id, consumer := range h.consumers {
			consumers = append(consumers, consumer)
			delete(h.consumers, id)
		}
		h.consumersMutex.Unlock()

		for _, consumer := range consumers {
			h.closeConsumer(consumer)
		}

		h.publisherMutex.Lock()
		entries := make([]*publisherEntry, 0, len(h.publisherPool))
		for key, entry := range h.publisherPool {
			entries = append(entries, entry)
			delete(h.publisherPool, key)
		}
		h.publisherMutex.Unlock()

		for _, entry := range entries {
			entry.retire()
		}

		logrus.Info("LOG_RABBITMQ_CLIENT_CLOSED")
	})
}

func newConsumerHandle(consumer *amqp.Consumer) *consumerHandle {
	return &consumerHandle{consumer: consumer}
}

func (c *consumerHandle) Close() {
	if c == nil {
		return
	}

	c.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), consumerCloseTimeout)
		defer cancel()

		c.consumer.CloseWithContext(ctx)
	})
}

func (h *rabbitmqClient) closeConsumer(consumer *consumerHandle) {
	consumer.Close()
}

func (h *rabbitmqClient) backgroundTasks() {
	ticker := time.NewTicker(replyQueueCleanupTicker)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.closedCh:
			return
		case <-h.cleanupDone:
			return
		case <-ticker.C:
			h.cleanupExpiredRequests()
			h.cleanupIdlePublishers()
		}
	}
}

func (h *rabbitmqClient) cleanupExpiredRequests() {
	now := time.Now()

	h.requests.Range(func(key, value any) bool {
		req, ok := value.(*rpcRequest)
		if !ok {
			h.requests.Delete(key)
			return true
		}

		if now.After(req.deadline.Add(requestCleanupGrace)) {
			req.settled.Store(true)
			h.requests.Delete(key)
		}

		return true
	})
}

func (h *rabbitmqClient) cleanupIdlePublishers() {
	now := time.Now()

	h.publisherMutex.Lock()

	stale := make([]*publisherEntry, 0)

	for key, entry := range h.publisherPool {
		lastUsed := time.Unix(0, entry.lastUsed.Load())

		if now.Sub(lastUsed) < publisherIdleTimeout {
			continue
		}

		stale = append(stale, entry)
		delete(h.publisherPool, key)
	}

	h.publisherMutex.Unlock()

	for _, entry := range stale {
		entry.retire()
	}
}

func (e *publisherEntry) acquire() bool {
	if e.retired.Load() {
		return false
	}

	e.refs.Add(1)

	if e.retired.Load() {
		e.release()
		return false
	}

	return true
}

func (e *publisherEntry) release() {
	if e.refs.Add(-1) == 0 && e.retired.Load() {
		e.close()
	}
}

func (e *publisherEntry) retire() {
	e.retired.Store(true)

	if e.refs.Load() == 0 {
		e.close()
	}
}

func (e *publisherEntry) close() {
	e.closeOnce.Do(func() {
		if e.publisher != nil {
			e.publisher.Close()
		}
	})
}

func (e *publisherEntry) markFailure() {
	e.failures.Add(1)
}

func (e *publisherEntry) markSuccess() {
	e.failures.Store(0)
}

func (h *rabbitmqClient) publisherKeyMutex(key string) *sync.Mutex {
	actual, _ := h.publisherLocks.LoadOrStore(key, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func publisherPoolKey(exchangeName, exchangeType string, args amqp.Table) string {
	if exchangeType == sdk_cons.EMPTY {
		exchangeType = sdk_cons.EXCHANGE_TYPE_DIRECT
	}

	return fmt.Sprintf("%s|%s|%s", exchangeName, exchangeType, argsFingerprint(args))
}

func argsFingerprint(args amqp.Table) string {
	if len(args) == 0 {
		return sdk_cons.EMPTY
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	builder := strings.Builder{}

	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(fmt.Sprintf("%v", args[key]))
		builder.WriteString(";")
	}

	return builder.String()
}

func (h *rabbitmqClient) acquirePublisher(key string) *publisherEntry {
	h.publisherMutex.RLock()
	entry := h.publisherPool[key]
	h.publisherMutex.RUnlock()

	if entry == nil {
		return nil
	}

	if !entry.acquire() {
		return nil
	}

	entry.lastUsed.Store(time.Now().UnixNano())
	return entry
}

func (h *rabbitmqClient) getOrCreatePublisher(exchangeName, exchangeType string, args amqp.Table) (*publisherEntry, error) {
	if h.closed.Load() {
		return nil, errClientClosed
	}

	if exchangeType == sdk_cons.EMPTY {
		exchangeType = sdk_cons.EXCHANGE_TYPE_DIRECT
	}

	key := publisherPoolKey(exchangeName, exchangeType, args)

	if entry := h.acquirePublisher(key); entry != nil {
		return entry, nil
	}

	mutex := h.publisherKeyMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	if entry := h.acquirePublisher(key); entry != nil {
		return entry, nil
	}

	h.publisherMutex.Lock()
	stale := h.publisherPool[key]
	delete(h.publisherPool, key)
	h.publisherMutex.Unlock()

	if stale != nil {
		stale.retire()
	}

	entry, err := h.createPublisher(exchangeName, exchangeType, args)
	if err != nil {
		return nil, err
	}

	h.publisherMutex.Lock()
	h.publisherPool[key] = entry
	h.publisherMutex.Unlock()

	if !entry.acquire() {
		return nil, errors.New("rabbitmq publisher was retired during creation")
	}

	return entry, nil
}

func (h *rabbitmqClient) createPublisher(exchangeName, exchangeType string, args amqp.Table) (*publisherEntry, error) {
	base := []func(*amqp.PublisherOptions){
		amqp.WithPublisherOptionsExchangeName(exchangeName),
		amqp.WithPublisherOptionsExchangeKind(exchangeType),
		amqp.WithPublisherOptionsLogger(rabbitLogger{}),
	}

	var (
		publisher *amqp.Publisher
		err       error
	)

	if exchangeName == sdk_cons.EMPTY {
		publisher, err = amqp.NewPublisher(h.rabbitmq, base...)
	} else {
		passive := make([]func(*amqp.PublisherOptions), 0, len(base)+3)
		passive = append(passive, base...)
		passive = append(passive, amqp.WithPublisherOptionsExchangeDeclare, amqp.WithPublisherOptionsExchangePassive)

		publisher, err = amqp.NewPublisher(h.rabbitmq, passive...)

		if err != nil {
			logrus.WithError(err).WithField("exchange", exchangeName).Warn("LOG_RABBITMQ_EXCHANGE_NOT_FOUND_TRYING_DECLARE")

			active := make([]func(*amqp.PublisherOptions), 0, len(base)+2)
			active = append(active, base...)
			active = append(active, amqp.WithPublisherOptionsExchangeDurable, amqp.WithPublisherOptionsExchangeDeclare)

			if len(args) > 0 {
				active = append(active, amqp.WithPublisherOptionsExchangeArgs(args))
			}

			publisher, err = amqp.NewPublisher(h.rabbitmq, active...)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create publisher for exchange %q: %w", exchangeName, err)
	}

	entry := &publisherEntry{
		publisher: publisher,
		exchange:  exchangeName,
		kind:      exchangeType,
		createdAt: time.Now(),
	}

	entry.lastUsed.Store(time.Now().UnixNano())

	return entry, nil
}

func (h *rabbitmqClient) releasePublisher(entry *publisherEntry, err error) {
	if entry == nil {
		return
	}

	if err == nil {
		entry.markSuccess()
		entry.release()
		return
	}

	entry.markFailure()

	if entry.failures.Load() >= maxPublishFailures {
		h.publisherMutex.Lock()

		for key, poolEntry := range h.publisherPool {
			if poolEntry == entry {
				delete(h.publisherPool, key)
			}
		}

		h.publisherMutex.Unlock()

		if entry.retired.CompareAndSwap(false, true) {
			logrus.WithFields(logrus.Fields{
				"exchange": entry.exchange,
				"kind":     entry.kind,
			}).Warn("LOG_RABBITMQ_PUBLISHER_RETIRED_AFTER_REPEATED_FAILURES")
		}
	}

	entry.release()
}

func normalizeRabbitOptions(req *sdk_dto.RabbitOptions) sdk_dto.RabbitOptions {
	opts := *req

	if opts.ContentType == sdk_cons.EMPTY {
		opts.ContentType = "application/json"
	}

	if opts.AppID == sdk_cons.EMPTY {
		opts.AppID = shortuuid.New()
	}

	if opts.Timestamp.IsZero() || time.Until(opts.Timestamp) < time.Second {
		opts.Timestamp = time.Now()
	}

	if opts.ExchangeType == sdk_cons.EMPTY {
		opts.ExchangeType = sdk_cons.EXCHANGE_TYPE_DIRECT
	}

	return opts
}

func queueArgsOf(opts *sdk_dto.RabbitOptions) amqp.Table {
	if len(opts.QueueArgs) > 0 {
		return opts.QueueArgs
	}

	return opts.Args
}

func routingKeyOf(opts *sdk_dto.RabbitOptions) string {
	if opts.BindingKey != sdk_cons.EMPTY {
		return opts.BindingKey
	}

	return opts.QueueName
}

func normalizeExpiration(value string) string {
	if value == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	if _, err := strconv.Atoi(value); err != nil {
		logrus.WithField("expiration", value).Warn("LOG_RABBITMQ_INVALID_EXPIRATION_IGNORED")
		return sdk_cons.EMPTY
	}

	return value
}

func (h *rabbitmqClient) safeUserID(userID string) string {
	if userID == sdk_cons.EMPTY || h.connUser == sdk_cons.EMPTY {
		return sdk_cons.EMPTY
	}

	if userID != h.connUser {
		logrus.WithFields(logrus.Fields{
			"connection_user": h.connUser,
			"requested_user":  userID,
		}).Debug("LOG_RABBITMQ_USER_ID_MISMATCH_IGNORED")

		return sdk_cons.EMPTY
	}

	return userID
}

func (h *rabbitmqClient) publishOptions(opts *sdk_dto.RabbitOptions) []func(*amqp.PublishOptions) {
	options := []func(*amqp.PublishOptions){
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsExchange(opts.ExchangeName),
		amqp.WithPublishOptionsContentType(opts.ContentType),
		amqp.WithPublishOptionsTimestamp(opts.Timestamp),
		amqp.WithPublishOptionsAppID(opts.AppID),
	}

	if opts.CorrelationID != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsCorrelationID(opts.CorrelationID))
	}

	if opts.ReplyTo != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsReplyTo(opts.ReplyTo))
	}

	if userID := h.safeUserID(opts.UserID); userID != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsUserID(userID))
	}

	if len(opts.Headers) > 0 {
		options = append(options, amqp.WithPublishOptionsHeaders(opts.Headers))
	}

	return options
}

func (h *rabbitmqClient) publishContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(h.ctx, h.publishTimeout)
}

func (h *rabbitmqClient) Publisher(req *sdk_dto.RabbitOptions) error {
	if req == nil {
		return errors.New("rabbit options cannot be nil")
	}

	opts := normalizeRabbitOptions(req)

	bodyByte, err := sdk_helper.NewParser().Marshal(&opts.Body)
	if err != nil {
		return err
	}

	entry, err := h.getOrCreatePublisher(opts.ExchangeName, opts.ExchangeType, opts.ExchangeArgs)
	if err != nil {
		return err
	}

	ctx, cancel := h.publishContext()
	defer cancel()

	err = entry.publisher.PublishWithContext(ctx, bodyByte, []string{routingKeyOf(&opts)}, h.publishOptions(&opts)...)

	h.releasePublisher(entry, err)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"exchange":    opts.ExchangeName,
			"routing_key": routingKeyOf(&opts),
		}).Error("LOG_RABBITMQ_PUBLISH_FAILED")
	}

	return err
}

func (h *rabbitmqClient) ReplyToDeliveryPublisher(req *sdk_dto.RabbitOptions) error {
	if req == nil {
		return errors.New("rabbit options cannot be nil")
	}

	if req.Delivery.ReplyTo == sdk_cons.EMPTY || req.Delivery.CorrelationId == sdk_cons.EMPTY {
		return errors.New("invalid delivery: ReplyTo or CorrelationId is empty")
	}

	bodyByte, err := sdk_helper.NewParser().Marshal(&req.Body)
	if err != nil {
		return err
	}

	entry, err := h.getOrCreatePublisher(sdk_cons.EMPTY, sdk_cons.EXCHANGE_TYPE_DIRECT, nil)
	if err != nil {
		return err
	}

	ctx, cancel := h.publishContext()
	defer cancel()

	options := []func(*amqp.PublishOptions){
		amqp.WithPublishOptionsPersistentDelivery,
		amqp.WithPublishOptionsCorrelationID(req.Delivery.CorrelationId),
	}

	if req.Delivery.AppId != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsAppID(req.Delivery.AppId))
	}

	if userID := h.safeUserID(req.Delivery.UserId); userID != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsUserID(userID))
	}

	if req.Delivery.ContentType != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsContentType(req.Delivery.ContentType))
	}

	options = append(options, amqp.WithPublishOptionsTimestamp(time.Now()))

	if len(req.Headers) > 0 {
		options = append(options, amqp.WithPublishOptionsHeaders(req.Headers))
	}

	if expiration := normalizeExpiration(req.Expired); expiration != sdk_cons.EMPTY {
		options = append(options, amqp.WithPublishOptionsExpiration(expiration))
	}

	err = entry.publisher.PublishWithContext(ctx, bodyByte, []string{req.Delivery.ReplyTo}, options...)

	h.releasePublisher(entry, err)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"reply_to":       req.Delivery.ReplyTo,
			"correlation_id": req.Delivery.CorrelationId,
		}).Error("LOG_RABBITMQ_REPLY_PUBLISH_FAILED")
	}

	return err
}

func (h *rabbitmqClient) ensureReplyConsumer() error {
	h.replyMutex.Lock()
	defer h.replyMutex.Unlock()

	if h.closed.Load() || h.ctx.Err() != nil {
		return errClientClosed
	}

	if h.replyReady && h.replyConsumer != nil {
		return nil
	}

	if h.replyConsumer != nil {
		consumer := h.replyConsumer
		h.replyConsumer = nil
		h.replyReady = false
		go h.closeConsumer(consumer)
	}

	consumer, err := amqp.NewConsumer(h.rabbitmq, h.replyQueueName,
		amqp.WithConsumerOptionsConsumerExclusive,
		amqp.WithConsumerOptionsConsumerName(fmt.Sprintf("rpc-reply-%s", h.instanceID)),
		amqp.WithConsumerOptionsConcurrency(h.replyConcurrency),
		amqp.WithConsumerOptionsQOSPrefetch(h.replyPrefetch),
		amqp.WithConsumerOptionsQueueExclusive,
		amqp.WithConsumerOptionsQueueAutoDelete,
		amqp.WithConsumerOptionsLogger(rabbitLogger{}),
	)
	if err != nil {
		return fmt.Errorf("failed to create RPC reply consumer: %w", err)
	}

	h.replyGen++
	gen := h.replyGen

	handle := newConsumerHandle(consumer)

	exitCh := make(chan error, 1)

	go func() {
		err := consumer.Run(h.handleReply)
		exitCh <- err

		h.replyMutex.Lock()
		if h.replyGen == gen {
			h.replyReady = false
			h.replyConsumer = nil
		}
		h.replyMutex.Unlock()

		if err != nil {
			logrus.WithError(err).Warn("LOG_RABBITMQ_RPC_REPLY_CONSUMER_STOPPED")
		}

		h.closeConsumer(handle)
	}()

	h.replyConsumer = handle

	if err := h.waitReplyConsumerReady(exitCh); err != nil {
		h.replyReady = false
		return err
	}

	h.replyReady = true

	logrus.WithFields(logrus.Fields{
		"queue":       h.replyQueueName,
		"prefetch":    h.replyPrefetch,
		"concurrency": h.replyConcurrency,
	}).Info("LOG_RABBITMQ_RPC_REPLY_CONSUMER_READY")

	return nil
}

func (h *rabbitmqClient) waitReplyConsumerReady(exitCh <-chan error) error {
	var lastErr error

	for attempt := 0; attempt < replyProbeTries; attempt++ {
		correlationID := fmt.Sprintf("probe.%s", shortuuid.New())

		probe := &rpcRequest{
			ch:        make(chan []byte, 1),
			createdAt: time.Now(),
			deadline:  time.Now().Add(replyProbeTimeout),
		}

		h.requests.Store(correlationID, probe)

		err := h.publishProbe(correlationID)

		if err != nil {
			h.requests.Delete(correlationID)
			probe.settled.Store(true)
			lastErr = err
		} else {
			select {
			case <-probe.ch:
				h.requests.Delete(correlationID)
				return nil

			case runErr := <-exitCh:
				h.requests.Delete(correlationID)
				probe.settled.Store(true)

				if runErr == nil {
					runErr = errors.New("consumer stopped before it became ready")
				}

				return fmt.Errorf("RPC reply consumer failed to start: %w", runErr)

			case <-time.After(replyProbeTimeout):
				h.requests.Delete(correlationID)
				probe.settled.Store(true)
				lastErr = fmt.Errorf("RPC reply probe timed out after %s", replyProbeTimeout)

			case <-h.closedCh:
				h.requests.Delete(correlationID)
				probe.settled.Store(true)
				return errClientClosed

			}
		}

		select {
		case <-time.After(replyProbeRetryDelay):
		case <-h.closedCh:
			return errClientClosed
		}
	}

	return fmt.Errorf("RPC reply consumer not ready: %w", lastErr)
}

func (h *rabbitmqClient) publishProbe(correlationID string) error {
	entry, err := h.getOrCreatePublisher(sdk_cons.EMPTY, sdk_cons.EXCHANGE_TYPE_DIRECT, nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(h.ctx, replyProbeTimeout)
	defer cancel()

	err = entry.publisher.PublishWithContext(ctx, []byte(`{"_probe":true}`), []string{h.replyQueueName},
		amqp.WithPublishOptionsCorrelationID(correlationID),
		amqp.WithPublishOptionsContentType("application/json"),
	)

	h.releasePublisher(entry, err)

	return err
}

func (h *rabbitmqClient) handleReply(d amqp.Delivery) amqp.Action {
	if d.CorrelationId == sdk_cons.EMPTY {
		return amqp.Ack
	}

	value, ok := h.requests.Load(d.CorrelationId)
	if !ok {
		return amqp.Ack
	}

	req, ok := value.(*rpcRequest)
	if !ok {
		h.requests.Delete(d.CorrelationId)
		return amqp.Ack
	}

	if !req.settled.CompareAndSwap(false, true) {
		return amqp.Ack
	}

	select {
	case req.ch <- d.Body:
	default:
	}

	return amqp.Ack
}

func (h *rabbitmqClient) PublisherRPC(req *sdk_dto.RabbitOptions) ([]byte, error) {
	if req == nil {
		return nil, errors.New("rabbit options cannot be nil")
	}

	opts := normalizeRabbitOptions(req)

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultRPCTimeout
	}

	if timeout < minRPCTimeout {
		timeout = minRPCTimeout
	}

	if timeout > maxRPCTimeout {
		timeout = maxRPCTimeout
	}

	if opts.CorrelationID == sdk_cons.EMPTY {
		opts.CorrelationID = shortuuid.New()
	}

	opts.ReplyTo = h.replyQueueName

	if h.closed.Load() {
		return nil, errClientClosed
	}

	// The reply queue must exist before the request is published, otherwise the
	// broker silently drops the reply (unroutable message).
	if err := h.ensureReplyConsumer(); err != nil {
		return nil, err
	}

	bodyByte, err := sdk_helper.NewParser().Marshal(&opts.Body)
	if err != nil {
		return nil, err
	}

	entry, err := h.getOrCreatePublisher(opts.ExchangeName, opts.ExchangeType, opts.ExchangeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher: %w", err)
	}

	rpcReq := &rpcRequest{
		ch:        make(chan []byte, 1),
		createdAt: time.Now(),
		deadline:  time.Now().Add(timeout),
	}

	h.requests.Store(opts.CorrelationID, rpcReq)

	defer func() {
		rpcReq.settled.Store(true)
		h.requests.Delete(opts.CorrelationID)
	}()

	publishCtx, publishCancel := h.publishContext()
	err = entry.publisher.PublishWithContext(publishCtx, bodyByte, []string{routingKeyOf(&opts)}, h.publishOptions(&opts)...)
	publishCancel()

	h.releasePublisher(entry, err)

	if err != nil {
		return nil, fmt.Errorf("failed to publish RPC request: %w", err)
	}

	ctx, cancel := context.WithTimeout(h.ctx, timeout)
	defer cancel()

	select {
	case response := <-rpcReq.ch:
		return response, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logrus.WithFields(logrus.Fields{
				"correlation_id": opts.CorrelationID,
				"exchange":       opts.ExchangeName,
				"routing_key":    routingKeyOf(&opts),
				"timeout":        timeout.String(),
			}).Warn("LOG_RABBITMQ_RPC_TIMEOUT")

			return nil, fmt.Errorf("RPC timeout after %v for correlation_id: %s", timeout, opts.CorrelationID)
		}

		return nil, fmt.Errorf("RPC cancelled: %w", ctx.Err())
	case <-h.closedCh:
		return nil, errClientClosed
	}
}

func buildConsumerSpec(req *sdk_dto.RabbitOptions) (consumerSpec, error) {
	opts := *req

	if opts.QueueName == sdk_cons.EMPTY {
		return consumerSpec{}, errors.New("rabbit queue name is required")
	}

	if opts.ConsumerID == sdk_cons.EMPTY {
		opts.ConsumerID = shortuuid.New()
	}

	if opts.Concurrency < 1 {
		opts.Concurrency = int(float64(runtime.NumCPU()) * 0.75)

		if opts.Concurrency < 1 {
			opts.Concurrency = defaultConcurrency
		}
	}

	if opts.Prefetch < 1 {
		opts.Prefetch = defaultPrefetch
	}

	if opts.ExchangeType == sdk_cons.EMPTY {
		opts.ExchangeType = sdk_cons.EXCHANGE_TYPE_DIRECT
	}

	return consumerSpec{
		queueName:      opts.QueueName,
		exchangeName:   opts.ExchangeName,
		exchangeType:   opts.ExchangeType,
		bindingKey:     routingKeyOf(&opts),
		consumerName:   opts.ConsumerID,
		queueArgs:      queueArgsOf(&opts),
		exchangeArgs:   opts.ExchangeArgs,
		concurrency:    opts.Concurrency,
		prefetch:       opts.Prefetch,
		autoAck:        opts.Ack,
		queueDurable:   true,
		queueExclusive: false,
	}, nil
}

func (h *rabbitmqClient) Consumer(req *sdk_dto.RabbitOptions, callback func(d amqp.Delivery) (action amqp.Action)) error {
	if req == nil {
		return errors.New("rabbit options cannot be nil")
	}

	if callback == nil {
		return errors.New("rabbit consumer callback cannot be nil")
	}

	spec, err := buildConsumerSpec(req)
	if err != nil {
		return err
	}

	return h.runConsumerLoop(spec, callback)
}

func (h *rabbitmqClient) ConsumerRPC(req *sdk_dto.RabbitOptions, handler func(delivery amqp.Delivery) (action amqp.Action)) error {
	return h.Consumer(req, handler)
}

func (h *rabbitmqClient) runConsumerLoop(spec consumerSpec, callback func(d amqp.Delivery) amqp.Action) error {
	handler := wrapConsumerHandler(spec, callback)

	backoff := consumerRetryMinDelay
	passiveExchange := true
	attempt := 0

	for {
		if h.closed.Load() || h.ctx.Err() != nil {
			return nil
		}

		attempt++

		consumer, err := h.newConsumer(spec, passiveExchange)

		if err == nil {
			handle := newConsumerHandle(consumer)
			id := h.registerConsumer(handle)
			startedAt := time.Now()

			runErr := consumer.Run(handler)

			h.unregisterConsumer(id)
			h.closeConsumer(handle)

			if h.closed.Load() || h.ctx.Err() != nil {
				return nil
			}

			if runErr == nil {
				runErr = errors.New("consumer stopped unexpectedly")
			}

			err = runErr

			if time.Since(startedAt) > consumerRetryMaxDelay {
				backoff = consumerRetryMinDelay
				attempt = 1
				passiveExchange = true
			}
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"queue":    spec.queueName,
			"exchange": spec.exchangeName,
			"attempt":  attempt,
			"backoff":  backoff.String(),
		}).Error("LOG_RABBITMQ_CONSUMER_START_FAILED_RETRYING")

		select {
		case <-time.After(backoff):
		case <-h.closedCh:
			return nil
		case <-h.ctx.Done():
			return nil
		}

		backoff *= 2
		if backoff > consumerRetryMaxDelay {
			backoff = consumerRetryMaxDelay
		}

		if passiveExchange && isExchangeNotFound(err) {
			passiveExchange = false
		}
	}
}

func isExchangeNotFound(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()

	return strings.Contains(message, "NOT_FOUND") && strings.Contains(message, "exchange")
}

func (h *rabbitmqClient) newConsumer(spec consumerSpec, passiveExchange bool) (*amqp.Consumer, error) {
	options := []func(*amqp.ConsumerOptions){
		amqp.WithConsumerOptionsConsumerName(spec.consumerName),
		amqp.WithConsumerOptionsConsumerAutoAck(spec.autoAck),
		amqp.WithConsumerOptionsConcurrency(spec.concurrency),
		amqp.WithConsumerOptionsQOSPrefetch(spec.prefetch),
		amqp.WithConsumerOptionsQueueArgs(spec.queueArgs),
		amqp.WithConsumerOptionsLogger(rabbitLogger{}),
	}

	if spec.queueDurable {
		options = append(options, amqp.WithConsumerOptionsQueueDurable)
	}

	if spec.queueExclusive {
		options = append(options, amqp.WithConsumerOptionsQueueExclusive)
	}

	if spec.queueAutoDelete {
		options = append(options, amqp.WithConsumerOptionsQueueAutoDelete)
	}

	if spec.exchangeName != sdk_cons.EMPTY {
		options = append(options,
			amqp.WithConsumerOptionsExchangeName(spec.exchangeName),
			amqp.WithConsumerOptionsExchangeKind(spec.exchangeType),
			amqp.WithConsumerOptionsExchangeDurable,
			amqp.WithConsumerOptionsExchangeDeclare,
			amqp.WithConsumerOptionsBinding(amqp.Binding{
				RoutingKey: spec.bindingKey,
				BindingOptions: amqp.BindingOptions{
					Declare: true,
				},
			}),
		)

		if passiveExchange {
			options = append(options, amqp.WithConsumerOptionsExchangePassive)
		}

		if len(spec.exchangeArgs) > 0 {
			options = append(options, amqp.WithConsumerOptionsExchangeArgs(spec.exchangeArgs))
		}
	}

	consumer, err := amqp.NewConsumer(h.rabbitmq, spec.queueName, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer for queue %q: %w", spec.queueName, err)
	}

	return consumer, nil
}

func wrapConsumerHandler(spec consumerSpec, callback func(d amqp.Delivery) amqp.Action) amqp.Handler {
	return func(d amqp.Delivery) (action amqp.Action) {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"queue":          spec.queueName,
					"exchange":       spec.exchangeName,
					"correlation_id": d.CorrelationId,
					"redelivered":    d.Redelivered,
					"panic":          fmt.Sprintf("%v", r),
				}).Error("LOG_RABBITMQ_HANDLER_PANIC_RECOVERED")

				// Requeue only once so a poison message cannot loop forever.
				if d.Redelivered {
					action = amqp.NackDiscard
				} else {
					action = amqp.NackRequeue
				}
			}
		}()

		return callback(d)
	}
}

func (h *rabbitmqClient) registerConsumer(consumer *consumerHandle) int {
	h.consumersMutex.Lock()
	defer h.consumersMutex.Unlock()

	h.consumerSeq++
	h.consumers[h.consumerSeq] = consumer

	return h.consumerSeq
}

func (h *rabbitmqClient) unregisterConsumer(id int) {
	h.consumersMutex.Lock()
	delete(h.consumers, id)
	h.consumersMutex.Unlock()
}
