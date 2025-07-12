package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType represents the type of event being sent
type EventType string

const (
	EventTypeConnectionEstablished EventType = "connection.established"
	EventTypeSystemAlert           EventType = "system.alert"
	EventTypeSystemError           EventType = "system.error"
	EventTypeSystemInfo            EventType = "system.info"
	EventTypeSystemWarning         EventType = "system.warning"
	EventTypeSystemDebug           EventType = "system.debug"
	EventTypeSystemTrace           EventType = "system.trace"
	EventTypeSystemFatal           EventType = "system.fatal"
	EventTypeSystemPanic           EventType = "system.panic"
	EventTypeSystemRecover         EventType = "system.recover"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ToSSEFormat converts the event to SSE format
func (e *Event) ToSSEFormat() string {
	data, err := json.Marshal(e)
	if err != nil {
		errorData := map[string]interface{}{
			"id":        e.ID,
			"type":      "system.alert",
			"tenant_id": e.TenantID,
			"timestamp": e.Timestamp.Format(time.RFC3339),
			"data": map[string]interface{}{
				"message": "Failed to serialize event",
				"error":   err.Error(),
			},
		}
		data, _ = json.Marshal(errorData)
	}

	return fmt.Sprintf("data: %s\n\n", string(data))
}

type Client struct {
	ID          string
	TenantID    string
	Channel     chan *Event
	Username    string
	ConnectedAt time.Time
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan *Event, 100),
	}
}
