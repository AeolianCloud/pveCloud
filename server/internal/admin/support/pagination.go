package support

import (
	"math"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
)

func NormalizePage(page int, perPage int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return page, perPage
}

func PageResponse[T any](items []T, total int64, page int, perPage int) admindto.PageResponse[T] {
	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	}
	return admindto.PageResponse[T]{
		List:     items,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
		LastPage: lastPage,
	}
}
