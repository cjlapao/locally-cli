package events

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	sseService *SseService
	sseOnce    sync.Once
)

type SseService struct {
	eventService *EventService
	authService  *auth.AuthService
}

func InitializeSseService(eventService *EventService, authService *auth.AuthService) *SseService {
	sseOnce.Do(func() {
		sseService = newSSEService(eventService, authService)
	})
	return sseService
}

func GetSseServiceInstance() *SseService {
	if sseService == nil {
		logging.Fatal("SseService not initialized")
	}

	return sseService
}

func newSSEService(eventService *EventService, authService *auth.AuthService) *SseService {
	return &SseService{
		eventService: eventService,
		authService:  authService,
	}
}

func (h *SseService) validateJWTFromRequest(r *http.Request) (*auth.AuthClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return claims, nil
}

// isGlobalTenant checks if the tenant ID represents a global tenant (empty or all zeros)
func (h *SseService) isGlobalTenant(tenantID string) bool {
	if tenantID == "" || tenantID == config.GlobalTenantID {
		return true
	}

	// Check if it's all zeros (empty UUID)
	zeroUUID := uuid.Nil.String()
	return tenantID == zeroUUID
}

func (h *SseService) sendHeartbeat(w http.ResponseWriter, clientID string, tenantID string) {
	heartbeatEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventTypeSystemAlert,
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type":      "heartbeat",
			"client_id": clientID,
			"message":   "Connection heartbeat",
		},
	}

	fmt.Fprint(w, heartbeatEvent.ToSSEFormat())
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *SseService) HandleSSE(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}
	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
	}).Info("SSE connection attempt")

	claims, err := h.validateJWTFromRequest(r)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"error":     err,
		}).Error("SSE authentication failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"username":  claims.Username,
		"tenant_id": claims.TenantID,
	}).Info("SSE authentication successful")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")

	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
		}).Error("ResponseWriter does not support flushing")
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	if _, err := fmt.Fprint(w, ": SSE connection established\n\n"); err != nil {
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"error":     err,
		}).Error("Failed to write initial SSE comment")
		return
	}
	flusher.Flush()
	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
	}).Info("SSE initial comment sent and flushed")

	clientID := uuid.New().String()

	// Determine if this is a global tenant
	isGlobal := h.isGlobalTenant(claims.TenantID)

	// Create the main client for tenant-specific messages
	client := &Client{
		ID:          clientID,
		TenantID:    claims.TenantID,
		Username:    claims.Username,
		Channel:     make(chan *Event, 50),
		ConnectedAt: time.Now(),
	}

	// Create a global client for global messages (if not already global)
	var globalClient *Client
	if !isGlobal {
		globalClient = &Client{
			ID:          clientID + "-global",
			TenantID:    uuid.Nil.String(), // Use nil UUID for global messages
			Username:    claims.Username,
			Channel:     client.Channel, // Use the same channel as main client
			ConnectedAt: time.Now(),
		}
	}

	welcomeEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventTypeConnectionEstablished,
		TenantID:  claims.TenantID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id": clientID,
			"connected": true,
			"message":   "SSE connection established",
			"username":  claims.Username,
			"tenant_id": claims.TenantID,
			"is_global": isGlobal,
		},
	}

	if _, err := fmt.Fprint(w, welcomeEvent.ToSSEFormat()); err != nil {
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"error":     err,
		}).Error("Failed to write welcome event")
		return
	}
	flusher.Flush()
	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"client_id": clientID,
	}).Info("SSE welcome event sent to client")

	// Register the main client
	h.eventService.RegisterClient(client)
	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"client_id": clientID,
		"username":  claims.Username,
		"tenant_id": claims.TenantID,
	}).Info("SSE client registered")

	// Register global client if needed
	if globalClient != nil {
		h.eventService.RegisterClient(globalClient)
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"client_id": globalClient.ID,
			"username":  claims.Username,
		}).Info("SSE global client registered")
	}

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	defer func() {
		// Unregister both clients
		h.eventService.UnregisterClient(clientID)
		if globalClient != nil {
			h.eventService.UnregisterClient(globalClient.ID)
		}
		logging.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"client_id": clientID,
			"username":  claims.Username,
		}).Info("SSE client disconnected")
	}()

	clientGone := r.Context().Done()

	logging.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"client_id": clientID,
	}).Info("SSE entering event loop for client")

	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				logging.WithFields(logrus.Fields{
					"client_ip": clientIP,
					"client_id": clientID,
				}).Info("SSE client channel closed")
				return
			}

			if _, err := fmt.Fprint(w, event.ToSSEFormat()); err != nil {
				logging.WithFields(logrus.Fields{
					"client_ip": clientIP,
					"client_id": clientID,
					"error":     err,
				}).Error("Failed to write event to client")
				return
			}
			flusher.Flush()
			logging.WithFields(logrus.Fields{
				"client_ip":  clientIP,
				"client_id":  clientID,
				"event_type": event.Type,
			}).Info("SSE event sent to client")

		case <-heartbeatTicker.C:
			h.sendHeartbeat(w, clientID, claims.TenantID)

		case <-clientGone:
			logging.WithFields(logrus.Fields{
				"client_ip": clientIP,
				"client_id": clientID,
			}).Info("SSE client disconnected")
			return
		}
	}
}
