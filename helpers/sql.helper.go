package sdk_helper

import (
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/uptrace/bun"
)

type sql struct{}

func NewSql() sdk_inf.ISql {
	return &sql{}
}

func (h *sql) IncColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {
		if len(cols) == 0 {
			return sq
		}
		return sq.Column(cols...)
	}
}

func (h *sql) ExcColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {
		if len(cols) == 0 {
			return sq
		}
		return sq.ExcludeColumn(cols...)
	}
}

func (h *sql) Column(opt *sdk_dto.ColumnOptions) func(sq *bun.SelectQuery) *bun.SelectQuery {
	return func(sq *bun.SelectQuery) *bun.SelectQuery {
		if opt == nil {
			return sq
		}

		if len(opt.Inc) > 0 {
			sq.Column(opt.Inc...)
		}

		if len(opt.Exc) > 0 {
			sq.ExcludeColumn(opt.Exc...)
		}

		return sq
	}
}
