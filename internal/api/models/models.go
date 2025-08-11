// Package models contains the models for the API service.
package models

// Handler represents the main API handler
// This can be used for general API functionality that doesn't belong to specific domains
type Handler struct{}

type PaginatedResponse[T any] struct {
	TotalCount int64      `json:"total_count"`
	Pagination Pagination `json:"pagination"`
	Data       []T        `json:"data"`
}

type Pagination struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type PaginationRequest struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Filter   string `json:"filter,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"` // asc or desc
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
