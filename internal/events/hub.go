// Package events provides a service for publishing and subscribing to events.
package events

import (
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/sirupsen/logrus"
)

type EventService struct {
	hub            *Hub
	running        bool
	mu             sync.RWMutex
	stats          *ConnectionStats
	closedChannels map[chan *Event]bool // Track closed channels to avoid double-closing
	closedMu       sync.RWMutex
}

// ConnectionStats tracks connection statistics
type ConnectionStats struct {
	TotalConnections    int64
	ActiveConnections   int
	ConnectionsByTenant map[string]int
	mu                  sync.RWMutex
}

// Global singleton instance
var (
	globalService *EventService
	once          sync.Once
)

// NewConnectionStats creates new connection statistics tracker
func NewConnectionStats() *ConnectionStats {
	return &ConnectionStats{
		ConnectionsByTenant: make(map[string]int),
	}
}

// NewService creates a new event service
func newService() *EventService {
	return &EventService{
		hub:            NewHub(),
		stats:          NewConnectionStats(),
		closedChannels: make(map[chan *Event]bool),
	}
}

// Initialize sets up the global singleton service
func Initialize() *EventService {
	once.Do(func() {
		globalService = newService()
		logging.Info("Global event service initialized")
	})
	return globalService
}

// GetInstance returns the global singleton service instance
func GetInstance() *EventService {
	if globalService == nil {
		logging.Warn("Global event service not initialized, creating new instance")
		return Initialize()
	}
	return globalService
}

// PublishSystemError publishes a system error event
func PublishSystemError(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemError, message, data)
}

// PublishSystemInfo publishes a system info event
func PublishSystemInfo(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemInfo, message, data)
}

// PublishSystemWarning publishes a system warning event
func PublishSystemWarning(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemWarning, message, data)
}

// PublishSystemDebug publishes a system debug event
func PublishSystemDebug(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemDebug, message, data)
}

// PublishSystemTrace publishes a system trace event
func PublishSystemTrace(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemTrace, message, data)
}

// PublishSystemFatal publishes a system fatal event
func PublishSystemFatal(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemFatal, message, data)
}

// PublishSystemPanic publishes a system panic event
func PublishSystemPanic(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemPanic, message, data)
}

// PublishSystemRecover publishes a system recover event
func PublishSystemRecover(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemRecover, message, data)
}

// PublishSystemAlert publishes a system alert event
func PublishSystemAlert(tenantID string, message string, data map[string]interface{}) {
	PublishEvent(tenantID, EventTypeSystemAlert, message, data)
}

// PublishEvent publishes a generic event to the global service
func PublishEvent(tenantID string, eventType EventType, message string, data map[string]interface{}) {
	service := GetInstance()
	if service == nil {
		logging.WithField("event_type", eventType).Warn("Cannot publish event - no global service available")
		return
	}

	// Ensure data map exists
	if data == nil {
		data = make(map[string]interface{})
	}

	// Add message to data if not already present
	if message != "" {
		data["message"] = message
	}

	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data:      data,
	}

	service.PublishEvent(event)
	logging.WithFields(logrus.Fields{
		"event_type": eventType,
		"tenant_id":  tenantID,
	}).Info("Published global event")
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Start begins the hub's background processing
func (s *EventService) Start(ctx api.ApiContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // Already running
	}

	s.running = true
	go s.runHub(ctx)
	logging.Info("Event service started successfully")
	return nil
}

// Stop shuts down the hub
func (s *EventService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil // Already stopped
	}

	s.running = false

	// Close all client channels gracefully
	clientCount := 0
	for _, client := range s.hub.clients {
		s.safeCloseChannel(client.Channel)
		clientCount++
	}

	logging.WithField("client_count", clientCount).Info("Event service stopped. Disconnected clients gracefully")
	return nil
}

// safeCloseChannel safely closes a channel, preventing double-closing
func (s *EventService) safeCloseChannel(ch chan *Event) {
	s.closedMu.Lock()
	defer s.closedMu.Unlock()

	if !s.closedChannels[ch] {
		close(ch)
		s.closedChannels[ch] = true
	}
}

// RegisterClient adds a new SSE client
func (s *EventService) RegisterClient(client *Client) {
	logging.WithField("client_id", client.ID).Info("Attempting to register SSE client")

	select {
	case s.hub.register <- client:
		logging.WithField("client_id", client.ID).Info("SSE client queued for registration")
	default:
		logging.WithField("client_id", client.ID).Warn("Failed to register SSE client - hub not ready or channel full")
		// Don't close the channel here as it might be shared with another client
	}
}

// UnregisterClient removes an SSE client
func (s *EventService) UnregisterClient(clientID string) {
	s.hub.unregister <- &Client{ID: clientID}
}

