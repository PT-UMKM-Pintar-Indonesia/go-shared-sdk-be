package sdk_opt

import (
	"context"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/go-chi/chi/v5"

	"github.com/uptrace/bun"
)

type (
	ServiceOptions struct {
		CTX    context.Context
		ENV    *sdk_dto.Environment
		ENVB   any
		DB     *bun.DB
		RDS    sdk_inf.IRedis
		AMQP   sdk_inf.IRabbitMQ
		HELPER any
	}

	UsecaseOptions[T any] struct {
		CTX         context.Context
		ENV         *sdk_dto.Environment
		ENVB        any
		DB          *bun.DB
		RDS         sdk_inf.IRedis
		AMQP        sdk_inf.IRabbitMQ
		REPOSITORIE T
	}

	ControllerOptions[T any] struct {
		CTX     context.Context
		ENV     *sdk_dto.Environment
		ENVB    any
		DB      *bun.DB
		RDS     sdk_inf.IRedis
		AMQP    sdk_inf.IRabbitMQ
		ROUTER  chi.Router
		USECASE T
	}

	RouteOptions[T any] struct {
		CTX        context.Context
		ENV        *sdk_dto.Environment
		ENVB       any
		DB         *bun.DB
		RDS        sdk_inf.IRedis
		ROUTER     chi.Router
		CONTROLLER T
	}

	SchedulerOptions struct {
		CTX  context.Context
		ENV  *sdk_dto.Environment
		ENVB any
		DB   *bun.DB
		RDS  sdk_inf.IRedis
		AMQP sdk_inf.IRabbitMQ
	}

	WorkerOptions struct {
		CTX  context.Context
		ENV  *sdk_dto.Environment
		ENVB any
		DB   *bun.DB
		RDS  sdk_inf.IRedis
		AMQP sdk_inf.IRabbitMQ
	}

	CallbackOptions[T any] struct {
		CTX        context.Context
		ENV        *sdk_dto.Environment
		ENVB       any
		DB         *bun.DB
		AMQP       sdk_inf.IRabbitMQ
		ROUTER     *chi.Mux
		CONTROLLER T
	}

	EventOptions[T any] struct {
		CTX     context.Context
		ENV     *sdk_dto.Environment
		ENVB    any
		DB      *bun.DB
		RDS     sdk_inf.IRedis
		AMQP    sdk_inf.IRabbitMQ
		USECASE T
	}

	ModuleOptions[T any] struct {
		CTX         context.Context
		ENV         *sdk_dto.Environment
		ENVB        any
		DB          *bun.DB
		RDS         sdk_inf.IRedis
		AMQP        sdk_inf.IRabbitMQ
		ROUTER      chi.Router
		REPOSITORIE T
	}

	HelperOptions struct {
		CTX  context.Context
		ENV  *sdk_dto.Environment
		ENVB any
		DB   *bun.DB
		RDS  sdk_inf.IRedis
		AMQP sdk_inf.IRabbitMQ
	}

	MessagingOptions[T any] struct {
		CTX     context.Context
		ENV     *sdk_dto.Environment
		ENVB    any
		AMQP    sdk_inf.IRabbitMQ
		USECASE T
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

	WhatsAppOptions struct {
		Phone    string `json:"phone"`
		Message  string `json:"message"`
		Provider string `json:"provider"`
		Source   string `json:"source,omitempty"`
		RefID    string `json:"refId,omitempty"`
	}
)
