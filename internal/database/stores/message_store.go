package stores

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/google/uuid"
)

var (
	messageDataStoreInstance *MessageDataStore
	messageDataStoreOnce     sync.Once
)

type MessageDataStore struct {
	database.BaseDataStore
}

// GetMessageDataStoreInstance returns the singleton instance of the message store
func GetMessageDataStoreInstance() *MessageDataStore {
	return messageDataStoreInstance
}

func InitializeMessageDataStore() error {
	var initErr error
	cfg := config.GetInstance().Get()
	messageDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			initErr = fmt.Errorf("database service not initialized")
			return
		}

		store := &MessageDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running message migrations")
			if err := store.Migrate(); err != nil {
				initErr = fmt.Errorf("failed to run message migrations: %w", err)
				return
			}
			logging.Info("Message migrations completed")
		}

		messageDataStoreInstance = store
	})

	return initErr
}

// Migrate implements the DataStore interface
func (s *MessageDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&types.Message{}); err != nil {
		return fmt.Errorf("failed to migrate message table: %w", err)
	}
	if err := s.GetDB().AutoMigrate(&types.MessageEvent{}); err != nil {
		return fmt.Errorf("failed to migrate message event table: %w", err)
	}
	if err := s.GetDB().AutoMigrate(&types.Worker{}); err != nil {
		return fmt.Errorf("failed to migrate worker table: %w", err)
	}
	return nil
}

// CreateMessage creates a new message in the queue
func (s *MessageDataStore) CreateMessage(ctx context.Context, message *types.Message) error {
	message.ID = uuid.New().String()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

// GetPendingMessages retrieves pending messages, ordered by priority and creation time
func (s *MessageDataStore) GetPendingMessages(ctx context.Context, limit int) ([]*types.Message, error) {
	var messages []*types.Message

	query := s.GetDB().WithContext(ctx).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)", types.MessageStatusPending, time.Now()).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}

	return messages, nil
}

// GetScheduledMessages retrieves messages that are scheduled to run now
func (s *MessageDataStore) GetScheduledMessages(ctx context.Context) ([]*types.Message, error) {
	var messages []*types.Message

	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", types.MessageStatusPending, time.Now()).
		Order("priority DESC, scheduled_at ASC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get scheduled messages: %w", err)
	}

	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func (s *MessageDataStore) UpdateMessageStatus(ctx context.Context, messageID string, status types.MessageStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if errorMsg != "" {
		updates["error"] = errorMsg
	}

	if status == types.MessageStatusCompleted || status == types.MessageStatusFailed {
		updates["processed_at"] = time.Now()
	}

	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("id = ?", messageID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// MarkMessageProcessing marks a message as being processed
func (s *MessageDataStore) MarkMessageProcessing(ctx context.Context, messageID string) error {
	return s.UpdateMessageStatus(ctx, messageID, types.MessageStatusProcessing, "")
}

// CompleteMessage marks a message as completed
func (s *MessageDataStore) CompleteMessage(ctx context.Context, messageID string) error {
	return s.UpdateMessageStatus(ctx, messageID, types.MessageStatusCompleted, "")
}

// FailMessage marks a message as failed and increments retry count
func (s *MessageDataStore) FailMessage(ctx context.Context, messageID string, errorMsg string) error {
	// First, increment the retry count and check if we should retry or mark as failed
	var message types.Message
	if err := s.GetDB().WithContext(ctx).First(&message, "id = ?", messageID).Error; err != nil {
		return fmt.Errorf("failed to get message for retry: %w", err)
	}

	message.RetryCount++
	message.Error = errorMsg
	message.UpdatedAt = time.Now()

	// If we've exceeded max retries, mark as failed, otherwise mark for retry
	if message.RetryCount >= message.MaxRetries {
		message.Status = types.MessageStatusFailed
		message.ProcessedAt = &message.UpdatedAt
	} else {
		message.Status = types.MessageStatusRetrying
		// Schedule for retry (could add exponential backoff here)
		retryAt := time.Now().Add(time.Duration(message.RetryCount) * time.Minute)
		message.ScheduledAt = &retryAt
	}

	if err := s.GetDB().WithContext(ctx).Save(&message).Error; err != nil {
		return fmt.Errorf("failed to update message for retry: %w", err)
	}

	return nil
}

// DeleteMessage removes a message from the queue
func (s *MessageDataStore) DeleteMessage(ctx context.Context, messageID string) error {
	if err := s.GetDB().WithContext(ctx).Delete(&types.Message{}, "id = ?", messageID).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// GetMessageStats returns statistics about messages
func (s *MessageDataStore) GetMessageStats(ctx context.Context) (*types.MessageStats, error) {
	stats := &types.MessageStats{}

	// Count pending messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusPending).
		Count(&stats.TotalPending).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending messages: %w", err)
	}

	// Count processing messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusProcessing).
		Count(&stats.TotalProcessing).Error; err != nil {
		return nil, fmt.Errorf("failed to count processing messages: %w", err)
	}

	// Count completed messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusCompleted).
		Count(&stats.TotalCompleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed messages: %w", err)
	}

	// Count failed messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusFailed).
		Count(&stats.TotalFailed).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed messages: %w", err)
	}

	// Count retrying messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusRetrying).
		Count(&stats.TotalRetrying).Error; err != nil {
		return nil, fmt.Errorf("failed to count retrying messages: %w", err)
	}

	// Count abandoned messages
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusAbandoned).
		Count(&stats.TotalAbandoned).Error; err != nil {
		return nil, fmt.Errorf("failed to count abandoned messages: %w", err)
	}

	return stats, nil
}

