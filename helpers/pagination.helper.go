package sdk_helper

import (
	"errors"
	"math"

	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func Pagination(limit, offset, total int) (*sdk_opt.Pagination, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	if offset < 0 {
		return nil, errors.New("offset must be non-negative")
	}

	if total < 0 {
		return nil, errors.New("total must be non-negative")
	}

	res := &sdk_opt.Pagination{
		Limit:     limit,
		Page:      offset,
		TotalData: total,
	}

	if limit > 0 {
		res.TotalPage = math.Ceil(float64(total) / float64(limit))
	} else {
		res.TotalPage = 0
	}

	return res, nil
}

func CalculateOffset(page, pageSize int) (int, error) {
	if page <= 0 {
		return 0, errors.New("page must be greater than 0")
	}

	if pageSize <= 0 {
		return 0, errors.New("pageSize must be greater than 0")
	}

	return (page - 1) * pageSize, nil
}

func CalculatePage(offset, pageSize int) (int, error) {
	if offset < 0 {
		return 0, errors.New("offset must be non-negative")
	}

	if pageSize <= 0 {
		return 0, errors.New("pageSize must be greater than 0")
	}

	return offset/pageSize + 1, nil
}
