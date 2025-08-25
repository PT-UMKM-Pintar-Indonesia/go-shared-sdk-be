package sdk_con

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

type connectionPoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func defaultConnectionPoolConfig(req sdk_dto.Request[sdk_dto.Environtment]) connectionPoolConfig {
	return connectionPoolConfig{
		MaxOpenConns:    req.Config.DBCONFIG.MAX_CONN,
		MaxIdleConns:    req.Config.DBCONFIG.MAX_IDLE,
		ConnMaxLifetime: time.Duration(req.Config.DBCONFIG.CON_MAX) * time.Minute,
		ConnMaxIdleTime: time.Duration(req.Config.DBCONFIG.CON_IDLE) * time.Minute,
	}
}

func sqlConnectionWithConfig(ctx context.Context, driver string, req sdk_dto.Request[sdk_dto.Environtment], poolConfig connectionPoolConfig) (*bun.DB, error) {
	var (
		bundb *bun.DB
		db    *sql.DB
		err   error
	)

	switch driver {
	case sdk_cons.POSTGRES:
		connector := pgdriver.NewConnector(
			pgdriver.WithDSN(req.Config.POSTGRES.URL),
			pgdriver.WithTimeout(time.Duration(req.Config.DBCONFIG.TIMEOUT)*time.Second),
			pgdriver.WithDialTimeout(time.Duration(req.Config.DBCONFIG.DIAL_TIMEOUT)*time.Second),
			pgdriver.WithReadTimeout(time.Duration(req.Config.DBCONFIG.READ_TIMEOUT)*time.Second),
			pgdriver.WithWriteTimeout(time.Duration(req.Config.DBCONFIG.WRITE_TIMEOUT)*time.Second),
		)
		db = sql.OpenDB(connector)

	case sdk_cons.MYSQL:
		db, err = sql.Open(sdk_cons.MYSQL, req.Config.MYSQL.URL)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("driver unsupported")
	}

	db.SetMaxOpenConns(poolConfig.MaxOpenConns)
	db.SetMaxIdleConns(poolConfig.MaxIdleConns)
	db.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	switch driver {
	case sdk_cons.POSTGRES:
		bundb = bun.NewDB(db, pgdialect.New())

	case sdk_cons.MYSQL:
		bundb = bun.NewDB(db, mysqldialect.New())
	}

	if req.Config.APP.ENV != sdk_cons.PROD {
		bundb.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithEnabled(true),
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	return bundb, nil
}

func SqlConnection(ctx context.Context, driver string, req sdk_dto.Request[sdk_dto.Environtment]) (*bun.DB, error) {
	return sqlConnectionWithConfig(ctx, driver, req, defaultConnectionPoolConfig(req))
}
