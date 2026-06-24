package sdk_helper

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type GracefulWorkerOptions struct {
	Name                string
	WaitGroup           *sync.WaitGroup
	Handler             func(ctx context.Context, wg *sync.WaitGroup)
	ShutdownTimeout     time.Duration
	HealthCheckEnabled  bool
	HealthCheckInterval time.Duration
	OnShutdown          func() error
	OnHealthCheck       func() bool
}

const (
	DefaultWorkerShutdownTimeout = 30 * time.Second
	DefaultHealthCheckInterval   = 10 * time.Second
)

func graceful(opts *GracefulWorkerOptions) error {
	if opts == nil || opts.Handler == nil {
		return errors.New("options and handler cannot be nil")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if opts.ShutdownTimeout <= 0 {
		opts.ShutdownTimeout = DefaultWorkerShutdownTimeout
	}

	if opts.HealthCheckEnabled {
		go runHealthCheck(ctx, opts)
	}

	logrus.WithField("name", opts.Name).Info("Worker started")
	go opts.Handler(ctx, opts.WaitGroup)

	<-ctx.Done()
	logrus.WithField("name", opts.Name).Info("Worker initiating graceful shutdown")

	if opts.OnShutdown != nil {
		if err := opts.OnShutdown(); err != nil {
			logrus.WithError(err).WithField("name", opts.Name).Error("Pre-shutdown hook error")
		}
	}

	if opts.WaitGroup != nil {
		done := make(chan struct{})
		go func() {
			opts.WaitGroup.Wait()
			close(done)
		}()

		select {
		case <-done:
			logrus.WithField("name", opts.Name).Info("Worker completed all tasks")
		case <-time.After(opts.ShutdownTimeout):
			logrus.WithField("name", opts.Name).Warn("Shutdown timeout exceeded, forcing exit")
		}
	}

	logrus.WithField("name", opts.Name).Info("Worker gracefully terminated")
	return nil
}

func runHealthCheck(ctx context.Context, opts *GracefulWorkerOptions) {
	interval := opts.HealthCheckInterval
	if interval <= 0 {
		interval = DefaultHealthCheckInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if opts.OnHealthCheck != nil && !opts.OnHealthCheck() {
				logrus.WithField("name", opts.Name).Warn("Health check failed")
			}
		}
	}
}

func GracefulWorker(name string, ctx context.Context, wg *sync.WaitGroup, handler func(ctx context.Context, wg *sync.WaitGroup)) {
	_ = graceful(&GracefulWorkerOptions{
		Name:      name,
		WaitGroup: wg,
		Handler:   handler,
	})
}
