package sdk_helper

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func discardLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)

	return l
}

func TestGracefulOptionsValidateDefaults(t *testing.T) {
	opts := &GracefulOptions{}
	opts.validate()

	if opts.Logger == nil {
		t.Fatal("logger must be defaulted")
	}

	if opts.Address != "0.0.0.0" {
		t.Fatalf("address got %q", opts.Address)
	}

	if opts.ReadTimeout != 15*time.Second {
		t.Fatalf("read timeout got %v", opts.ReadTimeout)
	}

	if opts.ShutdownTimeout != 25*time.Second {
		t.Fatalf("shutdown timeout got %v", opts.ShutdownTimeout)
	}

	if opts.ReadinessDrainDelay != 5*time.Second {
		t.Fatalf("drain delay got %v", opts.ReadinessDrainDelay)
	}

	// Explicit values survive validation.
	custom := &GracefulOptions{
		Address:             "127.0.0.1",
		ReadTimeout:         time.Second,
		ShutdownTimeout:     time.Second,
		ReadinessDrainDelay: time.Second,
		Logger:              discardLogger(),
	}
	custom.validate()

	if custom.Address != "127.0.0.1" || custom.ReadTimeout != time.Second {
		t.Fatal("validate must not overwrite explicit values")
	}
}

func TestGracefulOptionsIsReady(t *testing.T) {
	opts := &GracefulOptions{}

	if opts.IsReady() {
		t.Fatal("fresh options must not be ready")
	}

	opts.ready.Store(true)

	if !opts.IsReady() {
		t.Fatal("IsReady must reflect the flag")
	}
}

func TestDrainWithCountdown(t *testing.T) {
	logger := discardLogger().WithField("test", true)

	// Non-positive drain returns immediately.
	done := make(chan struct{})
	go func() {
		drainWithCountdown(context.Background(), 0, logger)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("zero drain blocked")
	}

	// A cancelled context bypasses the wait.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done = make(chan struct{})
	go func() {
		drainWithCountdown(ctx, time.Hour, logger)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled drain did not bypass")
	}

	// A one-second drain completes on the first tick.
	start := time.Now()
	drainWithCountdown(context.Background(), time.Second, logger)

	if elapsed := time.Since(start); elapsed < 900*time.Millisecond || elapsed > 3*time.Second {
		t.Fatalf("drain took %v", elapsed)
	}
}

func TestGracefulHTTPNilOptions(t *testing.T) {
	if err := GracefulHTTP(context.Background(), nil); err == nil {
		t.Fatal("nil options must error")
	}
}

func TestGracefulHTTPBindFailure(t *testing.T) {
	err := GracefulHTTP(context.Background(), &GracefulOptions{
		Address: "256.256.256.256",
		Port:    "80",
		Logger:  discardLogger(),
	})
	if err == nil {
		t.Fatal("invalid bind address must error")
	}
}

func TestGracefulHTTPServeAndShutdown(t *testing.T) {
	// Reserve a free port, release it, then let the server bind it.
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	port := strconv.Itoa(probe.Addr().(*net.TCPAddr).Port)
	probe.Close()

	var (
		preRan   atomic.Bool
		postRan  atomic.Bool
		metricsR atomic.Bool
	)

	opts := &GracefulOptions{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		Address:             "127.0.0.1",
		Port:                port,
		ReadinessDrainDelay: time.Second,
		ShutdownTimeout:     2 * time.Second,
		Logger:              discardLogger(),
		PreShutdownHook:     func() { preRan.Store(true) },
		PostShutdownHook:    func() { postRan.Store(true) },
		MetricsCollector:    func(uptimeMs int64) { metricsR.Store(true) },
	}

	ctx, cancel := context.WithCancel(context.Background())

	resCh := make(chan error, 1)
	go func() { resCh <- GracefulHTTP(ctx, opts) }()

	deadline := time.Now().Add(3 * time.Second)
	for !opts.IsReady() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !opts.IsReady() {
		t.Fatal("server never became ready")
	}

	resp, err := http.Get("http://127.0.0.1:" + port + "/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if string(body) != "ok" {
		t.Fatalf("body %q", body)
	}

	// Cancelling the parent context trips the signal path and shuts down.
	cancel()

	select {
	case err := <-resCh:
		if err != nil {
			t.Fatalf("graceful shutdown err %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete")
	}

	if opts.IsReady() {
		t.Fatal("readiness must be cleared after shutdown")
	}

	if !preRan.Load() || !postRan.Load() || !metricsR.Load() {
		t.Fatalf("hooks pre=%v post=%v metrics=%v", preRan.Load(), postRan.Load(), metricsR.Load())
	}
}
