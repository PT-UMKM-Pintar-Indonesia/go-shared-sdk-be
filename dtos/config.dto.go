package sdk_dto

import sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"

type Config struct {
	ENV          string `env:"GO_ENV" mapstructure:"GO_ENV" default:"development"`
	PORT         string `env:"PORT" mapstructure:"PORT" default:"4000"`
	INBOUND_SIZE int    `env:"INBOUND_SIZE" mapstructure:"INBOUND_SIZE" default:"3145728"`

	PG_DSN      string `env:"PG_DSN" mapstructure:"PG_DSN"`
	PG_HOST     string `env:"PG_HOST" mapstructure:"PG_HOST"`
	PG_PORT     int    `env:"PG_PORT" mapstructure:"PG_PORT"`
	PG_USER     string `env:"PG_USER" mapstructure:"PG_USER"`
	PG_PASSWORD string `env:"PG_PASSWORD" mapstructure:"PG_PASSWORD"`
	PG_DB       string `env:"PG_DB" mapstructure:"PG_DB"`

	MYSQL_DSN      string `env:"MYSQL_DSN" mapstructure:"MYSQL_DSN"`
	MYSQL_HOST     string `env:"MYSQL_HOST" mapstructure:"MYSQL_HOST"`
	MYSQL_PORT     int    `env:"MYSQL_PORT" mapstructure:"MYSQL_PORT"`
	MYSQL_USER     string `env:"MYSQL_USER" mapstructure:"MYSQL_USER"`
	MYSQL_PASSWORD string `env:"MYSQL_PASSWORD" mapstructure:"MYSQL_PASSWORD"`
	MYSQL_DB       string `env:"MYSQL_DB" mapstructure:"MYSQL_DB"`

	REDIS_CSN      string `env:"REDIS_CSN" mapstructure:"REDIS_CSN"`
	REDIS_HOST     string `env:"REDIS_HOST" mapstructure:"REDIS_HOST"`
	REDIS_PORT     int    `env:"REDIS_PORT" mapstructure:"REDIS_PORT"`
	REDIS_USER     string `env:"REDIS_USER" mapstructure:"REDIS_USER"`
	REDIS_PASSWORD string `env:"REDIS_PASSWORD" mapstructure:"REDIS_PASSWORD"`
	REDIS_DB       string `env:"REDIS_DB" mapstructure:"REDIS_DB"`

	JWT_SECRET_KEY string `env:"JWT_SECRET_KEY" mapstructure:"JWT_SECRET_KEY"`
	JWT_EXPIRED    int    `env:"JWT_EXPIRED" mapstructure:"JWT_EXPIRED"`

	RABBITMQ_QSN        string `env:"RABBITMQ_QSN" mapstructure:"RABBITMQ_QSN"`
	RABBITMQ_VSN        string `env:"RABBITMQ_VSN" mapstructure:"RABBITMQ_VSN"`
	RABBITMQ_HOST       string `env:"RABBITMQ_HOST" mapstructure:"RABBITMQ_HOST"`
	RABBITMQ_PORT       int    `env:"RABBITMQ_PORT" mapstructure:"RABBITMQ_PORT"`
	RABBITMQ_USER       string `env:"RABBITMQ_USER" mapstructure:"RABBITMQ_USER"`
	RABBITMQ_PASSWORD   string `env:"RABBITMQ_PASSWORD" mapstructure:"RABBITMQ_PASSWORD"`
	RABBITMQ_SECRET_KEY string `env:"RABBITMQ_SECRET_KEY" mapstructure:"RABBITMQ_SECRET_KEY"`

	SMTP_HOST     string `env:"SMTP_HOST" mapstructure:"SMTP_HOST"`
	SMTP_PORT     int    `env:"SMTP_PORT" mapstructure:"SMTP_PORT"`
	SMTP_USERNAME string `env:"SMTP_USERNAME" mapstructure:"SMTP_USERNAME"`
	SMTP_PASSWORD string `env:"SMTP_PASSWORD" mapstructure:"SMTP_PASSWORD"`
}

type (
	Environtment struct {
		APP      sdk_opt.Application
		REDIS    sdk_opt.Redis
		POSTGRES sdk_opt.Postgres
		MYSQL    sdk_opt.Mysql
		JWT      sdk_opt.Jwt
		RABBITMQ sdk_opt.RabbitMQ
		SMTP     sdk_opt.Smtp
		BIND     any
	}
)
