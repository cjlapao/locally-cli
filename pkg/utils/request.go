package utils

import (
	"net/http"
	"strconv"
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
