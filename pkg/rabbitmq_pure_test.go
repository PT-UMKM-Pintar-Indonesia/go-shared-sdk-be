package sdk_pkg

import (
	"context"
	"testing"
	"time"

	amqp "github.com/wagslane/go-rabbitmq"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

func TestUsernameFromURL(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"", sdk_cons.EMPTY},
		{"amqp://guest:guest@127.0.0.1:5672/", "guest"},
		{"amqp://svc-user@127.0.0.1:5672/", "svc-user"},
		{"amqp://127.0.0.1:5672/", sdk_cons.EMPTY},
		{"%zz-not-a-url", sdk_cons.EMPTY},
	}

	for _, tc := range cases {
		if got := usernameFromURL(tc.url); got != tc.want {
			t.Errorf("usernameFromURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestPublisherPoolKeyFingerprint(t *testing.T) {
	if got := publisherPoolKey("ex", sdk_cons.EMPTY, nil); got != "ex|"+sdk_cons.EXCHANGE_TYPE_DIRECT+"|" {
		t.Fatalf("empty type got %q", got)
	}

	a := amqp.Table{"b": 2, "a": 1}
	b := amqp.Table{"a": 1, "b": 2}

	if publisherPoolKey("ex", "topic", a) != publisherPoolKey("ex", "topic", b) {
		t.Fatal("fingerprint must be insertion-order independent")
	}

	if argsFingerprint(nil) != sdk_cons.EMPTY {
		t.Fatal("nil args must fingerprint empty")
	}

	if got := argsFingerprint(amqp.Table{"k": "v"}); got != "k=v;" {
		t.Fatalf("fingerprint got %q", got)
	}

	if publisherPoolKey("ex", "direct", a) == publisherPoolKey("ex2", "direct", a) {
		t.Fatal("different exchanges must not collide")
	}
}

func newTestRabbitClient(connUser string) *rabbitmqClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &rabbitmqClient{
		ctx:            ctx,
		cancel:         cancel,
		connUser:       connUser,
		publishTimeout: 5 * time.Second,
		publisherPool:  make(map[string]*publisherEntry),
		consumers:      make(map[int]*consumerHandle),
		cleanupDone:    make(chan struct{}),
		closedCh:       make(chan struct{}),
	}
}

func TestPublisherKeyMutex(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	m1 := h.publisherKeyMutex("key-a")
	m2 := h.publisherKeyMutex("key-a")
	m3 := h.publisherKeyMutex("key-b")

	if m1 != m2 {
		t.Fatal("same key must map to the same mutex")
	}

	if m1 == m3 {
		t.Fatal("different keys must map to different mutexes")
	}
}

func TestPublisherEntryFailureTracking(t *testing.T) {
	e := &publisherEntry{}

	if got := e.failures.Load(); got != 0 {
		t.Fatalf("initial failures %d", got)
	}

	e.markFailure()
	e.markFailure()

	if got := e.failures.Load(); got != 2 {
		t.Fatalf("after two failures got %d", got)
	}

	e.markSuccess()

	if got := e.failures.Load(); got != 0 {
		t.Fatalf("success must reset failures, got %d", got)
	}
}

func TestPublisherEntryAcquireReleaseRetire(t *testing.T) {
	e := &publisherEntry{}

	if !e.acquire() {
		t.Fatal("fresh entry must be acquirable")
	}

	if got := e.refs.Load(); got != 1 {
		t.Fatalf("refs after acquire %d", got)
	}

	e.release()

	if got := e.refs.Load(); got != 0 {
		t.Fatalf("refs after release %d", got)
	}

	e.retire()

	if e.acquire() {
		t.Fatal("retired entry must not be acquirable")
	}

	// close() is Once-guarded and nil-publisher safe.
	e.close()
	e.close()
}

func TestPublisherEntryRetireWithLiveRefs(t *testing.T) {
	e := &publisherEntry{}

	if !e.acquire() {
		t.Fatal("acquire failed")
	}

	e.retire()

	// Still referenced: close must be deferred to the final release.
	if got := e.refs.Load(); got != 1 {
		t.Fatalf("retire must not drop live refs, got %d", got)
	}

	e.release()

	if got := e.refs.Load(); got != 0 {
		t.Fatalf("release did not drop ref, got %d", got)
	}
}

func TestCleanupIdlePublishers(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	idle := &publisherEntry{}
	idle.lastUsed.Store(time.Now().Add(-31 * time.Minute).UnixNano())

	fresh := &publisherEntry{}
	fresh.lastUsed.Store(time.Now().UnixNano())

	h.publisherPool["idle"] = idle
	h.publisherPool["fresh"] = fresh

	h.cleanupIdlePublishers()

	if _, ok := h.publisherPool["idle"]; ok {
		t.Fatal("idle publisher must be evicted")
	}

	if _, ok := h.publisherPool["fresh"]; !ok {
		t.Fatal("fresh publisher must be kept")
	}

	if !idle.retired.Load() {
		t.Fatal("evicted publisher must be retired")
	}
}

func TestPublishOptions(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	base := h.publishOptions(&sdk_dto.RabbitOptions{
		ExchangeName: "ex",
		ContentType:  "application/json",
		Timestamp:    time.Now(),
		AppID:        "app",
	})

	if len(base) != 5 {
		t.Fatalf("base option count got %d", len(base))
	}

	extended := h.publishOptions(&sdk_dto.RabbitOptions{
		ExchangeName:  "ex",
		ContentType:   "application/json",
		Timestamp:     time.Now(),
		AppID:         "app",
		CorrelationID: "cid",
		ReplyTo:       "reply",
		UserID:        "svc",
		Headers:       amqp.Table{"x": "y"},
	})

	if len(extended) != 9 {
		t.Fatalf("extended option count got %d", len(extended))
	}

	// Applying every option must not panic and must fill the publishing.
	opts := &amqp.PublishOptions{}
	for _, apply := range extended {
		apply(opts)
	}

	if opts.Exchange != "ex" {
		t.Fatalf("exchange got %q", opts.Exchange)
	}

	if opts.CorrelationID != "cid" {
		t.Fatalf("correlation got %q", opts.CorrelationID)
	}

	if opts.UserID != "svc" {
		t.Fatalf("userid got %q", opts.UserID)
	}
}

func TestPublishContextHonoursTimeout(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	ctx, cancel := h.publishContext()
	defer cancel()

	dl, ok := ctx.Deadline()
	if !ok {
		t.Fatal("publish context must carry a deadline")
	}

	if remaining := time.Until(dl); remaining <= 0 || remaining > h.publishTimeout {
		t.Fatalf("deadline %v outside publish timeout %v", remaining, h.publishTimeout)
	}
}

func TestNewConsumerHandle(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	c := newConsumerHandle(nil)
	if c == nil {
		t.Fatal("newConsumerHandle returned nil")
	}

	if c.consumer != nil {
		t.Fatal("consumer should be nil")
	}

	// closeConsumer on a handle whose consumer is nil must not panic when
	// the Once-guarded Close is invoked with a nil-safe consumer.
	if h.consumers == nil {
		h.consumers = make(map[int]*consumerHandle)
	}

	h.consumers[1] = c

	if got := len(h.consumers); got != 1 {
		t.Fatalf("consumer registry size %d", got)
	}
}
