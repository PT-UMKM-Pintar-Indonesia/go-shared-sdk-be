package sdk_helper

import (
	"math"

	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/v1/outputs"
)

func Pagination(limit, offset, total int) *sdk_opt.Pagination {
	res := new(sdk_opt.Pagination)

	res.Limit = limit
	res.Page = offset
	res.TotalPage = math.Ceil(float64(total) / float64(limit))
	res.TotalData = total

	return res
}
