package sdk_con

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/schema"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
)

type DBConfig struct {
	DSN             string
	Timeout         time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	IsProduction    bool
}

type DBOption func(*sql.DB)

type DriverFactory func(cfg *DBConfig) (*sql.DB, schema.Dialect, error)

var driverRegistry = map[string]DriverFactory{
	sdk_cons.POSTGRES: createPostgres,
	sdk_cons.MYSQL:    createMySQL,
	sdk_cons.SQLITE:   createSqlite,
	sdk_cons.SQLITE3:  createSqlite3,
}

func SqlConnection(ctx context.Context, driver string, cfg *DBConfig, opts ...DBOption) (*bun.DB, error) {
	factory, ok := driverRegistry[driver]
	if !ok {
		return nil, fmt.Errorf("driver %s unsupported", driver)
	}

	db, dialect, err := factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver %s: %w", driver, err)
	}

	if len(opts) == 0 {
		opts = append(opts, defaultPoolConfig(cfg))
	}

	for _, opt := range opts {
		opt(db)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err = db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	bundb := bun.NewDB(db, dialect)

	if !cfg.IsProduction {
		bundb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return bundb, nil
}

func createPostgres(cfg *DBConfig) (*sql.DB, schema.Dialect, error) {
	connector := pgdriver.NewConnector(
		pgdriver.WithDSN(cfg.DSN),
		pgdriver.WithTimeout(cfg.Timeout),
	)

	return sql.OpenDB(connector), pgdialect.New(), nil
}
func createMySQL(cfg *DBConfig) (*sql.DB, schema.Dialect, error) {
	dsn := cfg.DSN

	if !strings.Contains(dsn, "parseTime=true") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	return db, mysqldialect.New(), err
}

func createSqlite(cfg *DBConfig) (*sql.DB, schema.Dialect, error) {
	db, err := sql.Open("sqlite", cfg.DSN)
	return db, sqlitedialect.New(), err
}

func createSqlite3(cfg *DBConfig) (*sql.DB, schema.Dialect, error) {
	db, err := sql.Open("sqlite3", cfg.DSN)
	return db, sqlitedialect.New(), err
}

func defaultPoolConfig(cfg *DBConfig) DBOption {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
}
