package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapPaginationToDto(pagination *filters.Pagination) *models.Pagination {
	return &models.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}
}
