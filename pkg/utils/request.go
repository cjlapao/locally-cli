package utils

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
)

func GetPaginationFromRequest(r *http.Request) (int, int) {
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")

	if page == "" {
		page = "1"
	}

	if pageSize == "" {
		pageSize = "20"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, 0
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		return 0, 0
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
		pageSizeInt = cfg.GetInt(config.PaginationDefaultPageSizeKey, 20)
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
