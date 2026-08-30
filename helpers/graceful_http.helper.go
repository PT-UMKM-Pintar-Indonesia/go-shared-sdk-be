package sdk_helper

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type GracefulOptions struct {
	Handler             http.Handler
	Address             string
	Port                string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	KeepAlive           time.Duration
	ReadinessDrainDelay time.Duration
	Logger              *logrus.Logger
	PreShutdownHook     func()
	PostShutdownHook    func()
	MetricsCollector    func(uptimeMs int64)

	ready atomic.Bool
}

func (o *GracefulOptions) IsReady() bool {
	return o.ready.Load()
}

func (o *GracefulOptions) validate() {
	if o.Logger == nil {
		o.Logger = logrus.New()
		o.Logger.SetLevel(logrus.InfoLevel)
	}
	if o.Address == "" {
		o.Address = "0.0.0.0"
	}
	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 15 * time.Second
	}
	if o.ShutdownTimeout <= 0 {
		o.ShutdownTimeout = 25 * time.Second
	}
	if o.ReadinessDrainDelay <= 0 {
		o.ReadinessDrainDelay = 5 * time.Second
	}
}

func drainWithCountdown(ctx context.Context, d time.Duration, logger *logrus.Entry) {
	if d <= 0 {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(d)
	logger.WithField("duration", d.String()).Info("⏳ Draining traffic: waiting for load balancer to sync")

	for {
		select {
		case <-ctx.Done():
			logger.Warn("⚠️ Drain phase interrupted or bypassed")
			return
		case <-ticker.C:
			remaining := time.Until(deadline).Round(time.Second)
			if remaining <= 0 {
				logger.Info("✅ Drain phase completed")
				return
			}
			logger.WithField("remaining", remaining.String()).Debug("⏳ Draining...")
		}
	}
}

func GracefulHTTP(ctx context.Context, opts *GracefulOptions) error {
	if opts == nil {
		return errors.New("options cannot be nil")
	}
	opts.validate()

	startTime := time.Now()
	addr := net.JoinHostPort(opts.Address, opts.Port)

	logger := opts.Logger.WithFields(logrus.Fields{
		"addr":      addr,
		"component": "graceful_server",
	})

	lc := net.ListenConfig{KeepAlive: opts.KeepAlive}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind listener: %w", err)
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      opts.Handler,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		opts.ready.Store(true)
		logger.Info("🚀 Server is ready and listening")

		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	drainCtx, cancelDrain := context.WithCancel(context.Background())
	defer cancelDrain()

	select {
	case err := <-errCh:
		return fmt.Errorf("server crashed: %w", err)
	case <-sigCtx.Done():
		logger.Warn("🔔 Termination signal received")

		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			select {
			case <-sigCh:
				logger.Warn("🔴 Secondary signal received! Bypassing drain phase")
				cancelDrain()
			case <-drainCtx.Done():
			}
		}()
	}

	if opts.PreShutdownHook != nil {
		opts.PreShutdownHook()
	}

	opts.ready.Store(false)
	logger.Info("📡 Readiness set to false. Starting drain countdown...")

	drainWithCountdown(drainCtx, opts.ReadinessDrainDelay, logger)

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), opts.ShutdownTimeout)
	defer cancelShutdown()

	var shutdownErrs []error
	logger.Info("🔄 Closing active connections...")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		shutdownErrs = append(shutdownErrs, fmt.Errorf("graceful shutdown failed: %w", err))
		if err := srv.Close(); err != nil {
			shutdownErrs = append(shutdownErrs, fmt.Errorf("force close failed: %w", err))
		}
	}

	if opts.PostShutdownHook != nil {
		opts.PostShutdownHook()
	}

	uptime := time.Since(startTime)
	if opts.MetricsCollector != nil {
		opts.MetricsCollector(uptime.Milliseconds())
	}

	logger.WithField("uptime", uptime.Round(time.Second).String()).Info("🛑 Server stopped")
	return errors.Join(shutdownErrs...)
}
