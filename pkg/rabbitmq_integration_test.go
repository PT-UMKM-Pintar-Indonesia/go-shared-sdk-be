package sdk_pkg

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
	amqp "github.com/wagslane/go-rabbitmq"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

// TestRPCIntegrationWithHighPrefetch reproduces the reported production issue:
// RPC calls time out as soon as the prefetch is higher than 1. It runs against
// a real broker (set RABBITMQ_TEST_URL) and requires 500 concurrent RPC calls
// with prefetch=100 to pass without timeouts or mixed up correlation ids.
func TestRPCIntegrationWithHighPrefetch(t *testing.T) {
	url := os.Getenv("RABBITMQ_TEST_URL")
	if url == "" {
		t.Skip("RABBITMQ_TEST_URL is not set")
	}

	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	exchange := "ex.test.rpc." + suffix
	queue := "worker.test.rpc." + suffix

	server, _, err := NewRabbitMQ(ctx, &sdk_dto.RabbitClientOptions{
		Url:              url,
		PublishTimeout:   10 * time.Second,
		RPCReplyPrefetch: 100,
	})
	if err != nil {
		t.Fatalf("server client: %v", err)
	}
	defer server.Close()

	cleanup := func() {
		conn, err := amqp091.Dial(url)
		if err != nil {
			return
		}

		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			return
		}

		_, _ = ch.QueueDelete(queue, false, false, false)
		_ = ch.ExchangeDelete(exchange, false, false)
	}

	defer cleanup()

	serverOptions := &sdk_dto.RabbitOptions{
		ExchangeName: exchange,
		ExchangeType: "topic",
		QueueName:    queue,
		Prefetch:     100,
		Concurrency:  20,
		Timeout:      2 * time.Minute,
	}

	go func() {
		_ = server.ConsumerRPC(serverOptions, func(d amqp.Delivery) amqp.Action {
			// Per-delivery copy, exactly like the worker handler fix: a shared
			// struct here mixes up correlation ids once prefetch > 1.
			rreq := *serverOptions
			rreq.Delivery = d

			time.Sleep(5 * time.Millisecond)

			rreq.Body = string(d.Body)

			if err := server.ReplyToDeliveryPublisher(&rreq); err != nil {
				t.Logf("reply failed: %v", err)
			}

			return amqp.Ack
		})
	}()

	time.Sleep(2 * time.Second)

	const total = 500

	client, _, err := NewRabbitMQ(ctx, &sdk_dto.RabbitClientOptions{
		Url:                 url,
		RPCReplyPrefetch:    100,
		RPCReplyConcurrency: 20,
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer client.Close()

	var (
		failures atomic.Int64
		wg       sync.WaitGroup
		start    = time.Now()
	)

	for i := 0; i < total; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			payload := fmt.Sprintf("payload-%d", index)

			response, err := client.PublisherRPC(&sdk_dto.RabbitOptions{
				ExchangeName: exchange,
				ExchangeType: "topic",
				QueueName:    queue,
				Body:         payload,
				Timeout:      30 * time.Second,
			})
			if err != nil {
				failures.Add(1)
				t.Errorf("request %d failed: %v", index, err)
				return
			}

			// The handler echoes the JSON encoded request body.
			if !containsPayload(string(response), payload) {
				failures.Add(1)
				t.Errorf("request %d received foreign payload %q", index, string(response))
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)

	if failures.Load() > 0 {
		t.Fatalf("%d of %d RPC calls failed", failures.Load(), total)
	}

	t.Logf("%d RPC calls in %s (%.2f rps)", total, elapsed, float64(total)/elapsed.Seconds())
}

func containsPayload(body, payload string) bool {
	if len(body) < len(payload) {
		return false
	}

	for i := 0; i+len(payload) <= len(body); i++ {
		if body[i:i+len(payload)] == payload {
			return true
		}
	}

	return false
}

// TestRPCRecoversAfterMissingExchange verifies that a startup failure is not
// cached forever (the old sync.Once behaviour). The first call fails, the
// second one succeeds once the exchange exists.
func TestRPCRecoversAfterMissingExchange(t *testing.T) {
	url := os.Getenv("RABBITMQ_TEST_URL")
	if url == "" {
		t.Skip("RABBITMQ_TEST_URL is not set")
	}

	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	exchange := "ex.test.missing." + suffix
	queue := "worker.test.missing." + suffix

	defer func() {
		conn, err := amqp091.Dial(url)
		if err != nil {
			return
		}
		defer conn.Close()

		if ch, err := conn.Channel(); err == nil {
			_, _ = ch.QueueDelete(queue, false, false, false)
			_ = ch.ExchangeDelete(exchange, false, false)
		}
	}()

	client, _, err := NewRabbitMQ(ctx, &sdk_dto.RabbitClientOptions{Url: url, PublishTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer client.Close()

	request := &sdk_dto.RabbitOptions{
		ExchangeName: exchange,
		ExchangeType: "topic",
		QueueName:    queue,
		Body:         "ping",
		Timeout:      2 * time.Second,
	}

	if _, err := client.PublisherRPC(request); err == nil {
		t.Fatal("first call must fail while the exchange does not exist")
	}

	connect, err := amqp091.Dial(url)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connect.Close()

	ch, err := connect.Channel()
	if err != nil {
		t.Fatalf("channel: %v", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		t.Fatalf("declare exchange: %v", err)
	}

	go func() {
		_ = client.ConsumerRPC(&sdk_dto.RabbitOptions{
			ExchangeName: exchange,
			ExchangeType: "topic",
			QueueName:    queue,
			Prefetch:     100,
			Concurrency:  10,
		}, func(d amqp.Delivery) amqp.Action {
			reply := &sdk_dto.RabbitOptions{Delivery: d, Body: "pong"}
			_ = client.ReplyToDeliveryPublisher(reply)
			return amqp.Ack
		})
	}()

	time.Sleep(2 * time.Second)

	response, err := client.PublisherRPC(request)
	if err != nil {
		t.Fatalf("second call must succeed after the exchange exists: %v", err)
	}

	if !containsPayload(string(response), "pong") {
		t.Fatalf("unexpected response: %s", response)
	}
}
