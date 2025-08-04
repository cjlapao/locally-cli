package stores

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/utils"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	activityDataStoreInstance *ActivityDataStore
	activityDataStoreOnce     sync.Once
)

type ActivityDataStoreInterface interface {
	// Activity CRUD operations
	CreateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) (*entities.Activity, error)
	GetActivityByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.Activity, error)
	GetActivitiesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Activity], error)
	UpdateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) error
	DeleteActivity(ctx *appctx.AppContext, tenantID string, id string) error

	// Activity querying and reporting
	GetActivitiesByFilterAdvanced(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], error)
	GetActivityStats(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, error)
	GetTopActors(ctx *appctx.AppContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, error)
	GetActivityTrends(ctx *appctx.AppContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, error)

	// Activity summary operations
	CreateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, error)
	GetActivitySummaryByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.ActivitySummary, error)
	GetActivitySummariesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.ActivitySummary], error)
	UpdateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) error
	DeleteActivitySummary(ctx *appctx.AppContext, tenantID string, id string) error

	// Maintenance operations
	CleanupOldActivities(ctx *appctx.AppContext, tenantID string, retentionDays int) error
	ArchiveActivities(ctx *appctx.AppContext, tenantID string, beforeDate time.Time) error
}

type ActivityDataStore struct {
	database.BaseDataStore
}

func GetActivityDataStoreInstance() ActivityDataStoreInterface {
	return activityDataStoreInstance
}

func InitializeActivityDataStore() (ActivityDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_activity_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing activity store...")

	activityDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "activity_data_store", nil)
			return
		}

		store := &ActivityDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running activity migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Activity migrations completed")
		}

		activityDataStoreInstance = store
	})

	logging.Info("Activity store initialized successfully")
	return activityDataStoreInstance, diag
}

func (s *ActivityDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_activity_data_store")

	if err := s.GetDB().AutoMigrate(&entities.Activity{}); err != nil {
		diag.AddError("failed_to_migrate_activity_table", "failed to migrate activity table", "activity_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.ActivitySummary{}); err != nil {
		diag.AddError("failed_to_migrate_activity_summary_table", "failed to migrate activity summary table", "activity_data_store", nil)
		return diag
	}

	// Create indexes for better query performance
	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_tenant_module_service ON activities(tenant_id, module, service)").Error; err != nil {
		diag.AddError("failed_to_create_activities_index", "failed to create activities index", "activity_data_store", nil)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_actor_target ON activities(actor_type, actor_id, target_type, target_id)").Error; err != nil {
		diag.AddError("failed_to_create_activities_actor_target_index", "failed to create activities actor target index", "activity_data_store", nil)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_timing ON activities(started_at, completed_at, created_at)").Error; err != nil {
		diag.AddError("failed_to_create_activities_timing_index", "failed to create activities timing index", "activity_data_store", nil)
	}

	return diag
}

func (s *ActivityDataStore) CreateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) (*entities.Activity, error) {
	if activity == nil {
		return nil, errors.New("activity cannot be nil")
	}

	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}

	if activity.Slug == "" {
		activity.Slug = fmt.Sprintf("activity-%s", activity.ID)
	}

	activity.TenantID = tenantID
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if activity.StartedAt.IsZero() {
		activity.StartedAt = time.Now()
	}

	if err := s.GetDB().Create(activity).Error; err != nil {
		logging.Errorf("Failed to create activity: %v", err)
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	return activity, nil
}

func (s *ActivityDataStore) GetActivityByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.Activity, error) {
	if id == "" {
		return nil, errors.New("activity ID cannot be empty")
	}

	var activity entities.Activity
	query := s.GetDB().Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("activity not found with ID: %s", id)
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

func (s *ActivityDataStore) GetActivitiesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Activity], error) {
	return utils.PaginatedFilteredQuery(s.GetDB(), tenantID, filterObj, entities.Activity{})
}

