package sdk_con

import (
	"context"
	"database/sql"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/dtos"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func SqlConnection(ctx context.Context, req sdk_dto.Request[sdk_dto.Environtment]) (*bun.DB, error) {
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(req.Config.POSTGRES.URL)))

	if err := db.Ping(); err != nil {
		logrus.Error(err)
		return nil, err
	}

	if db != nil {
		logrus.Info("Database connection success")

		db.SetConnMaxIdleTime(time.Duration(time.Second * time.Duration(30)))
		db.SetConnMaxLifetime(time.Duration(time.Second * time.Duration(30)))
	}

	bundb := bun.NewDB(db, pgdialect.New())

	if req.Config.APP.ENV != sdk_cons.PROD {
		bundb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(true), bundebug.WithVerbose(true), bundebug.FromEnv("BUNDEBUG")))
	}

	return bundb, nil
}
