// Package seeds provides a service for managing database seeding and migrations
package seeds

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MigrationRecord represents a record in the _migrations table
type MigrationRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(255);not null;uniqueIndex"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	AppliedAt   time.Time `json:"applied_at" gorm:"column:applied_at;autoCreateTime"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(50);not null;default:'applied'"`
	Error       string    `json:"error,omitempty" gorm:"column:error;type:text"`
}

func (m *MigrationRecord) TableName() string {
	return "_migrations"
}

// MigrationService manages database seeding and migrations
type MigrationService struct {
	db         *gorm.DB
	workers    []interfaces.MigrationWorker
	mu         sync.RWMutex
	applied    map[string]bool
	failed     map[string]bool
	maxRetries int
}

// NewMigrationService creates a new seed service instance
func NewMigrationService(db *gorm.DB) *MigrationService {
	service := &MigrationService{
		db:         db,
		workers:    make([]interfaces.MigrationWorker, 0),
		applied:    make(map[string]bool),
		failed:     make(map[string]bool),
		maxRetries: 3, // Default max retries
	}

	// Initialize the migrations table
	service.initializeMigrationsTable()

	// Load existing migrations
	service.loadAppliedMigrations()

	return service
}

// SetMaxRetries sets the maximum number of retry attempts for failed migrations
func (s *MigrationService) SetMaxRetries(maxRetries int) {
	s.maxRetries = maxRetries
}

// initializeMigrationsTable creates the _migrations table if it doesn't exist
func (s *MigrationService) initializeMigrationsTable() {
	if err := s.db.AutoMigrate(&MigrationRecord{}); err != nil {
		panic(err)
	}
}

// loadAppliedMigrations loads the list of already applied migrations
func (s *MigrationService) loadAppliedMigrations() {
	var records []MigrationRecord
	if err := s.db.Find(&records).Error; err != nil {
		// If table doesn't exist yet, that's fine
		return
	}

	for _, record := range records {
		if record.Status == "applied" {
			s.applied[record.Name] = true
		} else if record.Status == "failed" {
			s.failed[record.Name] = true
		}
	}
}

// Register registers a seed worker with the service
func (s *MigrationService) Register(worker interfaces.MigrationWorker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if worker is already registered
	for _, existing := range s.workers {
		if existing.GetName() == worker.GetName() {
			return // Already registered
		}
	}

	s.workers = append(s.workers, worker)
}

// RunAll runs all registered seeds that haven't been applied yet
func (s *MigrationService) RunAll(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("run_all_seeds")
	defer diag.Complete()

	diag.AddPathEntry("start", "seeds", map[string]interface{}{
		"total_workers": len(s.workers),
		"max_retries":   s.maxRetries,
	})

	ctx.LogInfo("Starting to run all seeds")

	s.mu.RLock()
	workers := make([]interfaces.MigrationWorker, len(s.workers))
	copy(workers, s.workers)
	s.mu.RUnlock()

	for i, worker := range workers {
		workerName := worker.GetName()

		// Check if already successfully applied
		if s.isApplied(workerName) {
			ctx.LogWithField("seed_name", workerName).Debug("Seed already applied, skipping")
			continue
		}

		ctx.LogWithFields(map[string]interface{}{
			"seed_name":   workerName,
			"seed_index":  i + 1,
			"total_seeds": len(workers),
			"description": worker.GetDescription(),
		}).Info("Running seed")

		diag.AddPathEntry("running_seed", "seeds", map[string]interface{}{
			"seed_name":   workerName,
			"seed_index":  i + 1,
			"description": worker.GetDescription(),
		})

		// Run the migration with retry logic
		if err := s.runMigrationWithRetry(ctx, worker, diag); err != nil {
			return diag
		}
	}

	ctx.LogInfo("All seeds completed successfully")

	diag.AddPathEntry("completed", "seeds", map[string]interface{}{
		"total_workers": len(workers),
	})

	return diag
}

// runMigrationWithRetry runs a single migration with retry logic
func (s *MigrationService) runMigrationWithRetry(ctx *appctx.AppContext, worker interfaces.MigrationWorker, diag *diagnostics.Diagnostics) error {
	workerName := worker.GetName()

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		ctx.LogWithFields(map[string]interface{}{
			"seed_name":   workerName,
			"attempt":     attempt,
			"max_retries": s.maxRetries,
		}).Info("Running migration attempt")

		// Run the Up migration
		upDiag := worker.Up(ctx)

		if !upDiag.HasErrors() {
			// Success - record and mark as applied
			s.recordMigration(workerName, worker.GetDescription(), "applied", "")
			s.markAsApplied(workerName)
			s.removeFromFailed(workerName)

			ctx.LogWithField("seed_name", workerName).Info("Migration applied successfully")
			diag.AddPathEntry("seed_applied", "seeds", map[string]interface{}{
				"seed_name": workerName,
				"attempt":   attempt,
			})
			return nil
		}

		// Migration failed
		ctx.LogWithFields(map[string]interface{}{
			"seed_name": workerName,
			"attempt":   attempt,
			"errors":    upDiag.GetSummary(),
		}).Error("Migration attempt failed")

		diag.Append(upDiag)

		// Record the failure
		s.recordMigration(workerName, worker.GetDescription(), "failed", upDiag.GetSummary())
		s.markAsFailed(workerName)

		// Attempt rollback
		downDiag := worker.Down(ctx)
		if downDiag.HasErrors() {
			ctx.LogWithFields(map[string]interface{}{
				"seed_name": workerName,
				"errors":    downDiag.GetSummary(),
			}).Error("Migration rollback also failed")

			diag.AddError("ROLLBACK_FAILED", "Migration rollback failed", "seeds", map[string]interface{}{
				"seed_name":  workerName,
				"up_error":   upDiag.GetSummary(),
				"down_error": downDiag.GetSummary(),
			})
			diag.Append(downDiag)
		} else {
			ctx.LogWithField("seed_name", workerName).Info("Migration rollback successful")
		}

		// If this was the last attempt, return error
		if attempt == s.maxRetries {
			diag.AddError("MIGRATION_FAILED_MAX_RETRIES", "Migration failed after maximum retry attempts", "seeds", map[string]interface{}{
				"seed_name":   workerName,
				"max_retries": s.maxRetries,
				"errors":      upDiag.GetSummary(),
			})
			return fmt.Errorf("migration %s failed after %d attempts", workerName, s.maxRetries)
		}

		// Wait before retrying (exponential backoff)
		waitTime := time.Duration(attempt) * time.Second
		ctx.LogWithFields(map[string]interface{}{
			"seed_name": workerName,
			"wait_time": waitTime.String(),
		}).Info("Waiting before retry")
		time.Sleep(waitTime)
	}

	return nil
}

// findWorkerByName finds a worker by name
func (s *MigrationService) findWorkerByName(name string) interfaces.MigrationWorker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, worker := range s.workers {
		if worker.GetName() == name {
			return worker
		}
	}
	return nil
}

// Rollback rolls back a specific seed by name
func (s *MigrationService) Rollback(ctx *appctx.AppContext, seedName string) *diagnostics.Diagnostics {
	diag := diagnostics.New("rollback_seed")
	defer diag.Complete()

	diag.AddPathEntry("start", "seeds", map[string]interface{}{
		"seed_name": seedName,
	})

	ctx.LogWithField("seed_name", seedName).Info("Starting seed rollback")

	// Check if seed is applied
	if !s.isApplied(seedName) {
		diag.AddError("SEED_NOT_APPLIED", "Seed is not applied", "seeds", map[string]interface{}{
			"seed_name": seedName,
		})
		return diag
	}

	// Find the worker
	worker := s.findWorkerByName(seedName)
	if worker == nil {
		diag.AddError("WORKER_NOT_FOUND", "Worker not found", "seeds", map[string]interface{}{
			"seed_name": seedName,
		})
		return diag
	}

	// Run the Down migration
	downDiag := worker.Down(ctx)
	if downDiag.HasErrors() {
		ctx.LogWithFields(map[string]interface{}{
			"seed_name": seedName,
			"errors":    downDiag.GetSummary(),
		}).Error("Seed rollback failed")

		diag.AddError("ROLLBACK_FAILED", "Seed rollback failed", "seeds", map[string]interface{}{
			"seed_name": seedName,
			"errors":    downDiag.GetSummary(),
		})

		return diag
	}

	// Remove from applied list and database
	s.removeMigration(seedName)

	ctx.LogWithField("seed_name", seedName).Info("Seed rollback completed successfully")

	diag.AddPathEntry("rollback_completed", "seeds", map[string]interface{}{
		"seed_name": seedName,
	})

	return diag
}

// GetAppliedSeeds returns a list of applied seed names
func (s *MigrationService) GetAppliedSeeds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	applied := make([]string, 0, len(s.applied))
	for name := range s.applied {
		applied = append(applied, name)
	}

	return applied
}

// GetFailedSeeds returns a list of failed seed names
func (s *MigrationService) GetFailedSeeds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	failed := make([]string, 0, len(s.failed))
	for name := range s.failed {
		failed = append(failed, name)
	}

	return failed
}

// GetPendingSeeds returns a list of pending seed names
func (s *MigrationService) GetPendingSeeds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pending := make([]string, 0)
	for _, worker := range s.workers {
		name := worker.GetName()
		if !s.applied[name] && !s.failed[name] {
			pending = append(pending, name)
		}
	}

	return pending
}

// GetRegisteredSeeds returns a list of all registered seed names
func (s *MigrationService) GetRegisteredSeeds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	registered := make([]string, 0, len(s.workers))
	for _, worker := range s.workers {
		registered = append(registered, worker.GetName())
	}

	return registered
}

// isApplied checks if a seed has been successfully applied
func (s *MigrationService) isApplied(seedName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.applied[seedName]
}

// markAsApplied marks a seed as successfully applied
func (s *MigrationService) markAsApplied(seedName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applied[seedName] = true
	delete(s.failed, seedName)
}

// markAsFailed marks a seed as failed
func (s *MigrationService) markAsFailed(seedName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed[seedName] = true
	delete(s.applied, seedName)
}

// removeFromFailed removes a seed from the failed list
func (s *MigrationService) removeFromFailed(seedName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.failed, seedName)
}

// recordMigration records a migration in the database
func (s *MigrationService) recordMigration(name, description, status, error string) {
	// First, remove any existing record for this migration
	s.db.Where("name = ?", name).Delete(&MigrationRecord{})

	record := MigrationRecord{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      status,
		Error:       error,
	}

	s.db.Create(&record)
}

// removeMigration removes a migration from the database and applied list
func (s *MigrationService) removeMigration(seedName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from database
	s.db.Where("name = ?", seedName).Delete(&MigrationRecord{})

	// Remove from applied and failed lists
	delete(s.applied, seedName)
	delete(s.failed, seedName)
}
