package api

type PaginatedResponse[T any] struct {
	TotalCount int        `json:"total_count"`
	Pagination Pagination `json:"pagination"`
	Data       []T        `json:"data"`
}

type Pagination struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
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
