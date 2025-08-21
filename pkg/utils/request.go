package utils

import (
	"net/http"
	"strconv"
	"strings"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func HasPaginationRequest(r *http.Request) bool {
	page := r.URL.Query().Get("page")
	return page != ""
}

func GetRequestContextFromRequest(r *http.Request) *models.RequestContext {
	if r == nil {
		return &models.RequestContext{}
	}

	ctx := appctx.FromContext(r.Context())
	if ctx == nil {
		return &models.RequestContext{}
	}

	result := &models.RequestContext{}
	result.Pagination = ParseQueryRequest(r)
	tenantID := ctx.GetTenantID()
	result.TenantID = tenantID
	userID := ctx.GetUserID()
	result.UserID = userID
	username := ctx.GetUsername()
	result.Username = username
	result.RequestID = ctx.GetRequestID()
	return result
}

// ParseQueryRequest parses HTTP request query parameters and returns a PaginationRequest
// Supports query parameters like: ?page=1&page_size=20&filter=name=john&order_by=created_at desc
func ParseQueryRequest(r *http.Request) *api_models.PaginationRequest {
	if r == nil {
		return &api_models.PaginationRequest{}
	}

	cfg := config.GetInstance()
	defaultPageSize := config.DefaultPageSizeInt
	if cfg != nil && cfg.Get() != nil {
		defaultPageSize = cfg.Get().GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt)
	}

	query := r.URL.Query()

	// Parse page
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	// Parse page_size
	pageSize := defaultPageSize
	pageSizeKeys := []string{"page_size", "pageSize", "per_page", "perPage", "limit"}
	for _, key := range pageSizeKeys {
		if pageSizeStr := query.Get(key); pageSizeStr != "" {
			if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
				pageSize = pageSizeInt
				break
			}
		}
	}

	// Parse filter
	filter := ""
	filterKeys := []string{"filter", "filters", "where"}
	for _, key := range filterKeys {
		if filterStr := query.Get(key); filterStr != "" {
			filter = strings.Trim(filterStr, "\"'")
			break
		}
	}

	// Parse ordering - support multiple formats
	orderBy := ""

	// First check for combined ordering parameters
	combinedOrderKeys := []string{"order_by", "orderBy"}
	for _, key := range combinedOrderKeys {
		if orderStr := query.Get(key); orderStr != "" {
			orderBy = orderStr
			break
		}
	}

	// If no combined ordering found, check for separate sort and order parameters
	if orderBy == "" {
		sort := query.Get("sort")
		order := query.Get("order")
		if sort != "" {
			if order != "" {
				orderBy = sort + " " + order
			} else {
				orderBy = sort
			}
		}
	}

	return &api_models.PaginationRequest{
		Page:     page,
		PageSize: pageSize,
		Filter:   filter,
		Sort:     orderBy, // Using Sort field for order_by data
		Order:    "",      // Keep Order empty as we combine it into Sort
	}
}

// ParseQueryToQueryBuilder directly parses HTTP request query parameters and returns a QueryBuilder
// This is a convenience function that combines ParseQueryRequest with ToQueryBuilder
func ParseQueryToQueryBuilder(r *http.Request) *filters.QueryBuilder {
	req := ParseQueryRequest(r)
	return req.ToQueryBuilder()
}

func NewActivityFromContext(ctx *appctx.AppContext) *models.Activity {
	result := &models.Activity{}
	if ctx == nil {
		return result
	}
	result.TenantID = ctx.GetTenantID()
	result.ActorID = ctx.GetUserID()
	result.ActorName = ctx.GetUsername()
	result.RequestID = ctx.GetRequestID()
	result.CorrelationID = ctx.GetCorrelationID()
	result.UserAgent = ctx.GetUserAgent()
	result.ActorIP = ctx.GetUserIP()

	if result.TenantID == "" {
		result.TenantID = config.UnknownTenantID
	}

	if result.ActorID == "" {
		result.ActorID = config.UnknownUserID
	}

	if result.ActorName == "" {
		result.ActorName = "unknown"
	}

	if result.RequestID == "" {
		result.RequestID = "unknown"
	}

	if result.CorrelationID == "" {
		result.CorrelationID = "unknown"
	}

	if result.UserAgent == "" {
		result.UserAgent = "unknown"
	}

	if result.ActorIP == "" {
		result.ActorIP = "unknown"
	}

	return result
}
