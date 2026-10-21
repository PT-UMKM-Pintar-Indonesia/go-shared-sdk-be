package sdk_pkg

import (
	"fmt"
	"sync"
	"testing"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
	amqp "github.com/wagslane/go-rabbitmq"
)

func newTestClient() *rabbitmqClient {
	return &rabbitmqClient{
		replyQueueName: "rpc.reply.test",
		instanceID:     "test",
		publisherPool:  make(map[string]*publisherEntry),
		consumers:      make(map[int]*consumerHandle),
		closedCh:       make(chan struct{}),
	}
}

func delivery(correlationID, body string) amqp.Delivery {
	return amqp.Delivery{Delivery: amqp091.Delivery{CorrelationId: correlationID, Body: []byte(body)}}
}

func TestHandleReplyDeliversOnce(t *testing.T) {
	client := newTestClient()

	req := &rpcRequest{ch: make(chan []byte, 1), createdAt: time.Now(), deadline: time.Now().Add(time.Second)}
	client.requests.Store("corr-1", req)

	if action := client.handleReply(delivery("corr-1", "first")); action != amqp.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}

	select {
	case body := <-req.ch:
		if string(body) != "first" {
			t.Fatalf("expected first, got %s", body)
		}
	case <-time.After(time.Second):
		t.Fatal("reply was not delivered")
	}

	// A duplicate reply must be acknowledged without blocking or panicking.
	if action := client.handleReply(delivery("corr-1", "second")); action != amqp.Ack {
		t.Fatalf("expected Ack for duplicate, got %v", action)
	}

	select {
	case body := <-req.ch:
		t.Fatalf("unexpected duplicate delivery: %s", body)
	default:
	}
}

func TestHandleReplyIgnoresSettledRequest(t *testing.T) {
	client := newTestClient()

	req := &rpcRequest{ch: make(chan []byte, 1), createdAt: time.Now(), deadline: time.Now()}
	req.settled.Store(true)
	client.requests.Store("corr-2", req)

	client.handleReply(delivery("corr-2", "late"))

	select {
	case body := <-req.ch:
		t.Fatalf("late reply must be dropped, got %s", body)
	default:
	}
}

func TestHandleReplyUnknownCorrelationID(t *testing.T) {
	client := newTestClient()

	if action := client.handleReply(delivery("unknown", "x")); action != amqp.Ack {
		t.Fatalf("expected Ack, got %v", action)
	}
}

// TestHandleReplyConcurrentResponses mimics the worker scenario: many in-flight
// RPC calls (> prefetch) answered concurrently. Every caller must receive its
// own payload.
func TestHandleReplyConcurrentResponses(t *testing.T) {
	const total = 500

	client := newTestClient()

	var wg sync.WaitGroup

	for i := 0; i < total; i++ {
		req := &rpcRequest{ch: make(chan []byte, 1), createdAt: time.Now(), deadline: time.Now().Add(time.Minute)}
		correlationID := fmt.Sprintf("corr-%d", i)

		client.requests.Store(correlationID, req)

		wg.Add(1)

		go func(index int, id string, request *rpcRequest) {
			defer wg.Done()

			client.handleReply(delivery(id, id))

			select {
			case body := <-request.ch:
				if string(body) != id {
					t.Errorf("correlation %s received foreign payload %s", id, body)
				}
			case <-time.After(2 * time.Second):
				t.Errorf("correlation %s timed out", id)
			}

			client.requests.Delete(id)
		}(i, correlationID, req)
	}

	wg.Wait()
}

func TestCleanupExpiredRequests(t *testing.T) {
	client := newTestClient()

	expired := &rpcRequest{ch: make(chan []byte, 1), createdAt: time.Now().Add(-time.Hour), deadline: time.Now().Add(-time.Hour)}
	fresh := &rpcRequest{ch: make(chan []byte, 1), createdAt: time.Now(), deadline: time.Now().Add(time.Minute)}

	client.requests.Store("expired", expired)
	client.requests.Store("fresh", fresh)

	client.cleanupExpiredRequests()

	if _, ok := client.requests.Load("expired"); ok {
		t.Fatal("expired request should be removed")
	}

	if _, ok := client.requests.Load("fresh"); !ok {
		t.Fatal("fresh request should be kept")
	}

	if !expired.settled.Load() {
		t.Fatal("expired request must be settled")
	}
}

func TestPublisherEntryRefCounting(t *testing.T) {
	entry := &publisherEntry{}

	if !entry.acquire() {
		t.Fatal("acquire should succeed before retirement")
	}

	entry.retire()

	if entry.refs.Load() != 1 {
		t.Fatalf("refs must stay 1 while in use, got %d", entry.refs.Load())
	}

	if entry.acquire() {
		t.Fatal("acquire must fail on a retired publisher")
	}

	entry.release()

	if entry.refs.Load() != 0 {
		t.Fatalf("refs must be 0 after release, got %d", entry.refs.Load())
	}
}
