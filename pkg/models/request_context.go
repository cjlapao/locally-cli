package models

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/database/filters"
)

type RequestContext struct {
	Filter     *filters.Filter
	Pagination *api_models.Pagination
	TenantID   string
	UserID     string
	Username   string
	RequestID  string
}
