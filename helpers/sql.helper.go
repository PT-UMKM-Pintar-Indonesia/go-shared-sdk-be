package sdk_helper

import (
	"github.com/uptrace/bun"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

type sql struct{}

func NewSql() sdk_inf.ISql {
	return &sql{}
}

func (h *sql) IncColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {
		sq.Column(cols...)
		return sq
	}
}

func (h *sql) ExcColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {
		sq.ExcludeColumn(cols...)
		return sq
	}
}

func (h *sql) Column(options *sdk_dto.ColumnOptions) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {

		if len(options.Inc) > 0 {
			sq.Column(options.Inc...)
		}

		if len(options.Exc) > 0 {
			sq.ExcludeColumn(options.Exc...)
		}

		return sq
	}
}
