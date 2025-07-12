package types

import (
	"time"

	"gorm.io/gorm"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "pending"
	MessageStatusProcessing MessageStatus = "processing"
	MessageStatusCompleted  MessageStatus = "completed"
	MessageStatusFailed     MessageStatus = "failed"
	MessageStatusRetrying   MessageStatus = "retrying"
	MessageStatusAbandoned  MessageStatus = "abandoned"
)

// Message represents a message in the database
type Message struct {
	BaseModel
	Type        string         `gorm:"not null;index" json:"type"`
	Priority    int            `gorm:"not null;default:1;index" json:"priority"`
	Payload     string         `gorm:"type:text" json:"payload"` // JSON string
	TenantID    string         `gorm:"not null;index" json:"tenant_id"`
	Status      MessageStatus  `gorm:"not null;default:'pending';index" json:"status"`
	RetryCount  int            `gorm:"not null;default:0" json:"retry_count"`
	MaxRetries  int            `gorm:"not null;default:3" json:"max_retries"`
	ScheduledAt *time.Time     `gorm:"index" json:"scheduled_at,omitempty"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	FailedAt    *time.Time     `json:"failed_at,omitempty"`
	Error       string         `gorm:"type:text" json:"error,omitempty"`
	WorkerName  string         `gorm:"index" json:"worker_name,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// MessageStats represents statistics about messages
type MessageStats struct {
	TotalPending    int64 `json:"total_pending"`
	TotalProcessing int64 `json:"total_processing"`
	TotalCompleted  int64 `json:"total_completed"`
	TotalFailed     int64 `json:"total_failed"`
	TotalRetrying   int64 `json:"total_retrying"`
	TotalAbandoned  int64 `json:"total_abandoned"`
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// MessageEvent represents events that happen to messages
type MessageEvent struct {
	BaseModel
	MessageID  string        `gorm:"not null;index" json:"message_id"`
	Message    Message       `gorm:"foreignKey:MessageID" json:"message"`
	EventType  string        `gorm:"not null" json:"event_type"` // created, processing, completed, failed, retrying, abandoned
	Status     MessageStatus `gorm:"not null" json:"status"`
	WorkerName string        `json:"worker_name,omitempty"`
	Error      string        `gorm:"type:text" json:"error,omitempty"`
	Metadata   string        `gorm:"type:text" json:"metadata,omitempty"` // JSON string for additional data
	Timestamp  time.Time     `gorm:"not null" json:"timestamp"`
}

// TableName specifies the table name for MessageEvent
func (MessageEvent) TableName() string {
	return "message_events"
}

// Worker represents a registered worker
type Worker struct {
	BaseModel
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Type        string         `gorm:"not null" json:"type"` // rabbitmq, interval, hybrid, database
	MessageType string         `gorm:"not null;index" json:"message_type"`
	Interval    *int64         `json:"interval,omitempty"` // in seconds, for interval workers
	Enabled     bool           `gorm:"not null;default:true" json:"enabled"`
	IsRunning   bool           `gorm:"not null;default:false" json:"is_running"`
	LastSeen    *time.Time     `json:"last_seen,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Worker
func (Worker) TableName() string {
	return "workers"
}

// WorkerStats represents statistics about workers
type WorkerStats struct {
	WorkerID          string     `json:"worker_id"`
	WorkerName        string     `json:"worker_name"`
	MessagesProcessed int64      `json:"messages_processed"`
	MessagesFailed    int64      `json:"messages_failed"`
	LastProcessedAt   *time.Time `json:"last_processed_at,omitempty"`
	IsRunning         bool       `json:"is_running"`
}

// TableName specifies the table name for WorkerStats
func (WorkerStats) TableName() string {
	return "worker_stats"
}
