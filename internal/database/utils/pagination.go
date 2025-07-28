package utils

import (
	"math"

	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/logging"
	"gorm.io/gorm"
)

// PaginatedQuery executes a paginated query and returns a FilterResponse
// This is a generic helper function that can be used by any data store
// to avoid repeating pagination logic
func PaginatedQuery[T any](
	db *gorm.DB,
	tenantID string,
	filterObj *filters.Filter,
	model T,
) (*filters.FilterResponse[T], error) {
	var items []T
	total := int64(0)
	pageIndex := 0
	pageSize := 0

	// Get total count
	countQuery := db.Model(&model)
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get items with pagination, if no page size is provided, return all items
	if filterObj == nil {
		query := db
		if tenantID != "" {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.Find(&items).Error; err != nil {
			return nil, err
		}
		pageSize = int(total)
	} else {
		filterString, args := filterObj.Generate()
		pageIndex = filterObj.Page - 1
		pageSize = filterObj.PageSize

		offset := pageIndex * pageSize

		// checking if the page request is higher total records pages
		if offset > int(total) {
			logging.Warn("page_request_is_higher_than_total_records_pages", map[string]interface{}{
				"page_request":  filterObj.Page,
				"total_records": total,
			})
			pageIndex = 0
			offset = int(total) - pageSize
		}

		query := db
		if tenantID != "" {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.Where(filterString, args...).
			Offset(offset).
			Limit(pageSize).
			Find(&items).Error; err != nil {
			return nil, err
		}
	}

	response := filters.FilterResponse[T]{
		Items:      items,
		Total:      total,
		Page:       pageIndex + 1,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}

	return &response, nil
}

// PaginatedQueryWithPreload executes a paginated query with preloads and returns a FilterResponse
// This variant allows you to specify preloads for related data
func PaginatedQueryWithPreload[T any](
	db *gorm.DB,
	tenantID string,
	filterObj *filters.Filter,
	model T,
	preloads ...string,
) (*filters.FilterResponse[T], error) {
	var items []T
	total := int64(0)
	pageIndex := 0
	pageSize := 0

	// Get total count
	countQuery := db.Model(&model)
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Build query with preloads
	query := db
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	// Get items with pagination, if no page size is provided, return all items
	if filterObj == nil {
		query := db
		if tenantID != "" {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.Find(&items).Error; err != nil {
			return nil, err
		}
	} else {
		filterString, args := filterObj.Generate()
		pageIndex = filterObj.Page - 1
		pageSize = filterObj.PageSize
		offset := pageIndex * pageSize

		// checking if the page request is higher total records pages
		if offset > int(total) {
			logging.Warn("page_request_is_higher_than_total_records_pages", map[string]interface{}{
				"page_request":  filterObj.Page,
				"total_records": total,
			})
			pageIndex = 0
			offset = int(total) - pageSize
		}

		if tenantID != "" {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.Where(filterString, args...).Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
			return nil, err
		}
	}

	response := filters.FilterResponse[T]{
		Items:      items,
		Total:      total,
		Page:       pageIndex + 1,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}

	return &response, nil
}
