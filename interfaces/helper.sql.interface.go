package sdk_inf

import (
	"github.com/uptrace/bun"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

type ISql interface {
	IncColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery
	ExcColumn(cols ...string) func(sq *bun.SelectQuery) *bun.SelectQuery
	Column(options *sdk_dto.ColumnOptions) func(sq *bun.SelectQuery) *bun.SelectQuery
}
