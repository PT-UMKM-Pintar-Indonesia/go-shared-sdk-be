package sdk_config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/spf13/viper"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func NewEnvirontment(name, path, ext string, bind any) (*sdk_opt.Environtment, error) {
	cfg := &sdk_dto.Config{}

	if err := loadConfig(name, path, ext, cfg, bind); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return mapToOutput(cfg, bind), nil
}

func loadConfig(name, path, ext string, cfg, bind any) error {
	if _, ok := os.LookupEnv("GO_ENV"); !ok {
		viper.SetConfigName(name)
		viper.SetConfigType(ext)
		viper.AddConfigPath(path)
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		if err := viper.Unmarshal(cfg); err != nil {
			return err
		}

		if bind != nil {
			return viper.Unmarshal(bind)
		}
	} else {
		if err := env.Parse(cfg); err != nil {
			return err
		}

		if bind != nil {
			return env.Parse(bind)
		}
	}

	return nil
}

func mapToOutput(cfg *sdk_dto.Config, bind any) *sdk_opt.Environtment {
	return &sdk_opt.Environtment{
		APP: sdk_opt.Application{
			ENV:          cfg.ENV,
			PORT:         cfg.PORT,
			INBOUND_SIZE: cfg.INBOUND_SIZE,
		},
		REDIS: sdk_opt.Redis{
			URL:      cfg.REDIS_CSN,
			HOST:     cfg.REDIS_HOST,
			PORT:     cfg.REDIS_PORT,
			USER:     cfg.REDIS_USER,
			PASSWORD: cfg.REDIS_PASSWORD,
			DB:       cfg.REDIS_DB,
		},
		POSTGRES: sdk_opt.Postgres{
			URL:      cfg.PG_DSN,
			HOST:     cfg.PG_HOST,
			PORT:     cfg.PG_PORT,
			USER:     cfg.PG_USER,
			PASSWORD: cfg.PG_PASSWORD,
			DB:       cfg.PG_DB,
		},
		MYSQL: sdk_opt.Mysql{
			URL:      cfg.MYSQL_DSN,
			HOST:     cfg.MYSQL_HOST,
			PORT:     cfg.MYSQL_PORT,
			USER:     cfg.MYSQL_USER,
			PASSWORD: cfg.MYSQL_PASSWORD,
			DB:       cfg.MYSQL_DB,
		},
		DBCONFIG: sdk_opt.DatabaseConfig{
			TIMEOUT:       cfg.DB_TIMEOUT,
			DIAL_TIMEOUT:  cfg.DB_DIAL_TIMEOUT,
			READ_TIMEOUT:  cfg.DB_READ_TIMEOUT,
			WRITE_TIMEOUT: cfg.DB_WRITE_TIMEOUT,
			MAX_CONN:      cfg.DB_MAXCONN,
			MAX_IDLE:      cfg.DB_MAXIDLE,
			CON_MAX:       cfg.DB_CONMAX,
			CON_IDLE:      cfg.DB_CONIDLE,
		},
		JWT: sdk_opt.Jwt{
			SECRET:  cfg.JWT_SECRET_KEY,
			EXPIRED: cfg.JWT_EXPIRED,
		},
		RABBITMQ: sdk_opt.RabbitMQ{
			URL:         cfg.RABBITMQ_QSN,
			VSN:         cfg.RABBITMQ_VSN,
			HOST:        cfg.RABBITMQ_HOST,
			PORT:        cfg.RABBITMQ_PORT,
			USER:        cfg.RABBITMQ_USER,
			PASSWORD:    cfg.RABBITMQ_PASSWORD,
			SECRET:      cfg.RABBITMQ_SECRET_KEY,
			CONCURRENCY: cfg.RABBITMQ_CONCURRENCY,
			QOS:         cfg.RABBITMQ_QOS,
		},
		SMTP: sdk_opt.Smtp{
			HOST:     cfg.SMTP_HOST,
			PORT:     cfg.SMTP_PORT,
			USERNAME: cfg.SMTP_USERNAME,
			PASSWORD: cfg.SMTP_PASSWORD,
		},
		BIND: bind,
	}
}
