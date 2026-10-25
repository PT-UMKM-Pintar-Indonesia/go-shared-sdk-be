package sdk_pkg

import (
	"errors"
	"testing"
)

func TestRabbitLoggerWrappers(t *testing.T) {
	// Fatalf is deliberately not exercised: logrus.Fatalf exits the process.
	l := rabbitLogger{}

	l.Errorf("errorf %s", "x")
	l.Infof("infof %s", "x")
	l.Warnf("warnf %s", "x")
	l.Debugf("debugf %s", "x")
}

func TestAcquirePublisher(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	if got := h.acquirePublisher("missing"); got != nil {
		t.Fatal("unknown key must return nil")
	}

	entry := &publisherEntry{exchange: "ex"}
	h.publisherPool["ex|direct|"] = entry

	got := h.acquirePublisher("ex|direct|")
	if got != entry {
		t.Fatal("pooled entry must be returned")
	}

	if entry.refs.Load() != 1 {
		t.Fatalf("acquire must take a ref, got %d", entry.refs.Load())
	}

	if entry.lastUsed.Load() == 0 {
		t.Fatal("acquire must refresh lastUsed")
	}

	entry.release()
	entry.retire()

	if got := h.acquirePublisher("ex|direct|"); got != nil {
		t.Fatal("retired entry must not be acquirable")
	}
}

func TestReleasePublisher(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	// nil entry is a no-op.
	h.releasePublisher(nil, nil)

	ok := &publisherEntry{}
	h.publisherPool["ok"] = ok

	h.releasePublisher(ok, nil)

	if ok.failures.Load() != 0 {
		t.Fatal("success must not record failures")
	}

	failing := &publisherEntry{}
	h.publisherPool["failing"] = failing

	for i := 0; i < maxPublishFailures; i++ {
		h.releasePublisher(failing, errors.New("publish failed"))
	}

	if failing.failures.Load() != maxPublishFailures {
		t.Fatalf("failure count %d", failing.failures.Load())
	}

	if _, stillPooled := h.publisherPool["failing"]; stillPooled {
		t.Fatal("entry must be evicted after repeated failures")
	}

	if !failing.retired.Load() {
		t.Fatal("evicted entry must be retired")
	}
}

func TestGetOrCreatePublisherClosedClient(t *testing.T) {
	h := newTestRabbitClient("svc")
	defer h.cancel()

	h.closed.Store(true)

	if _, err := h.getOrCreatePublisher("ex", "direct", nil); !errors.Is(err, errClientClosed) {
		t.Fatalf("got %v want errClientClosed", err)
	}
}
