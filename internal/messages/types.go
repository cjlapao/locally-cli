package messages

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cjlapao/locally-cli/internal/database/types"
)

// MessageType represents the type of message a worker can handle
type MessageType string

// MessagePriority represents the priority of a message
type MessagePriority int

const (
	PriorityLow    MessagePriority = 0
	PriorityNormal MessagePriority = 1
	PriorityHigh   MessagePriority = 2
	PriorityUrgent MessagePriority = 3
)

// Message represents a message with generic payload
type Message[T any] struct {
	ID         string          `json:"id"`
	Type       MessageType     `json:"type"`
	Priority   MessagePriority `json:"priority"`
	Payload    T               `json:"payload"`
	TenantID   string          `json:"tenant_id"`
	CreatedAt  time.Time       `json:"created_at"`
	RetryCount int             `json:"retry_count"`
	MaxRetries int             `json:"max_retries"`
}

// WorkerMetadata contains metadata about a worker
type WorkerMetadata struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	MessageType MessageType `json:"message_type"`
	Enabled     bool        `json:"enabled"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Worker interface that all workers must implement
type Worker[T any] interface {
	// GetMetadata returns the worker's metadata
	GetMetadata() *WorkerMetadata

	// Initialize initializes the worker
	Initialize(ctx context.Context) error

	// Process processes a message with typed payload
	Process(ctx context.Context, message *Message[T]) error
}

// Service interface for the message processor service
type Service interface {
	// Initialize initializes the message processor service
	Initialize(ctx context.Context) error

	// Start starts the message processor service
	Start(ctx context.Context) error

	// Stop stops the message processor service
	Stop(ctx context.Context) error

	// RegisterWorker registers a worker
	RegisterWorker(worker interface{}) error

	// UnregisterWorker unregisters a worker
	UnregisterWorker(workerName string) error

	// GetWorkers returns all registered workers
	GetWorkers() []*WorkerMetadata

	// IsRunning returns true if the service is running
	IsRunning() bool
}

// ConvertToTypedMessage converts a database message to a typed message
func ConvertToTypedMessage[T any](dbMessage *types.Message) (*Message[T], error) {
	var payload T
	if dbMessage.Payload != "" {
		if err := json.Unmarshal([]byte(dbMessage.Payload), &payload); err != nil {
			return nil, err
		}
	}

	return &Message[T]{
		ID:         dbMessage.ID,
		Type:       MessageType(dbMessage.Type),
		Priority:   MessagePriority(dbMessage.Priority),
		Payload:    payload,
		TenantID:   dbMessage.TenantID,
		CreatedAt:  dbMessage.CreatedAt,
		RetryCount: dbMessage.RetryCount,
		MaxRetries: dbMessage.MaxRetries,
	}, nil
}
