package sdk_pkg

import (
	"errors"
	"testing"
	"time"

	amqp "github.com/wagslane/go-rabbitmq"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func TestNormalizeRabbitOptionsDoesNotMutateCaller(t *testing.T) {
	req := &sdk_dto.RabbitOptions{QueueName: "worker.test", ExchangeName: "ex.test"}

	opts := normalizeRabbitOptions(req)

	if opts.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", opts.ContentType)
	}

	if opts.AppID == "" {
		t.Fatal("app id must be generated")
	}

	if opts.Timestamp.IsZero() {
		t.Fatal("timestamp must be set")
	}

	if req.ContentType != "" || req.AppID != "" || !req.Timestamp.IsZero() {
		t.Fatal("caller struct must not be mutated")
	}
}

func TestQueueArgsAndRoutingKeyFallback(t *testing.T) {
	legacy := &sdk_dto.RabbitOptions{QueueName: "worker.queue", Args: amqp.Table{"x-queue-type": "quorum"}}

	if queueArgsOf(legacy)["x-queue-type"] != "quorum" {
		t.Fatal("legacy Args must still be used as queue args")
	}

	if routingKeyOf(legacy) != "worker.queue" {
		t.Fatal("routing key must fall back to the queue name")
	}

	override := &sdk_dto.RabbitOptions{
		QueueName:  "worker.queue",
		Args:       amqp.Table{"x-queue-type": "classic"},
		QueueArgs:  amqp.Table{"x-queue-type": "quorum"},
		BindingKey: "worker.binding",
	}

	if queueArgsOf(override)["x-queue-type"] != "quorum" {
		t.Fatal("QueueArgs must take precedence over Args")
	}

	if routingKeyOf(override) != "worker.binding" {
		t.Fatal("BindingKey must take precedence over QueueName")
	}
}

func TestNormalizeExpiration(t *testing.T) {
	if normalizeExpiration("60000") != "60000" {
		t.Fatal("numeric expiration must be kept")
	}

	if normalizeExpiration("10s") != "" {
		t.Fatal("non numeric expiration must be dropped, otherwise the channel is closed by the broker")
	}

	if normalizeExpiration("") != "" {
		t.Fatal("empty expiration must stay empty")
	}
}

func TestSafeUserID(t *testing.T) {
	client := &rabbitmqClient{connUser: "dgmi"}

	if got := client.safeUserID("dgmi"); got != "dgmi" {
		t.Fatalf("matching user id must be kept, got %q", got)
	}

	if got := client.safeUserID("other"); got != "" {
		t.Fatalf("mismatching user id must be dropped, got %q", got)
	}

	unknown := &rabbitmqClient{}

	if got := unknown.safeUserID("dgmi"); got != "" {
		t.Fatalf("user id must be dropped when the connection user is unknown, got %q", got)
	}
}

func TestPublisherPoolKeySeparatesExchangeArgs(t *testing.T) {
	plain := publisherPoolKey("ex.worker", "topic", nil)
	withArgs := publisherPoolKey("ex.worker", "topic", amqp.Table{"alternate-exchange": "ex.dlx"})

	if plain == withArgs {
		t.Fatal("different exchange arguments must not share a publisher")
	}

	first := publisherPoolKey("ex.worker", "topic", amqp.Table{"a": "1", "b": "2"})
	second := publisherPoolKey("ex.worker", "topic", amqp.Table{"b": "2", "a": "1"})

	if first != second {
		t.Fatal("argument order must not change the pool key")
	}
}

func TestBuildConsumerSpecDefaults(t *testing.T) {
	spec, err := buildConsumerSpec(&sdk_dto.RabbitOptions{
		QueueName:    "worker.test",
		ExchangeName: "ex.test",
		Args:         amqp.Table{"x-queue-type": "quorum"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.prefetch != defaultPrefetch {
		t.Fatalf("unexpected prefetch: %d", spec.prefetch)
	}

	if spec.concurrency < 1 {
		t.Fatalf("concurrency must be at least 1, got %d", spec.concurrency)
	}

	if spec.consumerName == "" {
		t.Fatal("consumer name must be generated")
	}

	if spec.bindingKey != "worker.test" {
		t.Fatalf("unexpected binding key: %s", spec.bindingKey)
	}

	if !spec.queueDurable {
		t.Fatal("queues must stay durable")
	}

	if _, err := buildConsumerSpec(&sdk_dto.RabbitOptions{}); err == nil {
		t.Fatal("missing queue name must be rejected")
	}
}

func TestPublishTimeoutBounds(t *testing.T) {
	if got := publishTimeoutOrDefault(0); got != defaultPublishTimeout {
		t.Fatalf("unexpected default publish timeout: %s", got)
	}

	if got := publishTimeoutOrDefault(45 * time.Second); got != maxPublishTimeout {
		t.Fatalf("publish timeout must be capped, got %s", got)
	}

	if got := publishTimeoutOrDefault(3 * time.Second); got != 3*time.Second {
		t.Fatalf("unexpected publish timeout: %s", got)
	}
}

func TestIsExchangeNotFound(t *testing.T) {
	if !isExchangeNotFound(errors.New(`Exception (404) Reason: "NOT_FOUND - no exchange 'ex.worker' in vhost '/'"`)) {
		t.Fatal("404 no exchange must be detected")
	}

	if isExchangeNotFound(errors.New(`Exception (406) Reason: "PRECONDITION_FAILED - inequivalent arg 'x-queue-type'"`)) {
		t.Fatal("precondition failures must not trigger an active declaration")
	}

	if isExchangeNotFound(nil) {
		t.Fatal("nil error must not be treated as missing exchange")
	}
}
