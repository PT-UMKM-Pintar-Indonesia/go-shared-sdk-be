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
		CTX    context.Context
		ENV    Request[Environment]
		ENVB   any
		DB     *bun.DB
		RDS    *redis.Client
		AMQP   *rabbitmq.Conn
		HELPER any
	}

	UsecaseOptions[T any] struct {
		SERVICE T
	}

	ControllerOptions[T any] struct {
		CTX     context.Context
		ENV     Request[Environment]
		ENVB    any
		DB      *bun.DB
		RDS     *redis.Client
		AMQP    *rabbitmq.Conn
		ROUTER  chi.Router
		USECASE T
	}

	RouteOptions[T any] struct {
		CTX        context.Context
		ENV        Request[Environment]
		ENVB       any
		DB         *bun.DB
		RDS        *redis.Client
		AMQP       *rabbitmq.Conn
		ROUTER     chi.Router
		CONTROLLER T
	}

	SchedulerOptions struct {
		CTX  context.Context
		ENV  Request[Environment]
		ENVB any
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	WorkerOptions struct {
		CTX  context.Context
		ENV  Request[Environment]
		ENVB any
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
	}

	CallbackOptions[T any] struct {
		CTX        context.Context
		ENV        Request[Environment]
		ENVB       any
		DB         *bun.DB
		RDS        *redis.Client
		AMQP       *rabbitmq.Conn
		ROUTER     *chi.Mux
		CONTROLLER T
	}

	EventOptions[T any] struct {
		CTX     context.Context
		ENV     Request[Environment]
		ENVB    any
		DB      *bun.DB
		RDS     *redis.Client
		AMQP    *rabbitmq.Conn
		USECASE T
	}

	ModuleOptions struct {
		CTX    context.Context
		ENV    Request[Environment]
		ENVB   any
		DB     *bun.DB
		RDS    *redis.Client
		AMQP   *rabbitmq.Conn
		ROUTER chi.Router
	}

	HelperOptions struct {
		CTX  context.Context
		ENV  Request[Environment]
		ENVB any
		DB   *bun.DB
		RDS  *redis.Client
		AMQP *rabbitmq.Conn
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

type (
	Shipper struct {
		Name         string `json:"name"`
		Address      string `json:"address"`
		Latitude     string `json:"latitude,omitempty"`
		Longitude    string `json:"longitude,omitempty"`
		MobileNumber string `json:"mobile_number,omitempty"`
		Poi          string `json:"poi,omitempty"`
	}

	Receiver struct {
		Name         string `json:"name"`
		Address      string `json:"address"`
		Latitude     string `json:"latitude,omitempty"`
		Longitude    string `json:"longitude,omitempty"`
		MobileNumber string `json:"mobile_number,omitempty"`
		Poi          string `json:"poi,omitempty"`
		ItemType     string `json:"item_type"`
		ItemNote     string `json:"item_note,omitempty"`
	}

	ShippingOptions struct {
		Shipper        Shipper        `json:"shipper"`
		Receiver       Receiver       `json:"receiver"`
		RequestID      string         `json:"request_id"`
		AdditionalInfo map[string]any `json:"additional_info,omitempty"`
	}
)
