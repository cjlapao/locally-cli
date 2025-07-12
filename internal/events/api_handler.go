package events

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/google/uuid"
)

// APIHandler provides REST API endpoints for events
type APIHandler struct {
	eventService *EventService
	sseHandler   *SseService
}

// NewApiHandler creates a new API handler for events
func NewApiHandler(eventService *EventService, authService *auth.AuthService) *APIHandler {
	return &APIHandler{
		eventService: eventService,
		sseHandler:   newSSEService(eventService, authService),
	}
}

// Routes implements the RouteRegistrar interface
func (h *APIHandler) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/events/stream",
			Handler:      h.sseHandler.HandleSSE,
			Description:  "Server-Sent Events stream for real-time updates",
			AuthRequired: true,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/events/push",
			Handler:      h.HandlePushEvent,
			Description:  "Push an event to the event service",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/events/stats",
			Handler:      h.HandleGetStats,
			Description:  "Get event service statistics",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/events/health",
			Handler:      h.HandleHealthCheck,
			Description:  "Health check for event service",
			AuthRequired: false,
		},
	}
}

// PushEventRequest represents a push event request
type PushEventRequest struct {
	Type     string                 `json:"type"`
	Message  string                 `json:"message"`
	TenantID string                 `json:"tenant_id"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// StatsResponse represents event service statistics
type StatsResponse struct {
	ConnectedClients     int            `json:"connected_clients"`
	TenantID             string         `json:"tenant_id"`
	AllTenantConnections map[string]int `json:"all_tenant_connections,omitempty"`
	ServiceStatus        string         `json:"service_status"`
	Timestamp            string         `json:"timestamp"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// HandlePushEvent handles push event requests
func (h *APIHandler) HandlePushEvent(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (added by auth middleware)
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		api.WriteUnauthorized(w, r, "No authentication claims found", "")
		return
	}

	// Parse request body
	var req PushEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+err.Error())
		return
	}

	// Validate request
	if req.Type == "" {
		api.WriteBadRequest(w, r, "Invalid request", "Event type is required")
		return
	}
	if req.Message == "" {
		api.WriteBadRequest(w, r, "Invalid request", "Message is required")
		return
	}

	// Create event data
	eventData := map[string]interface{}{
		"message": req.Message,
		"source":  "api_test",
		"user":    claims.Username,
	}

	if req.Data != nil {
		for k, v := range req.Data {
			eventData[k] = v
		}
	}

	tenantID := claims.TenantID
	if req.TenantID != "" {
		tenantID = req.TenantID
	}

	if tenantID == "" || tenantID == "global" || tenantID == "default" {
		tenantID = uuid.Nil.String()
	}

	event := &Event{
		ID:        uuid.New().String(),
		Type:      EventType(req.Type),
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data:      eventData,
	}

	h.eventService.PublishEvent(event)

	// Return success response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Event published successfully",
		"event_id":  event.ID,
		"tenant_id": tenantID,
		"type":      event.Type,
		"timestamp": event.Timestamp.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

// HandleGetStats handles requests for event service statistics
func (h *APIHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// Get claims from context (added by auth middleware)
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		api.WriteUnauthorized(w, r, "No authentication claims found", "")
		return
	}

	// Get statistics for the user's tenant
	connectedClients := h.eventService.GetConnectedClients(claims.TenantID)

	response := StatsResponse{
		ConnectedClients: connectedClients,
		TenantID:         claims.TenantID,
		ServiceStatus:    "active",
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	// If user has admin role, include all tenant statistics
	if claims.Role == "admin" {
		response.AllTenantConnections = h.eventService.GetAllConnectedClients()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

// HandleHealthCheck handles health check requests
func (h *APIHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0", // You can make this configurable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
