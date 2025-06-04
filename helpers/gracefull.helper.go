package sdk_helper

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func Gracefull(name string, ctx context.Context, handler func(wg *sync.WaitGroup)) {
	wg := sync.WaitGroup{}
	handler(&wg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM, syscall.SIGABRT, syscall.SIGUSR1)

	for {
		select {
		case <-ch:
			logrus.Info(fmt.Sprintf("Gracefull worker %s is terminated...", name))
			time.Sleep(time.Second * 10)

			wg.Wait()
			os.Exit(0)
			return

		case <-ctx.Done():
			wg.Wait()
			return

		default:
			wg.Wait()
			time.Sleep(time.Second * 5)
			logrus.Info(fmt.Sprintf("Worker %s is running...", name))
		}
	}
}
