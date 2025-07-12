package messages

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/config"
)

type APIHandler struct {
	service *SystemMessageService
}

func NewApiHandler(service *SystemMessageService) *APIHandler {
	return &APIHandler{service: service}
}

func (h *APIHandler) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodPost,
			Path:         "/v1/messages",
			Handler:      h.PostMessage,
			Description:  "Post a message to the processor",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/messages/workers",
			Handler:      h.GetWorkers,
			Description:  "Get all registered workers",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/messages/health",
			Handler:      h.HealthCheck,
			Description:  "Check service health",
			AuthRequired: false,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/messages/config",
			Handler:      h.GetConfig,
			Description:  "Get service configuration",
			AuthRequired: true,
		},
	}
}

// PostMessageRequest represents the request body for posting a message
type PostMessageRequest struct {
	Type     string                 `json:"type"`      // Message type (e.g., "email", "notification")
	Payload  map[string]interface{} `json:"payload"`   // Message payload
	TenantID string                 `json:"tenant_id"` // Tenant identifier
	Priority int                    `json:"priority"`  // Message priority (0=low, 1=normal, 2=high, 3=urgent)
}

// PostMessageResponse represents the response for posting a message
type PostMessageResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"message_id,omitempty"`
}

// GetWorkersResponse represents the response for getting workers
type GetWorkersResponse struct {
	Workers []*WorkerMetadata `json:"workers"`
	Count   int               `json:"count"`
}

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	PollInterval         string `json:"poll_interval"`
	RecoveryEnabled      bool   `json:"recovery_enabled"`
	MaxProcessingAge     string `json:"max_processing_age"`
	ProcessingTimeout    string `json:"processing_timeout"`
	DefaultMaxRetries    int    `json:"default_max_retries"`
	CleanupEnabled       bool   `json:"cleanup_enabled"`
	CleanupMaxAge        string `json:"cleanup_max_age"`
	CleanupInterval      string `json:"cleanup_interval"`
	KeepCompleteMessages bool   `json:"keep_complete_messages"`
	Debug                bool   `json:"debug"`
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status  string          `json:"status"`
	Service string          `json:"service"`
	Running bool            `json:"running"`
	Workers int             `json:"workers"`
	Config  *ConfigResponse `json:"config,omitempty"`
}

// GetConfigResponse represents the configuration response
type GetConfigResponse struct {
	Config *ConfigResponse `json:"config"`
}

func (h *APIHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var request PostMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if request.Type == "" {
		api.WriteBadRequest(w, r, "Message type is required", "type field cannot be empty")
		return
	}

	if request.TenantID == "" {
		api.WriteBadRequest(w, r, "Tenant ID is required", "tenant_id field cannot be empty")
		return
	}

	if request.Payload == nil {
		request.Payload = make(map[string]interface{})
	}

	// Set default priority if not provided
	if request.Priority == 0 {
		request.Priority = 1 // Normal priority
	}

	// Post the message
	ctx := context.Background()
	messageID, err := h.service.PostMessage(ctx, request.Type, request.Payload, request.TenantID, request.Priority)
	if err != nil {
		api.WriteBadRequest(w, r, "Failed to post message", err.Error())
		return
	}

	response := PostMessageResponse{
		Success:   true,
		Message:   "Message posted successfully",
		MessageID: messageID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) GetWorkers(w http.ResponseWriter, r *http.Request) {
	workers := h.service.GetRegisteredWorkers()

	response := GetWorkersResponse{
		Workers: workers,
		Count:   len(workers),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	workers := h.service.GetRegisteredWorkers()
	config := h.getConfigResponse()

	response := HealthCheckResponse{
		Status:  "healthy",
		Service: "message-processor",
		Running: h.service.IsRunning(),
		Workers: len(workers),
		Config:  config,
	}

	if !h.service.IsRunning() {
		response.Status = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.getConfigResponse()

	response := GetConfigResponse{
		Config: config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getConfigResponse creates a config response from the service configuration
func (h *APIHandler) getConfigResponse() *ConfigResponse {
	// Get configuration from the main config service
	configInstance := config.GetInstance()
	if configInstance == nil {
		return &ConfigResponse{
			PollInterval:         "1s",
			RecoveryEnabled:      true,
			MaxProcessingAge:     "5m",
			ProcessingTimeout:    "30s",
			DefaultMaxRetries:    3,
			CleanupEnabled:       true,
			CleanupMaxAge:        "168h",
			CleanupInterval:      "1h",
			KeepCompleteMessages: false,
			Debug:                false,
		}
	}

	systemConfig := configInstance.Get()

	getString := func(key string, defaultValue string) string {
		if item := systemConfig.Get(key); item != nil && item.IsSet() {
			return item.GetString()
		}
		return defaultValue
	}

	getBool := func(key string, defaultValue bool) bool {
		if item := systemConfig.Get(key); item != nil && item.IsSet() {
			return item.GetBool()
		}
		return defaultValue
	}

	getInt := func(key string, defaultValue int) int {
		if item := systemConfig.Get(key); item != nil && item.IsSet() {
			return item.GetInt()
		}
		return defaultValue
	}

	return &ConfigResponse{
		PollInterval:         getString(config.MessageProcessorPollIntervalKey, "1s"),
		RecoveryEnabled:      getBool(config.MessageProcessorRecoveryEnabledKey, true),
		MaxProcessingAge:     getString(config.MessageProcessorMaxProcessingAgeKey, "5m"),
		ProcessingTimeout:    getString(config.MessageProcessorProcessingTimeoutKey, "30s"),
		DefaultMaxRetries:    getInt(config.MessageProcessorDefaultMaxRetriesKey, 3),
		CleanupEnabled:       getBool(config.MessageProcessorCleanupEnabledKey, true),
		CleanupMaxAge:        getString(config.MessageProcessorCleanupMaxAgeKey, "168h"),
		CleanupInterval:      getString(config.MessageProcessorCleanupIntervalKey, "1h"),
		KeepCompleteMessages: getBool(config.MessageProcessorKeepCompleteMessagesKey, false),
		Debug:                getBool(config.MessageProcessorDebugKey, false),
	}
}