// PublishEvent sends an event to all relevant clients
func (s *EventService) PublishEvent(event *Event) {
	if event == nil {
		logging.Warn("Attempted to publish nil event")
		return
	}

	logging.WithFields(logrus.Fields{
		"event_type": event.Type,
		"tenant_id":  event.TenantID,
	}).Info("Publishing event")
	s.hub.broadcast <- event
}

// GetConnectedClients returns the number of connected clients for a tenant
func (s *EventService) GetConnectedClients(tenantID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, client := range s.hub.clients {
		if client.TenantID == tenantID {
			count++
		}
	}
	return count
}

// GetAllConnectedClients returns connection counts by tenant
func (s *EventService) GetAllConnectedClients() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantCounts := make(map[string]int)
	for _, client := range s.hub.clients {
		tenantCounts[client.TenantID]++
	}
	return tenantCounts
}

// GetClientInfo returns information about a specific client
func (s *EventService) GetClientInfo(clientID string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, exists := s.hub.clients[clientID]; exists {
		// Return a copy to avoid race conditions
		return &Client{
			ID:          client.ID,
			TenantID:    client.TenantID,
			Username:    client.Username,
			ConnectedAt: client.ConnectedAt,
			// Don't return the channel for safety
		}
	}
	return nil
}

// runHub is the main event loop that manages connections and broadcasts
func (s *EventService) runHub(ctx api.ApiContext) {
	logging.Info("Event hub starting main loop")

	// Statistics ticker for periodic logging
	statsTicker := time.NewTicker(5 * time.Minute)
	defer statsTicker.Stop()

	// Track channel references to only close when all clients using it are gone
	channelRefs := make(map[chan *Event]int)

	for {
		select {
		case <-ctx.Done():
			logging.Info("Hub context cancelled, shutting down event loop")
			return

		case client := <-s.hub.register:
			s.mu.Lock()
			s.hub.clients[client.ID] = client
			channelRefs[client.Channel]++

			// Update statistics
			s.stats.mu.Lock()
			s.stats.TotalConnections++
			s.stats.ActiveConnections++
			s.stats.ConnectionsByTenant[client.TenantID]++
			s.stats.mu.Unlock()

			s.mu.Unlock()

			logging.WithFields(logrus.Fields{
				"client_id": client.ID,
				"username":  client.Username,
				"tenant_id": client.TenantID,
				"active":    s.stats.ActiveConnections,
			}).Info("SSE client registered")

		case client := <-s.hub.unregister:
			s.mu.Lock()
			if existingClient, ok := s.hub.clients[client.ID]; ok {
				delete(s.hub.clients, client.ID)

				// Decrease channel reference count
				ch := existingClient.Channel
				channelRefs[ch]--

				// Only close the channel if no more clients are using it
				if channelRefs[ch] <= 0 {
					s.safeCloseChannel(ch)
					delete(channelRefs, ch)
				}

				// Update statistics
				s.stats.mu.Lock()
				s.stats.ActiveConnections--
				if s.stats.ConnectionsByTenant[existingClient.TenantID] > 0 {
					s.stats.ConnectionsByTenant[existingClient.TenantID]--
				}
				s.stats.mu.Unlock()

				duration := time.Since(existingClient.ConnectedAt)
				logging.WithFields(logrus.Fields{
					"client_id": client.ID,
					"username":  existingClient.Username,
					"duration":  duration,
					"active":    s.stats.ActiveConnections,
				}).Info("SSE client unregistered")
			}
			s.mu.Unlock()

		case event := <-s.hub.broadcast:
			s.mu.RLock()

			// Count how many clients will receive this event
			targetClients := 0
			sentCount := 0

			// Send event only to clients of the same tenant
			for _, client := range s.hub.clients {
				if client.TenantID == event.TenantID {
					targetClients++
					select {
					case client.Channel <- event:
						sentCount++
					default:
						// Client channel is full or closed, remove the client
						logging.WithFields(logrus.Fields{
							"client_id": client.ID,
							"username":  client.Username,
						}).Warn("SSE client channel full/closed, scheduling removal")
						go func(clientID string) {
							s.UnregisterClient(clientID)
						}(client.ID)
					}
				}
			}
			s.mu.RUnlock()

			logging.WithFields(logrus.Fields{
				"event_type":     event.Type,
				"sent_count":     sentCount,
				"target_clients": targetClients,
				"tenant_id":      event.TenantID,
			}).Debug("Event broadcasted")

		case <-statsTicker.C:
			// Log periodic statistics
			s.stats.mu.RLock()
			logging.WithFields(logrus.Fields{
				"active_connections":    s.stats.ActiveConnections,
				"total_connections":     s.stats.TotalConnections,
				"connections_by_tenant": s.stats.ConnectionsByTenant,
			}).Debug("SSE Statistics")
			s.stats.mu.RUnlock()
		}
	}
}
