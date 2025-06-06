package sdk_dto

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/wagslane/go-rabbitmq"
)

type (
	ServiceOptions[T any] struct {
		CTX  context.Context
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	UsecaseOptions[T any] struct {
		SERVICE T
	}

	ControllerOptions[T any] struct {
		CTX     context.Context
		ENV     Request[Environtment]
		DB      *bun.DB
		RDS     *redis.Client
		AMQP    *rabbitmq.Conn
		ROUTER  chi.Router
		USECASE T
	}

	RouteOptions[T any] struct {
		CTX        context.Context
		ENV        Request[Environtment]
		DB         *bun.DB
		RDS        *redis.Client
		AMQP       *rabbitmq.Conn
		ROUTER     chi.Router
		CONTROLLER T
	}

	SchedulerOptions[T any] struct {
		CTX  context.Context
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	WorkerOptions[T any] struct {
		CTX  context.Context
		ENV  Request[Environtment]
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	CallbackOptions[T any] struct {
		CTX        context.Context
		ENV        Request[Environtment]
		DB         *bun.DB
		RDS        *redis.Client
		AMQP       *rabbitmq.Conn
		ROUTER     *chi.Mux
		CONTROLLER T
	}

	EventOptions[T any] struct {
		CTX     context.Context
		ENV     Request[Environtment]
		DB      *bun.DB
		RDS     *redis.Client
		AMQP    *rabbitmq.Conn
		USECASE T
	}

	ModuleOptions[T any] struct {
		CTX    context.Context
		ENV    Request[Environtment]
		DB     *bun.DB
		RDS    *redis.Client
		AMQP   *rabbitmq.Conn
		ROUTER chi.Router
	}
)

type (
	EmailOptions struct {
		Channel    string         `json:"channel,omitempty"`
		Sender     string         `json:"sender,omitempty"`
		Type       string         `json:"type"`
		Recipients []string       `json:"recipient"`
		Subject    string         `json:"subject"`
		CC         []string       `json:"cc,omitempty"`
		BCC        []string       `json:"bcc,omitempty"`
		Vars       map[string]any `json:"vars,omitempty"`
	}
)
