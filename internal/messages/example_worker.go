package messages

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/sirupsen/logrus"
)

// EmailPayload is an example payload type
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type NotificationPayload struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// EmailWorker implements ServiceWorker
type EmailWorker struct {
	name        string
	description string
	version     string
}

// NewEmailWorker creates a new email worker
func NewEmailWorker() *EmailWorker {
	return &EmailWorker{
		name:        "email-worker",
		description: "Processes email messages",
		version:     "1.0.0",
	}
}

// GetMetadata returns the worker's metadata
func (w *EmailWorker) GetMetadata() *WorkerMetadata {
	return &WorkerMetadata{
		Name:        w.name,
		Description: w.description,
		Version:     w.version,
		MessageType: "email",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Initialize initializes the worker
func (w *EmailWorker) Initialize(ctx context.Context) error {
	logging.WithField("worker_name", w.name).Info("Initializing email worker")
	return nil
}

// Process processes an email message
func (w *EmailWorker) Process(ctx context.Context, message *ServiceMessage) error {
	// Parse the payload
	var payload EmailPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"message_id": message.ID,
		"to":         payload.To,
		"subject":    payload.Subject,
	}).Info("Email worker processing message")

	// Simulate email processing
	time.Sleep(100 * time.Millisecond)

	logging.WithField("to", payload.To).Info("Email sent successfully")
	return nil
}

// NotificationWorker implements ServiceWorker
type NotificationWorker struct {
	name        string
	description string
	version     string
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker() *NotificationWorker {
	return &NotificationWorker{
		name:        "notification-worker",
		description: "Processes notification messages",
		version:     "1.0.0",
	}
}

// GetMetadata returns the worker's metadata
func (w *NotificationWorker) GetMetadata() *WorkerMetadata {
	return &WorkerMetadata{
		Name:        w.name,
		Description: w.description,
		Version:     w.version,
		MessageType: "notification",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Initialize initializes the worker
func (w *NotificationWorker) Initialize(ctx context.Context) error {
	logging.WithField("worker_name", w.name).Info("Initializing notification worker")
	return nil
}

// Process processes a notification message
func (w *NotificationWorker) Process(ctx context.Context, message *ServiceMessage) error {
	// Parse the payload
	var payload NotificationPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"message_id": message.ID,
		"user_id":    payload.UserID,
		"message":    payload.Message,
	}).Info("Notification worker processing message")

	// Simulate notification processing
	time.Sleep(50 * time.Millisecond)

	logging.WithField("user_id", payload.UserID).Info("Notification sent successfully")
	return nil
}

// ExampleUsage is an example usage function
func ExampleUsage() {
	// Get the singleton service
	service := GetInstance()
	if service == nil {
		logging.Error("Service not initialized. Call Initialize() first.")
		return
	}

	// Create workers
	emailWorker := NewEmailWorker()
	notificationWorker := NewNotificationWorker()

	// Register workers with the service
	if err := service.RegisterWorker(emailWorker); err != nil {
		logging.WithError(err).Error("Failed to register email worker")
	}

	if err := service.RegisterWorker(notificationWorker); err != nil {
		logging.WithError(err).Error("Failed to register notification worker")
	}

	// Start the service
	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		logging.WithError(err).Error("Failed to start service")
	}

	logging.Info("Message processor service started successfully!")

	// The service will now poll for messages and dispatch them to workers
	// You can stop it later with: service.Stop(ctx)
}
