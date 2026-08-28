package sdk_con

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	_ "github.com/go-sql-driver/mysql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type connectionPoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func defaultConnectionPoolConfig(env *sdk_dto.Environment) *connectionPoolConfig {
	return &connectionPoolConfig{
		MaxOpenConns:    env.DBCONFIG.MAX_CONN,
		MaxIdleConns:    env.DBCONFIG.MAX_IDLE,
		ConnMaxLifetime: time.Duration(env.DBCONFIG.CON_MAX) * time.Second,
		ConnMaxIdleTime: time.Duration(env.DBCONFIG.CON_IDLE) * time.Second,
	}
}

func sqlConnectionWithConfig(ctx context.Context, appName, driver string, env *sdk_dto.Environment, poolConfig *connectionPoolConfig) (*bun.DB, error) {
	var (
		bundb *bun.DB
		db    *sql.DB
		err   error
	)

	switch driver {
	case sdk_cons.POSTGRES:
		connector := pgdriver.NewConnector(
			pgdriver.WithDSN(env.POSTGRES.URL),
			pgdriver.WithApplicationName(appName),
			pgdriver.WithTimeout(time.Duration(env.DBCONFIG.TIMEOUT)*time.Second),
			pgdriver.WithDialTimeout(time.Duration(env.DBCONFIG.DIAL_TIMEOUT)*time.Second),
			pgdriver.WithReadTimeout(time.Duration(env.DBCONFIG.READ_TIMEOUT)*time.Second),
			pgdriver.WithWriteTimeout(time.Duration(env.DBCONFIG.WRITE_TIMEOUT)*time.Second),
		)

		db = sql.OpenDB(connector)

	case sdk_cons.MYSQL:
		db, err = sql.Open(sdk_cons.MYSQL, env.MYSQL.URL)
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

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err = db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	switch driver {
	case sdk_cons.POSTGRES:
		bundb = bun.NewDB(db, pgdialect.New())
	case sdk_cons.MYSQL:
		bundb = bun.NewDB(db, mysqldialect.New())
	}

	if env.APP.ENV != sdk_cons.PROD {
		bundb.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithEnabled(true),
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	return bundb, nil
}

func SqlConnection(ctx context.Context, appName, driver string, env *sdk_dto.Environment) (*bun.DB, error) {
	return sqlConnectionWithConfig(ctx, appName, driver, env, defaultConnectionPoolConfig(env))
}
