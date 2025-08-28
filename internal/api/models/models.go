// Package models contains the models for the API service.
package models

import (
	"fmt"
	"strconv"

	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// Handler represents the main API handler
// This can be used for general API functionality that doesn't belong to specific domains
type Handler struct{}

// APIError represents a standardized API error response
type APIError struct {
	Error     ErrorDetails `json:"error"`
	Timestamp string       `json:"timestamp"`
	Path      string       `json:"path,omitempty"`
}

type ErrorDetailsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorDetails contains the specific error information
type ErrorDetails struct {
	Code        string                   `json:"code"`
	Message     string                   `json:"message"`
	Details     string                   `json:"details,omitempty"`
	Errors      []ErrorDetailsError      `json:"errors,omitempty"`
	Diagnostics *diagnostics.Diagnostics `json:"diagnostics,omitempty"`
}

type PaginationRequest struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Filter   string `json:"filter,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
}

// ToQueryBuilder converts a PaginationRequest to a QueryBuilder
// This allows seamless integration between HTTP request parsing and database queries
func (pr *PaginationRequest) ToQueryBuilder() *filters.QueryBuilder {
	if pr == nil {
		return filters.NewQueryBuilder("")
	}

	// Build the query string from the request components
	queryParts := make([]string, 0, 4)

	// Add filter parameter
	if pr.Filter != "" {
		queryParts = append(queryParts, fmt.Sprintf("filter=%s", pr.Filter))
	}

	// Add ordering parameter
	if pr.Sort != "" {
		queryParts = append(queryParts, fmt.Sprintf("order_by=%s", pr.Sort))
	} else if pr.Order != "" {
		// Fallback to Order field if Sort is empty
		queryParts = append(queryParts, fmt.Sprintf("order_by=%s", pr.Order))
	}

	// Add pagination parameters
	if pr.Page > 0 {
		queryParts = append(queryParts, fmt.Sprintf("page=%s", strconv.Itoa(pr.Page)))
	}
	if pr.PageSize > 0 {
		queryParts = append(queryParts, fmt.Sprintf("page_size=%s", strconv.Itoa(pr.PageSize)))
	}

	// Combine into query string format
	var queryString string
	if len(queryParts) > 0 {
		queryString = "?" + joinStrings(queryParts, "&")
	}

	return filters.NewQueryBuilder(queryString)
}

// ToQueryString converts a PaginationRequest to a URL query string
// Useful for generating URLs or debugging
func (pr *PaginationRequest) ToQueryString() string {
	if pr == nil {
		return ""
	}

	queryParts := make([]string, 0, 5)

	if pr.Page > 0 {
		queryParts = append(queryParts, fmt.Sprintf("page=%d", pr.Page))
	}
	if pr.PageSize > 0 {
		queryParts = append(queryParts, fmt.Sprintf("page_size=%d", pr.PageSize))
	}
	if pr.Filter != "" {
		queryParts = append(queryParts, fmt.Sprintf("filter=%s", pr.Filter))
	}
	if pr.Sort != "" {
		queryParts = append(queryParts, fmt.Sprintf("sort=%s", pr.Sort))
	}
	if pr.Order != "" {
		queryParts = append(queryParts, fmt.Sprintf("order=%s", pr.Order))
	}

	if len(queryParts) > 0 {
		return "?" + joinStrings(queryParts, "&")
	}
	return ""
}

// helper function to join strings
func joinStrings(parts []string, separator string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += separator + parts[i]
	}
	return result
}

type PaginationResponse[T any] struct {
	TotalCount int64      `json:"total_count"`
	Pagination Pagination `json:"pagination"`
	Data       []T        `json:"data"`
}

type Pagination struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type StatusResponse struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

type ObjectResponse[T any] struct {
	ID   string `json:"id,omitempty"`
	Data T      `json:"data,omitempty"`
}

type SuccessResponse struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}
