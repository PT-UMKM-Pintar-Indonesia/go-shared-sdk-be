package sdk_dto

import sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"

type Config struct {
	ENV            string `env:"GO_ENV" mapstructure:"GO_ENV" default:"development"`
	PORT           string `env:"PORT" mapstructure:"PORT" default:"4000"`
	INBOUND_SIZE   int    `env:"INBOUND_SIZE" mapstructure:"INBOUND_SIZE" default:"3145728"`
	PG_DSN         string `env:"PG_DSN" mapstructure:"PG_DSN"`
	REDIS_CSN      string `env:"REDIS_CSN" mapstructure:"REDIS_CSN"`
	JWT_SECRET_KEY string `env:"JWT_SECRET_KEY" mapstructure:"JWT_SECRET_KEY"`
	JWT_EXPIRED    int    `env:"JWT_EXPIRED" mapstructure:"JWT_EXPIRED"`
	RABBITMQ_QSN   string `env:"RABBITMQ_QSN" mapstructure:"RABBITMQ_QSN"`
	RABBITMQ_VSN   string `env:"RABBITMQ_VSN" mapstructure:"RABBITMQ_VSN"`
	SMTP_HOST      string `env:"SMTP_HOST" mapstructure:"SMTP_HOST"`
	SMTP_PORT      int    `env:"SMTP_PORT" mapstructure:"SMTP_PORT"`
	SMTP_USERNAME  string `env:"SMTP_USERNAME" mapstructure:"SMTP_USERNAME"`
	SMTP_PASSWORD  string `env:"SMTP_PASSWORD" mapstructure:"SMTP_PASSWORD"`
}

type (
	Environtment struct {
		APP      sdk_opt.Application
		REDIS    sdk_opt.Redis
		POSTGRES sdk_opt.Postgres
		JWT      sdk_opt.Jwt
		RABBITMQ sdk_opt.RabbitMQ
		SMTP     sdk_opt.Smtp
		BIND     any
	}
)
