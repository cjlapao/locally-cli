package filters

import (
	"math"
	"strconv"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"gorm.io/gorm"
)

// PaginationFilter represents pagination parameters for URL query parsing
type PaginationFilter struct {
	Page     int   `json:"page" yaml:"page"`
	PageSize int   `json:"page_size" yaml:"page_size"`
	Total    int64 `json:"total" yaml:"total"`
}

// NewPaginationFilter creates a new PaginationFilter instance and parses the raw string
func NewPaginationFilter(raw string) *PaginationFilter {
	// Use default page size, fallback to constant if config is not available
	defaultPageSize := config.DefaultPageSizeInt
	if configInstance := config.GetInstance(); configInstance != nil {
		if cfg := configInstance.Get(); cfg != nil {
			defaultPageSize = cfg.GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt)
		}
	}
	
	pf := &PaginationFilter{
		Page:     1,               // Default to page 1
		PageSize: defaultPageSize, // Default page size from config or fallback
	}
	if raw != "" {
		pf.Parse(raw)
	}
	return pf
}

// Parse parses pagination parameters from a raw string
// If a full query string is provided (e.g. "filter=active&page=1&page_size=10&order_by=name asc"),
// it will extract only the pagination parameters and ignore the rest.
func (pf *PaginationFilter) Parse(raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}

	// Check if this looks like a full query string with multiple parameters
	if strings.Contains(trimmed, "&") || strings.Contains(trimmed, "?") {
		// Extract pagination parameters from query string
		pf.extractPaginationFromQuery(trimmed)
	} else {
		// Assume it's just pagination parameters (e.g. "page=1&page_size=10")
		pf.extractPaginationFromQuery(trimmed)
	}
}

// extractPaginationFromQuery extracts pagination parameters from a query string
func (pf *PaginationFilter) extractPaginationFromQuery(query string) {
	// Remove leading ? if present
	query = strings.TrimPrefix(query, "?")

	// Split by & to get individual parameters
	params := strings.Split(query, "&")

	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}

		// Split by = to get key-value pairs
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse page parameter
		if strings.EqualFold(key, "page") {
			if page, err := strconv.Atoi(value); err == nil && page > 0 {
				pf.Page = page
			}
		}

		// Parse page_size parameter (also supports pageSize, limit, per_page)
		if strings.EqualFold(key, "page_size") || strings.EqualFold(key, "pagesize") ||
			strings.EqualFold(key, "limit") || strings.EqualFold(key, "per_page") ||
			strings.EqualFold(key, "perpage") {
			if pageSize, err := strconv.Atoi(value); err == nil && pageSize > 0 {
				pf.PageSize = pageSize
			}
		}
	}
}

// GetPage returns the current page number
func (pf *PaginationFilter) GetPage() int {
	return pf.Page
}

// GetPageSize returns the page size
func (pf *PaginationFilter) GetPageSize() int {
	return pf.PageSize
}

// GetTotal returns the total count
func (pf *PaginationFilter) GetTotal() int64 {
	return pf.Total
}

// SetTotal sets the total count (useful after query execution)
func (pf *PaginationFilter) SetTotal(total int64) {
	pf.Total = total
}

// GetTotalPages returns the total number of pages
func (pf *PaginationFilter) GetTotalPages() int {
	if pf.Total == 0 || pf.PageSize == 0 {
		return 0
	}
	return int(math.Ceil(float64(pf.Total) / float64(pf.PageSize)))
}

// GetOffset returns the offset for database queries
func (pf *PaginationFilter) GetOffset() int {
	offset := pf.GetPageIndex() * pf.PageSize
	
	// Only adjust for total if we have a total count and we're beyond it
	if pf.Total > 0 && offset > int(pf.Total) {
		pf.Page = 1
		offset = int(pf.Total) - pf.PageSize
		if offset < 0 {
			offset = 0
		}
	}
	return offset
}

// GetPageIndex returns the zero-based page index
func (pf *PaginationFilter) GetPageIndex() int {
	return pf.Page - 1
}

// IsValid returns true if the pagination parameters are valid
func (pf *PaginationFilter) IsValid() bool {
	return pf.Page > 0 && pf.PageSize > 0
}

// Apply applies pagination to a gorm DB and returns the updated DB
func (pf *PaginationFilter) Apply(db *gorm.DB) *gorm.DB {
	if db == nil || !pf.IsValid() {
		return db
	}

	// Calculate offset directly when total is not known
	offset := pf.GetPageIndex() * pf.PageSize
	return db.Offset(offset).Limit(pf.PageSize)
}
