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

func GetPaginationFromRequest(r *http.Request) (int, int) {
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")
	cfg := config.GetInstance().Get()
	defaultPageSize := cfg.GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt)

	if page == "" {
		page = "1"
	}

	if pageSize == "" {
		pageSize = strconv.Itoa(defaultPageSize)
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 1, defaultPageSize
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return 1, defaultPageSize
	}

	return pageInt, pageSizeInt
}

func HasPaginationRequest(r *http.Request) bool {
	page := r.URL.Query().Get("page")
	return page != ""
}

func GetFilterFromRequest(r *http.Request) (*filters.Filter, error) {
	cfg := config.GetInstance().Get()
	urlFilter := r.URL.Query().Get("filter")
	urlFilter = strings.Trim(urlFilter, "\"")
	urlPage := r.URL.Query().Get("page")
	var err error
	pageInt, err := strconv.Atoi(urlPage)
	if err != nil {
		pageInt = -1
	}
	urlPageSize := r.URL.Query().Get("page_size")
	pageSizeInt, err := strconv.Atoi(urlPageSize)
	if err != nil {
		pageSizeInt = cfg.GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt)
	}
	var dbFilter *filters.Filter
	if urlFilter != "" {
		dbFilter, err = filters.Parse(urlFilter)
		if err != nil {
			return nil, err
		}
		if urlPage != "" {
			dbFilter.Page = pageInt
		}
		if urlPageSize != "" {
			dbFilter.PageSize = pageSizeInt
		}
	}
	if pageSizeInt >= 0 && dbFilter == nil {
		if pageInt == -1 {
			pageInt = 1
		}
		dbFilter = filters.NewFilter().WithPage(pageInt).WithPageSize(pageSizeInt)
	}

	return dbFilter, nil
}

func HasFilterRequest(r *http.Request) bool {
	urlFilter := r.URL.Query().Get("filter")
	return urlFilter != ""
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
	filter, _ := GetFilterFromRequest(r)
	result.Filter = filter
	page, pageSize := GetPaginationFromRequest(r)
	result.Pagination = &api_models.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	tenantID := ctx.GetTenantID()
	result.TenantID = tenantID
	userID := ctx.GetUserID()
	result.UserID = userID
	username := ctx.GetUsername()
	result.Username = username
	result.RequestID = ctx.GetRequestID()
	return result
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
