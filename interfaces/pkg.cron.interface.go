package sdk_inf

import "github.com/go-co-op/gocron/v2"

type ICron interface {
	RegisterJob(name, crontime string, task func()) (gocron.Job, error)
}
