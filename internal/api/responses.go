package api

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api/models"
)

func WritePaginatedResponse[T any](w http.ResponseWriter, r *http.Request, data []T, pagination models.Pagination, totalCount int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.PaginatedResponse[T]{
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

func WriteObjectResponseWithStatus[T any](w http.ResponseWriter, r *http.Request, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func WriteSuccessResponse(w http.ResponseWriter, r *http.Request, id, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.SuccessResponse{
		ID:      id,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}
