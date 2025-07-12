// Package messages provides a service for processing messages
package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/sirupsen/logrus"
)

// ServiceMessage represents a message for the service (non-generic version)
type ServiceMessage struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	TenantID  string          `json:"tenant_id"`
	Priority  int             `json:"priority"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ServiceWorker interface for the service (non-generic version)
type ServiceWorker interface {
	GetMetadata() *WorkerMetadata
	Initialize(ctx context.Context) error
	Process(ctx context.Context, message *ServiceMessage) error
}

// registeredWorker holds a worker instance and its metadata
type registeredWorker struct {
	worker      ServiceWorker
	metadata    *WorkerMetadata
	messageType string
}

// SystemMessageService is a singleton service that manages message processing
type SystemMessageService struct {
	config    *config.Config
	workers   map[string]registeredWorker // key: worker name
	store     *stores.MessageDataStore
	workersMu sync.RWMutex
	running   bool
	stopCh    chan struct{}
}

var (
	instance *SystemMessageService
	once     sync.Once
)

func Initialize(store *stores.MessageDataStore) (*SystemMessageService, error) {
	var initErr error
	once.Do(func() {
		svc, initErr := newService(store, config.GetInstance().Get())
		if initErr == nil {
			instance = svc
		}
	})
	return instance, initErr
}

func GetInstance() *SystemMessageService {
	if instance == nil {
		logging.Warn("MessageProcessorService not initialized. Call Initialize() first.")
	}
	return instance
}

func newService(store *stores.MessageDataStore, config *config.Config) (*SystemMessageService, error) {
	return &SystemMessageService{
		workers: make(map[string]registeredWorker),
		store:   store,
		stopCh:  make(chan struct{}),
		config:  config,
	}, nil
}

// getPollInterval gets the poll interval from config with default
func (s *SystemMessageService) getPollInterval() time.Duration {
	return s.config.GetDuration(config.MessageProcessorPollIntervalKey, 1*time.Second)
}

// getRecoveryEnabled gets recovery enabled setting from config with default
func (s *SystemMessageService) getRecoveryEnabled() bool {
	return s.config.GetBool(config.MessageProcessorRecoveryEnabledKey, true)
}

// getMaxProcessingAge gets max processing age from config with default
func (s *SystemMessageService) getMaxProcessingAge() time.Duration {
	return s.config.GetDuration(config.MessageProcessorMaxProcessingAgeKey, 5*time.Minute)
}

// getRecoveryInterval gets recovery interval from config with default
func (s *SystemMessageService) getRecoveryInterval() time.Duration {
	// Use the same interval as max processing age for recovery
	return s.getMaxProcessingAge()
}

// getCleanupEnabled gets cleanup enabled setting from config with default
func (s *SystemMessageService) getCleanupEnabled() bool {
	return s.config.GetBool(config.MessageProcessorCleanupEnabledKey, true)
}

// getCleanupMaxAge gets cleanup max age from config with default
func (s *SystemMessageService) getCleanupMaxAge() time.Duration {
	return s.config.GetDuration(config.MessageProcessorCleanupMaxAgeKey, 7*24*time.Hour)
}

// getCleanupInterval gets cleanup interval from config with default
func (s *SystemMessageService) getCleanupInterval() time.Duration {
	return s.config.GetDuration(config.MessageProcessorCleanupIntervalKey, 1*time.Hour)
}

// getDefaultMaxRetries gets default max retries from config with default
func (s *SystemMessageService) getDefaultMaxRetries() int {
	return s.config.GetInt(config.MessageProcessorDefaultMaxRetriesKey, 3)
}

// RegisterWorker registers a worker that implements the ServiceWorker interface
func (s *SystemMessageService) RegisterWorker(worker ServiceWorker) error {
	s.workersMu.Lock()
	defer s.workersMu.Unlock()

	metadata := worker.GetMetadata()
	if metadata == nil {
		return fmt.Errorf("worker metadata is nil")
	}

	if metadata.Name == "" {
		return fmt.Errorf("worker name cannot be empty")
	}

	if metadata.MessageType == "" {
		return fmt.Errorf("worker message type cannot be empty")
	}

	// Check if worker with same name already exists
	if _, exists := s.workers[metadata.Name]; exists {
		return fmt.Errorf("worker with name '%s' already registered", metadata.Name)
	}

	// Check if another worker already handles this message type
	for _, existingWorker := range s.workers {
		if existingWorker.messageType == string(metadata.MessageType) {
			return fmt.Errorf("message type '%s' is already handled by worker '%s'",
				metadata.MessageType, existingWorker.metadata.Name)
		}
	}

	// Register the worker
	s.workers[metadata.Name] = registeredWorker{
		worker:      worker,
		metadata:    metadata,
		messageType: string(metadata.MessageType),
	}

	// Initialize the worker
	ctx := context.Background()
	if err := worker.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize worker '%s': %w", metadata.Name, err)
	}

	logging.WithFields(logrus.Fields{
		"worker_name":  metadata.Name,
		"message_type": metadata.MessageType,
	}).Info("Registered worker")

	return nil
}

// Start starts the message processing service
func (s *SystemMessageService) Start(ctx context.Context) error {
	s.workersMu.Lock()
	defer s.workersMu.Unlock()

	if s.running {
		return fmt.Errorf("service is already running")
	}

	if len(s.workers) == 0 {
		return fmt.Errorf("no workers registered")
	}

	// Perform startup recovery to handle orphaned processing messages
	if err := s.performStartupRecovery(ctx); err != nil {
		logging.WithError(err).Warn("Failed to perform startup recovery")
	} else {
		logging.WithField("recovered_count", 0).Info("Startup recovery completed")
	}

	s.running = true
	go s.listenForMessages(ctx)

	logging.WithField("worker_count", len(s.workers)).Info("Message processor service started")
	return nil
}

// performStartupRecovery recovers orphaned processing messages
func (s *SystemMessageService) performStartupRecovery(ctx context.Context) error {
	if s.store == nil {
		return fmt.Errorf("MessageDataStore is not initialized")
	}

	if !s.getRecoveryEnabled() {
		if s.config.IsDebug() {
			logging.Info("Recovery disabled, skipping startup recovery")
		}
		return nil
	}

	// Recover messages that have been stuck in processing state for too long
	recovered, err := s.store.RecoverOrphanedMessages(ctx, s.getMaxProcessingAge())
	if err != nil {
		return fmt.Errorf("failed to recover orphaned messages: %w", err)
	}

	if recovered > 0 {
		logging.WithField("recovered_count", recovered).Info("Startup recovery completed")
	}

	return nil
}

// Stop stops the message processing service
func (s *SystemMessageService) Stop(ctx context.Context) error {
	s.workersMu.Lock()
	defer s.workersMu.Unlock()

	if !s.running {
		return fmt.Errorf("service is not running")
	}

	s.running = false
	close(s.stopCh)

	logging.Info("Message processor service stopped")
	return nil
}

// IsRunning returns whether the service is currently running
func (s *SystemMessageService) IsRunning() bool {
	s.workersMu.RLock()
	defer s.workersMu.RUnlock()
	return s.running
}

// GetRegisteredWorkers returns information about all registered workers
func (s *SystemMessageService) GetRegisteredWorkers() []*WorkerMetadata {
	s.workersMu.RLock()
	defer s.workersMu.RUnlock()

	workers := make([]*WorkerMetadata, 0, len(s.workers))
	for _, worker := range s.workers {
		workers = append(workers, worker.metadata)
	}
	return workers
}

// Polls the DB for new messages and dispatches to workers
func (s *SystemMessageService) listenForMessages(ctx context.Context) {
	if s.store == nil {
		logging.Error("MessageDataStore is not initialized!")
		return
	}

	ticker := time.NewTicker(s.getPollInterval())
	var recoveryTicker *time.Ticker
	var cleanupTicker *time.Ticker

	if s.getRecoveryEnabled() {
		recoveryTicker = time.NewTicker(s.getRecoveryInterval())
		defer recoveryTicker.Stop()
	}

	if s.getCleanupEnabled() {
		cleanupTicker = time.NewTicker(s.getCleanupInterval())
		defer cleanupTicker.Stop()
	}

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logging.Info("Context cancelled, stopping message listener")
			return
		case <-s.stopCh:
			logging.Info("Stop signal received, stopping message listener")
			return
		case <-ticker.C:
			messages, err := s.store.GetPendingMessages(ctx, 10)
			if err != nil {
				logging.WithError(err).Error("Error getting pending messages")
				continue
			}

			for _, dbMsg := range messages {
				// Try to mark the message as processing (this acts as a lock)
				if err := s.store.MarkMessageProcessing(ctx, dbMsg.ID); err != nil {
					if s.config.IsDebug() {
						logging.WithFields(logrus.Fields{
							"message_id": dbMsg.ID,
							"error":      err,
						}).Debug("Failed to mark message as processing (likely already being processed)")
					}
					continue
				}

				// Dispatch the message for processing
				s.dispatchMessage(ctx, dbMsg)
			}
		case <-recoveryTicker.C:
			if s.getRecoveryEnabled() {
				// Periodic recovery of orphaned messages
				if recovered, err := s.store.RecoverOrphanedMessages(ctx, s.getMaxProcessingAge()); err != nil {
					logging.WithError(err).Error("Error during periodic recovery")
				} else if recovered > 0 {
					logging.WithField("recovered_count", recovered).Info("Periodic recovery completed")
				}
			}
		case <-cleanupTicker.C:
			if s.getCleanupEnabled() {
				// Periodic cleanup of old messages
				if cleaned, err := s.store.CleanupOldMessages(ctx, s.getCleanupMaxAge()); err != nil {
					logging.WithError(err).Error("Error during periodic cleanup")
				} else if cleaned > 0 {
					logging.WithField("cleaned_count", cleaned).Info("Periodic cleanup completed")
				}
			}
		}
	}
}

// dispatchMessage dispatches a message to the appropriate worker
func (s *SystemMessageService) dispatchMessage(ctx context.Context, dbMsg *types.Message) {
	s.workersMu.RLock()
	defer s.workersMu.RUnlock()

	// Find worker that handles this message type
	var targetWorker *registeredWorker
	for _, worker := range s.workers {
		if worker.messageType == dbMsg.Type {
			targetWorker = &worker
			break
		}
	}

	if targetWorker == nil {
		logging.WithField("message_type", dbMsg.Type).Warn("No worker found for message type")
		// Mark message as failed since no worker can handle it
		s.updateMessageStatus(ctx, dbMsg.ID, "failed")
		return
	}

	// Convert database message to service message
	message := &ServiceMessage{
		ID:        dbMsg.ID,
		Type:      dbMsg.Type,
		Payload:   []byte(dbMsg.Payload),
		TenantID:  dbMsg.TenantID,
		Priority:  dbMsg.Priority,
		Status:    string(dbMsg.Status),
		CreatedAt: dbMsg.CreatedAt,
		UpdatedAt: dbMsg.UpdatedAt,
	}

	// Process the message
	go func() {
		if err := targetWorker.worker.Process(ctx, message); err != nil {
			logging.WithFields(logrus.Fields{
				"message_id": message.ID,
				"error":      err,
			}).Error("Error processing message")
			// Update message status to failed
			s.updateMessageStatus(ctx, message.ID, "failed")
		} else {
			// Update message status to completed
			s.updateMessageStatus(ctx, message.ID, "completed")
		}
	}()
}

// updateMessageStatus updates the status of a message in the database
func (s *SystemMessageService) updateMessageStatus(ctx context.Context, messageID, status string) {
	if s.store == nil {
		logging.Error("Cannot update message status: MessageDataStore is not initialized")
		return
	}

	messageStatus := types.MessageStatus(status)
	var err error

	switch messageStatus {
	case types.MessageStatusCompleted:
		err = s.store.CompleteMessage(ctx, messageID)
	case types.MessageStatusFailed:
		err = s.store.FailMessage(ctx, messageID, "Processing failed")
	default:
		err = s.store.UpdateMessageStatus(ctx, messageID, messageStatus, "")
	}

	if err != nil {
		logging.WithFields(logrus.Fields{
			"message_id": messageID,
			"status":     status,
			"error":      err,
		}).Error("Error updating message status")
	}
}

// PostMessage posts a message with a generic payload into the database
func (s *SystemMessageService) PostMessage(ctx context.Context, messageType string, payload interface{}, tenantID string, priority int) (string, error) {
	if s.store == nil {
		return "", fmt.Errorf("MessageDataStore is not initialized")
	}

	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Create database message
	dbMsg := &types.Message{
		Type:       messageType,
		Payload:    string(payloadBytes),
		TenantID:   tenantID,
		Priority:   priority,
		Status:     types.MessageStatusPending,
		MaxRetries: s.getDefaultMaxRetries(),
	}

	err = s.store.CreateMessage(ctx, dbMsg)
	if err != nil {
		return "", err
	}

	logging.WithFields(logrus.Fields{
		"message_id":   dbMsg.ID,
		"message_type": messageType,
		"priority":     priority,
	}).Info("Posted message")

	return dbMsg.ID, nil
}

// PostSimple posts a simple message with just a string payload
func (s *SystemMessageService) PostSimple(ctx context.Context, messageType string, payload string, tenantID string) (string, error) {
	return s.PostMessage(ctx, messageType, payload, tenantID, 0)
}