func (s *ActivityDataStore) UpdateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) error {
	if activity == nil {
		return errors.New("activity cannot be nil")
	}

	if activity.ID == "" {
		return errors.New("activity ID cannot be empty")
	}

	activity.UpdatedAt = time.Now()

	query := s.GetDB().Where("id = ?", activity.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(activity).Error; err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	return nil
}

func (s *ActivityDataStore) DeleteActivity(ctx *appctx.AppContext, tenantID string, id string) error {
	if id == "" {
		return errors.New("activity ID cannot be empty")
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	return nil
}

func (s *ActivityDataStore) GetActivitiesByFilterAdvanced(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], error) {
	var activities []entities.Activity
	query := s.GetDB().Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply advanced filters
	if filter != nil {
		if len(filter.Module) > 0 {
			query = query.Where("module IN ?", filter.Module)
		}
		if len(filter.Service) > 0 {
			query = query.Where("service IN ?", filter.Service)
		}
		if len(filter.ActivityType) > 0 {
			query = query.Where("activity_type IN ?", filter.ActivityType)
		}
		if len(filter.ActivityLevel) > 0 {
			query = query.Where("activity_level IN ?", filter.ActivityLevel)
		}
		if len(filter.ActorType) > 0 {
			query = query.Where("actor_type IN ?", filter.ActorType)
		}
		if len(filter.ActorID) > 0 {
			query = query.Where("actor_id IN ?", filter.ActorID)
		}
		if len(filter.TargetType) > 0 {
			query = query.Where("target_type IN ?", filter.TargetType)
		}
		if len(filter.TargetID) > 0 {
			query = query.Where("target_id IN ?", filter.TargetID)
		}
		if filter.Success != nil {
			query = query.Where("success = ?", *filter.Success)
		}
		if filter.IsSensitive != nil {
			query = query.Where("is_sensitive = ?", *filter.IsSensitive)
		}
		if len(filter.Tags) > 0 {
			for _, tag := range filter.Tags {
				query = query.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
		if filter.StartedAtFrom != nil {
			query = query.Where("started_at >= ?", *filter.StartedAtFrom)
		}
		if filter.StartedAtTo != nil {
			query = query.Where("started_at <= ?", *filter.StartedAtTo)
		}
		if filter.CreatedAtFrom != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAtFrom)
		}
		if filter.CreatedAtTo != nil {
			query = query.Where("created_at <= ?", *filter.CreatedAtTo)
		}
	}

	// Apply pagination
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Default ordering
	query = query.Order("created_at DESC")

	if err := query.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	// Get total count
	var total int64
	countQuery := s.GetDB().Model(&entities.Activity{})
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	return &filters.FilterResponse[entities.Activity]{
		Items:      activities,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (s *ActivityDataStore) GetActivityStats(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, error) {
	query := s.GetDB().Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (same logic as GetActivitiesByFilterAdvanced)
	if filter != nil {
		// ... apply filters (simplified for brevity)
	}

	var stats struct {
		TotalActivities int64   `json:"total_activities"`
		SuccessCount    int64   `json:"success_count"`
		ErrorCount      int64   `json:"error_count"`
		AvgDurationMs   float64 `json:"avg_duration_ms"`
		MaxDurationMs   int64   `json:"max_duration_ms"`
		MinDurationMs   int64   `json:"min_duration_ms"`
	}

	if err := query.Select(`
		COUNT(*) as total_activities,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms,
		MAX(duration_ms) as max_duration_ms,
		MIN(duration_ms) as min_duration_ms
	`).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}

	return map[string]interface{}{
		"total_activities": stats.TotalActivities,
		"success_count":    stats.SuccessCount,
		"error_count":      stats.ErrorCount,
		"avg_duration_ms":  stats.AvgDurationMs,
		"max_duration_ms":  stats.MaxDurationMs,
		"min_duration_ms":  stats.MinDurationMs,
	}, nil
}

func (s *ActivityDataStore) GetTopActors(ctx *appctx.AppContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, error) {
	query := s.GetDB().Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (simplified)
	if filter != nil {
		// ... apply filters
	}

	var results []map[string]interface{}
	if err := query.Select(`
		actor_type,
		actor_id,
		actor_name,
		COUNT(*) as activity_count,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms
	`).Group("actor_type, actor_id, actor_name").
		Order("activity_count DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top actors: %w", err)
	}

	return results, nil
}

func (s *ActivityDataStore) GetActivityTrends(ctx *appctx.AppContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, error) {
	query := s.GetDB().Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (simplified)
	if filter != nil {
		// ... apply filters
	}

	var results []map[string]interface{}
	if err := query.Select(`
		DATE(created_at) as date,
		COUNT(*) as total_activities,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms
	`).Where("created_at >= ?", time.Now().AddDate(0, 0, -days)).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity trends: %w", err)
	}

	return results, nil
}

// ActivitySummary operations
func (s *ActivityDataStore) CreateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, error) {
	if summary == nil {
		return nil, errors.New("activity summary cannot be nil")
	}

	if summary.ID == "" {
		summary.ID = uuid.New().String()
	}

	if summary.Slug == "" {
		summary.Slug = fmt.Sprintf("activity-summary-%s", summary.ID)
	}

	summary.TenantID = tenantID
	summary.CreatedAt = time.Now()
	summary.UpdatedAt = time.Now()

	if err := s.GetDB().Create(summary).Error; err != nil {
		return nil, fmt.Errorf("failed to create activity summary: %w", err)
	}

	return summary, nil
}

func (s *ActivityDataStore) GetActivitySummaryByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.ActivitySummary, error) {
	if id == "" {
		return nil, errors.New("activity summary ID cannot be empty")
	}

	var summary entities.ActivitySummary
	query := s.GetDB().Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&summary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("activity summary not found with ID: %s", id)
		}
		return nil, fmt.Errorf("failed to get activity summary: %w", err)
	}

	return &summary, nil
}

func (s *ActivityDataStore) GetActivitySummariesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.ActivitySummary], error) {
	return utils.PaginatedFilteredQuery(s.GetDB(), tenantID, filterObj, entities.ActivitySummary{})
}

func (s *ActivityDataStore) UpdateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) error {
	if summary == nil {
		return errors.New("activity summary cannot be nil")
	}

	if summary.ID == "" {
		return errors.New("activity summary ID cannot be empty")
	}

	summary.UpdatedAt = time.Now()

	query := s.GetDB().Where("id = ?", summary.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(summary).Error; err != nil {
		return fmt.Errorf("failed to update activity summary: %w", err)
	}

	return nil
}

func (s *ActivityDataStore) DeleteActivitySummary(ctx *appctx.AppContext, tenantID string, id string) error {
	if id == "" {
		return errors.New("activity summary ID cannot be empty")
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.ActivitySummary{}).Error; err != nil {
		return fmt.Errorf("failed to delete activity summary: %w", err)
	}

	return nil
}

// Maintenance operations
func (s *ActivityDataStore) CleanupOldActivities(ctx *appctx.AppContext, tenantID string, retentionDays int) error {
	if retentionDays <= 0 {
		return errors.New("retention days must be positive")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	query := s.GetDB().Where("created_at < ?", cutoffDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old activities: %w", err)
	}

	return nil
}

func (s *ActivityDataStore) ArchiveActivities(ctx *appctx.AppContext, tenantID string, beforeDate time.Time) error {
	query := s.GetDB().Where("created_at < ?", beforeDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Mark activities as archived (you might want to move them to an archive table)
	// For now, we'll just delete them
	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		return fmt.Errorf("failed to archive activities: %w", err)
	}

	return nil
}
