package sdk_config

import (
	"os"

	genv "github.com/caarlos0/env"
	"github.com/spf13/viper"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func NewEnvirontment[T any](name, path, ext string, bind any) (sdk_opt.Environtment[T], error) {
	cfg := sdk_dto.Config{}

	if _, ok := os.LookupEnv("GO_ENV"); !ok {
		viper.SetConfigName(name)
		viper.SetConfigType(ext)
		viper.AddConfigPath(path)
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			return sdk_opt.Environtment[T]{}, err
		}

		if err := viper.Unmarshal(&cfg); err != nil {
			return sdk_opt.Environtment[T]{}, err
		}

		if bind != nil {
			if err := viper.Unmarshal(&bind); err != nil {
				return sdk_opt.Environtment[T]{}, err
			}
		}

	} else {
		if err := genv.Parse(&cfg); err != nil {
			return sdk_opt.Environtment[T]{}, err
		}
	}

	return sdk_opt.Environtment[T]{
		APP: sdk_opt.Application{
			ENV:          cfg.ENV,
			PORT:         cfg.PORT,
			INBOUND_SIZE: cfg.INBOUND_SIZE,
		},
		REDIS: sdk_opt.Redis{
			URL: cfg.REDIS_CSN,
		},
		POSTGRES: sdk_opt.Postgres{
			URL: cfg.PG_DSN,
		},
		JWT: sdk_opt.Jwt{
			SECRET:  cfg.JWT_SECRET_KEY,
			EXPIRED: cfg.JWT_EXPIRED,
		},
		RABBITMQ: sdk_opt.RabbitMQ{
			URL: cfg.RABBITMQ_QSN,
			VSN: cfg.RABBITMQ_VSN,
		},
		SMTP: sdk_opt.Smtp{
			HOST:     cfg.SMTP_HOST,
			PORT:     cfg.SMTP_PORT,
			USERNAME: cfg.SMTP_USERNAME,
			PASSWORD: cfg.SMTP_PASSWORD,
		},
		BIND: any(bind).(T),
	}, nil
}
