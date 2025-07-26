package api

import (
	"encoding/json"
	"net/http"
)

func WritePaginatedResponse[T any](w http.ResponseWriter, r *http.Request, data []T, pagination Pagination, totalCount int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := PaginatedResponse[T]{
		Data:       data,
		Pagination: pagination,
		TotalCount: totalCount,
	}
	json.NewEncoder(w).Encode(response)
}

func WriteObjectResponse[T any](w http.ResponseWriter, r *http.Request, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)
}

func WriteSuccessResponse(w http.ResponseWriter, r *http.Request, id, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		ID:      id,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}
