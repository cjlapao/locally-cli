package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapPaginationToEntity(pagination *models.Pagination) *filters.Pagination {
	return &filters.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}
}
