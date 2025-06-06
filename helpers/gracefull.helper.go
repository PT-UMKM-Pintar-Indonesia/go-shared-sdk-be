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

func Gracefull(name string, ctx context.Context, wg *sync.WaitGroup, handler func(wg *sync.WaitGroup)) {
	handler(wg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGALRM, syscall.SIGABRT, syscall.SIGUSR1)

	for {
		select {
		case <-ch:
			logrus.Info(fmt.Sprintf("Gracefull worker %s is terminated...", name))
			time.Sleep(time.Second * 10)

			if wg != nil {
				wg.Wait()
			}

			os.Exit(0)
			return

		case <-ctx.Done():
			if wg != nil {
				wg.Wait()
			}

			os.Exit(0)
			return

		default:
			if wg != nil {
				wg.Wait()
			}

			time.Sleep(time.Second * 5)
			logrus.Info(fmt.Sprintf("Worker %s is running...", name))
		}
	}
}
