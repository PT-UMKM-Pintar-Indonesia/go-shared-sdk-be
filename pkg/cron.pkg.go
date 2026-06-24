package sdk_pkg

import (
	"errors"
	"fmt"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/go-co-op/gocron/v2"
)

type cron struct {
	scheduler gocron.Scheduler
}

func NewCron() (sdk_inf.ICron, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &cron{scheduler: s}, nil
}

func (p *cron) RegisterJob(name, crontime string, task func()) (gocron.Job, error) {
	if name == sdk_cons.EMPTY || crontime == sdk_cons.EMPTY || task == nil {
		return nil, errors.New("invalid job parameters")
	}

	job, err := p.scheduler.NewJob(
		gocron.CronJob(crontime, true),
		gocron.NewTask(task),
		gocron.WithName(name),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create job %s: %w", name, err)
	}

	return job, nil
}

func (p *cron) Start() {
	p.scheduler.Start()
}

func (p *cron) Shutdown() error {
	return p.scheduler.Shutdown()
}