// RecoverOrphanedMessages finds messages that were stuck in processing state and resets them
func (s *MessageDataStore) RecoverOrphanedMessages(ctx context.Context, maxProcessingAge time.Duration) (int, error) {
	// Find messages that have been in processing state for too long
	cutoffTime := time.Now().Add(-maxProcessingAge)

	var orphanedMessages []*types.Message
	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND updated_at < ?", types.MessageStatusProcessing, cutoffTime).
		Find(&orphanedMessages).Error; err != nil {
		return 0, fmt.Errorf("failed to find orphaned messages: %w", err)
	}

	if len(orphanedMessages) == 0 {
		return 0, nil
	}

	// Reset orphaned messages to pending status
	result := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ? AND updated_at < ?", types.MessageStatusProcessing, cutoffTime).
		Updates(map[string]interface{}{
			"status":     types.MessageStatusPending,
			"updated_at": time.Now(),
			"error":      "Recovered from orphaned processing state",
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to recover orphaned messages: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// GetStuckRetryingMessages finds messages that are stuck in retrying state and resets them to pending
func (s *MessageDataStore) GetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) ([]*types.Message, error) {
	cutoffTime := time.Now().Add(-maxRetryAge)

	var messages []*types.Message
	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at < ?)", types.MessageStatusRetrying, cutoffTime).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to find stuck retrying messages: %w", err)
	}

	return messages, nil
}

// ResetStuckRetryingMessages resets messages that are stuck in retrying state back to pending
func (s *MessageDataStore) ResetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxRetryAge)

	result := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at < ?)", types.MessageStatusRetrying, cutoffTime).
		Updates(map[string]interface{}{
			"status":       types.MessageStatusPending,
			"updated_at":   time.Now(),
			"scheduled_at": nil,
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to reset stuck retrying messages: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// CleanupOldMessages removes old completed/failed messages to prevent database bloat
func (s *MessageDataStore) CleanupOldMessages(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("(status = ? OR status = ?) AND processed_at < ?", types.MessageStatusCompleted, types.MessageStatusFailed, cutoffTime).
		Delete(&types.Message{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old messages: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// CleanupOldAbandonedMessages removes old abandoned messages
func (s *MessageDataStore) CleanupOldAbandonedMessages(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("status = ? AND updated_at < ?", types.MessageStatusAbandoned, cutoffTime).
		Delete(&types.Message{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old abandoned messages: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// CleanupOldEvents removes old message events
func (s *MessageDataStore) CleanupOldEvents(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("timestamp < ?", cutoffTime).
		Delete(&types.MessageEvent{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old events: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// CreateMessageEvent creates a new message event
func (s *MessageDataStore) CreateMessageEvent(ctx context.Context, event *types.MessageEvent) error {
	event.Timestamp = time.Now()
	if err := s.GetDB().WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to create message event: %w", err)
	}
	return nil
}

// CreateWorker creates a new worker record
func (s *MessageDataStore) CreateWorker(ctx context.Context, worker *types.Worker) error {
	worker.CreatedAt = time.Now()
	worker.UpdatedAt = time.Now()
	if err := s.GetDB().WithContext(ctx).Create(worker).Error; err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}
	return nil
}

// DeleteWorker deletes a worker record
func (s *MessageDataStore) DeleteWorker(ctx context.Context, workerName string) error {
	if err := s.GetDB().WithContext(ctx).Where("name = ?", workerName).Delete(&types.Worker{}).Error; err != nil {
		return fmt.Errorf("failed to delete worker: %w", err)
	}
	return nil
}

// GetWorkerByName retrieves a worker by name
func (s *MessageDataStore) GetWorkerByName(ctx context.Context, workerName string) (*types.Worker, error) {
	var worker types.Worker
	if err := s.GetDB().WithContext(ctx).Where("name = ?", workerName).First(&worker).Error; err != nil {
		return nil, fmt.Errorf("failed to get worker: %w", err)
	}
	return &worker, nil
}

// GetAllWorkers retrieves all workers
func (s *MessageDataStore) GetAllWorkers(ctx context.Context) ([]*types.Worker, error) {
	var workers []*types.Worker
	if err := s.GetDB().WithContext(ctx).Find(&workers).Error; err != nil {
		return nil, fmt.Errorf("failed to get workers: %w", err)
	}
	return workers, nil
}

// UpdateWorkerStatus updates worker status
func (s *MessageDataStore) UpdateWorkerStatus(ctx context.Context, workerName string, isRunning bool) error {
	if err := s.GetDB().WithContext(ctx).
		Model(&types.Worker{}).
		Where("name = ?", workerName).
		Updates(map[string]interface{}{
			"is_running": isRunning,
			"updated_at": time.Now(),
			"last_seen":  time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update worker status: %w", err)
	}
	return nil
}

// PerformStartupRecovery performs startup recovery operations
func (s *MessageDataStore) PerformStartupRecovery(ctx context.Context) error {
	// Reset processing messages to pending
	err := s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ?", types.MessageStatusProcessing).
		Updates(map[string]interface{}{
			"status":      types.MessageStatusPending,
			"worker_name": "",
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to reset processing messages: %w", err)
	}

	// Reset retrying messages to pending if they haven't exceeded max retries
	err = s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ? AND retry_count < max_retries", types.MessageStatusRetrying).
		Updates(map[string]interface{}{
			"status":     types.MessageStatusPending,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to reset retrying messages: %w", err)
	}

	// Abandon messages that have exceeded max retries
	err = s.GetDB().WithContext(ctx).
		Model(&types.Message{}).
		Where("status = ? AND retry_count >= max_retries", types.MessageStatusRetrying).
		Updates(map[string]interface{}{
			"status":     types.MessageStatusAbandoned,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to abandon exceeded retry messages: %w", err)
	}

	return nil
}
