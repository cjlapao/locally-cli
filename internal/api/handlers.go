package api

import "net/http"

// Handler represents the main API handler
// This can be used for general API functionality that doesn't belong to specific domains
type Handler struct{}

// NewHandler creates a new API handler
func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Path:    "/v1/health",
			Handler: h.HealthCheck,
		},
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
