package models

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
)

type RequestContext struct {
	Pagination *api_models.PaginationRequest
	TenantID   string
	UserID     string
	Username   string
	RequestID  string
}
