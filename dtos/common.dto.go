package sdk_dto

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/wagslane/go-rabbitmq"
)

type (
	ServiceOptions struct {
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	UsecaseOptions[T any] struct {
		SERVICE T
	}

	ControllerOptions[T any] struct {
		USECASE T
	}

	RouteOptions[T any] struct {
		ENV        Request[Environtment]
		RDS        *redis.Client
		ROUTER     chi.Router
		CONTROLLER T
	}

	SchedulerOptions struct {
		CTX  context.Context
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	WorkerOptions struct {
		CTX  context.Context
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	ModuleOptions struct {
		ENV    Request[Environtment]
		DB     *bun.DB
		RDS    *redis.Client
		AMQP   *rabbitmq.Conn
		ROUTER chi.Router
	}
)

type (
	EmailOptions struct {
		Sender     string         `json:"sender,omitempty"`
		Recipients []string       `json:"recipients"`
		Subject    string         `json:"subject"`
		CC         []string       `json:"cc,omitempty"`
		BCC        []string       `json:"bcc,omitempty"`
		Vars       map[string]any `json:"vars,omitempty"`
	}
)
