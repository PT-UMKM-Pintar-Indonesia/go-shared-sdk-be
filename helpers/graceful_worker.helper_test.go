package sdk_helper

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// quietLogrus silences the package-level logger for the duration of a test
// so worker lifecycle logs do not pollute test output.
func quietLogrus(t *testing.T) {
	t.Helper()

	prev := logrus.StandardLogger().Out
	logrus.SetOutput(io.Discard)

	t.Cleanup(func() { logrus.SetOutput(prev) })
}

func TestGracefulWorkerNilOptions(t *testing.T) {
	if err := graceful(nil); err == nil {
		t.Fatal("nil options must error")
	}

	if err := graceful(&GracefulWorkerOptions{}); err == nil {
		t.Fatal("nil handler must error")
	}

	// The public wrapper swallows the error and must not panic.
	GracefulWorker(nil)
}

func TestRunHealthCheck(t *testing.T) {
	quietLogrus(t)

	called := make(chan struct{}, 4)

	opts := &GracefulWorkerOptions{
		Name:                "hc",
		HealthCheckInterval: 10 * time.Millisecond,
		OnHealthCheck: func() bool {
			select {
			case called <- struct{}{}:
			default:
			}

			return false // exercise the "failed" warning branch
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		runHealthCheck(ctx, opts)
		close(done)
	}()

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("health check never ran")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("health check loop did not exit on cancel")
	}

	// Zero interval falls back to the default and exits with the context.
	idleCtx, idleCancel := context.WithCancel(context.Background())
	idleCancel()

	runHealthCheck(idleCtx, &GracefulWorkerOptions{Name: "idle"})
}

func TestGracefulWorkerSignalShutdown(t *testing.T) {
	quietLogrus(t)

	var (
		onShutdownRan atomic.Bool
		handlerDone   = make(chan struct{})
		started       = make(chan struct{})
	)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	opts := &GracefulWorkerOptions{
		Name:                "worker",
		WaitGroup:           wg,
		ShutdownTimeout:     3 * time.Second,
		HealthCheckEnabled:  true,
		HealthCheckInterval: 20 * time.Millisecond,
		Handler: func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(handlerDone)

			close(started)

			<-ctx.Done()
		},
		OnShutdown: func() error {
			onShutdownRan.Store(true)

			return errors.New("shutdown hook boom") // error branch is logged, not fatal
		},
		OnHealthCheck: func() bool { return true },
	}

	resCh := make(chan error, 1)
	go func() { resCh <- graceful(opts) }()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler never started")
	}

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("signal: %v", err)
	}

	select {
	case <-handlerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not observe cancellation")
	}

	select {
	case err := <-resCh:
		if err != nil {
			t.Fatalf("graceful returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("graceful did not return")
	}

	if !onShutdownRan.Load() {
		t.Fatal("OnShutdown hook must run")
	}
}

func TestGracefulWorkerShutdownTimeout(t *testing.T) {
	quietLogrus(t)

	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	opts := &GracefulWorkerOptions{
		Name:            "stuck",
		WaitGroup:       wg,
		ShutdownTimeout: 200 * time.Millisecond,
		Handler: func(ctx context.Context, wg *sync.WaitGroup) {
			close(started)
			<-release // deliberately ignores ctx; forces the timeout branch
			wg.Done()
		},
	}

	resCh := make(chan error, 1)
	go func() { resCh <- graceful(opts) }()

	<-started

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-resCh:
		if err != nil {
			t.Fatalf("forced exit must still return nil, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown timeout branch never completed")
	}
}

func TestGracefulWorkerPublicRunsHandler(t *testing.T) {
	quietLogrus(t)

	started := make(chan struct{})
	finished := make(chan struct{})

	var ran atomic.Bool

	opts := &GracefulWorkerOptions{
		Name:            "public",
		ShutdownTimeout: time.Second,
		Handler: func(ctx context.Context, _ *sync.WaitGroup) {
			close(started)
			<-ctx.Done()
			close(finished)
		},
		OnShutdown: func() error {
			ran.Store(true)

			return nil
		},
	}

	go GracefulWorker(opts)

	<-started

	if err := syscall.Kill(os.Getpid(), syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}

	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("GracefulWorker did not propagate shutdown")
	}

	// Give graceful() a moment to finish hooks before asserting.
	deadline := time.Now().Add(2 * time.Second)
	for !ran.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !ran.Load() {
		t.Fatal("OnShutdown via GracefulWorker never ran")
	}
}
