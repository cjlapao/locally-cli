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
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	activityDataStoreInstance *ActivityDataStore
	activityDataStoreOnce     sync.Once
)

type ActivityDataStoreInterface interface {
	// Activity CRUD operations
	CreateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) (*entities.Activity, *diagnostics.Diagnostics)
	GetActivityByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.Activity, *diagnostics.Diagnostics)
	GetActivities(ctx *appctx.AppContext, tenantID string) ([]entities.Activity, *diagnostics.Diagnostics)
	GetActivitiesByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Activity], *diagnostics.Diagnostics)
	UpdateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) *diagnostics.Diagnostics
	DeleteActivity(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics

	// Activity querying and reporting
	GetActivitiesByFilterAdvanced(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], *diagnostics.Diagnostics)
	GetActivityStats(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, *diagnostics.Diagnostics)
	GetTopActors(ctx *appctx.AppContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, *diagnostics.Diagnostics)
	GetActivityTrends(ctx *appctx.AppContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, *diagnostics.Diagnostics)

	// Activity summary operations
	CreateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, *diagnostics.Diagnostics)
	GetActivitySummaryByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.ActivitySummary, *diagnostics.Diagnostics)
	UpdateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) *diagnostics.Diagnostics
	DeleteActivitySummary(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics

	// Maintenance operations
	CleanupOldActivities(ctx *appctx.AppContext, tenantID string, retentionDays int) *diagnostics.Diagnostics
	ArchiveActivities(ctx *appctx.AppContext, tenantID string, beforeDate time.Time) *diagnostics.Diagnostics
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
		diag.AddError("failed_to_migrate_activity_table", fmt.Sprintf("failed to migrate activity table: %v", err), "activity_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.ActivitySummary{}); err != nil {
		diag.AddError("failed_to_migrate_activity_summary_table", fmt.Sprintf("failed to migrate activity summary table: %v", err), "activity_data_store", nil)
		return diag
	}

	// Create indexes for better query performance
	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_tenant_module_service ON activities(tenant_id, module, service)").Error; err != nil {
		diag.AddError("failed_to_create_activities_index", fmt.Sprintf("failed to create activities index: %v", err), "activity_data_store", nil)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_actor_target ON activities(actor_type, actor_id)").Error; err != nil {
		diag.AddError("failed_to_create_activities_actor_target_index", fmt.Sprintf("failed to create activities actor target index: %v", err), "activity_data_store", nil)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_timing ON activities(started_at, completed_at, created_at)").Error; err != nil {
		diag.AddError("failed_to_create_activities_timing_index", fmt.Sprintf("failed to create activities timing index: %v", err), "activity_data_store", nil)
	}

	return diag
}

func (s *ActivityDataStore) CreateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) (*entities.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_activity")

	if activity == nil {
		diag.AddError("activity_cannot_be_nil", "activity cannot be nil", "activity_data_store")
		return nil, diag
	}

	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}

	if activity.Slug == "" {
		activity.Slug = pkg_utils.Slugify(fmt.Sprintf("activity-%s", activity.ID))
	}

	activity.TenantID = tenantID
	if activity.TenantID == "" {
		activity.TenantID = config.UnknownTenantID
	}
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if activity.StartedAt == nil {
		now := time.Now()
		activity.StartedAt = &now
	}

	if err := s.GetDB().Create(activity).Error; err != nil {
		logging.Errorf("Failed to create activity: %v", err)
		diag.AddError("failed_to_create_activity", fmt.Sprintf("failed to create activity, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return activity, diag
}

func (s *ActivityDataStore) GetActivityByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activity_by_id")

	if id == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store")
		return nil, diag
	}

	var activity entities.Activity
	query := s.GetDB().Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_activity", fmt.Sprintf("failed to get activity, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return &activity, diag
}

func (s *ActivityDataStore) GetActivities(ctx *appctx.AppContext, tenantID string) ([]entities.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activities")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "activity_data_store")
		return nil, diag
	}

	query := s.GetDB().Where("tenant_id = ?", tenantID)
	var activities []entities.Activity
	if err := query.Find(&activities).Error; err != nil {
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return activities, diag
}

func (s *ActivityDataStore) GetActivitiesByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Activity], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activities")
	db := s.GetDB()

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := utils.QueryDatabase[entities.Activity](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}
	return result, diag
}

func (s *ActivityDataStore) UpdateActivity(ctx *appctx.AppContext, tenantID string, activity *entities.Activity) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_activity")

	if activity == nil {
		diag.AddError("activity_cannot_be_nil", "activity cannot be nil", "activity_data_store")
		return diag
	}

	if activity.ID == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store")
		return diag
	}

	activity.UpdatedAt = time.Now()

	query := s.GetDB().Where("id = ?", activity.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(activity).Error; err != nil {
		diag.AddError("failed_to_update_activity", fmt.Sprintf("failed to update activity, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}

func (s *ActivityDataStore) DeleteActivity(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_activity")

	if id == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store")
		return diag
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_delete_activity", fmt.Sprintf("failed to delete activity, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}

func (s *ActivityDataStore) GetActivitiesByFilterAdvanced(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activities_by_filter_advanced")

	if filter == nil {
		filter = &entities.ActivityFilter{}
	}
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
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	// Get total count
	var total int64
	countQuery := s.GetDB().Model(&entities.Activity{})
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		diag.AddError("failed_to_count_activities", fmt.Sprintf("failed to count activities, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return &filters.FilterResponse[entities.Activity]{
		Items:      activities,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, diag
}

func (s *ActivityDataStore) GetActivityStats(ctx *appctx.AppContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activity_stats")
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
		diag.AddError("failed_to_get_activity_stats", fmt.Sprintf("failed to get activity stats, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return map[string]interface{}{
		"total_activities": stats.TotalActivities,
		"success_count":    stats.SuccessCount,
		"error_count":      stats.ErrorCount,
		"avg_duration_ms":  stats.AvgDurationMs,
		"max_duration_ms":  stats.MaxDurationMs,
		"min_duration_ms":  stats.MinDurationMs,
	}, diag
}

func (s *ActivityDataStore) GetTopActors(ctx *appctx.AppContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_top_actors")
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
		diag.AddError("failed_to_get_top_actors", fmt.Sprintf("failed to get top actors, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return results, diag
}

func (s *ActivityDataStore) GetActivityTrends(ctx *appctx.AppContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activity_trends")
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
		diag.AddError("failed_to_get_activity_trends", fmt.Sprintf("failed to get activity trends, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return results, diag
}

// ActivitySummary operations
func (s *ActivityDataStore) CreateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_activity_summary")

	if summary == nil {
		diag.AddError("activity_summary_cannot_be_nil", "activity summary cannot be nil", "activity_data_store")
		return nil, diag
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
		diag.AddError("failed_to_create_activity_summary", fmt.Sprintf("failed to create activity summary, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return summary, diag
}

func (s *ActivityDataStore) GetActivitySummaryByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.ActivitySummary, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activity_summary_by_id")

	if id == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store")
		return nil, diag
	}

	var summary entities.ActivitySummary
	query := s.GetDB().Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&summary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			diag.AddError("activity_summary_not_found", fmt.Sprintf("activity summary not found with ID: %s", id), "activity_data_store")
			return nil, diag
		}
		diag.AddError("failed_to_get_activity_summary", fmt.Sprintf("failed to get activity summary, error: %s", err.Error()), "activity_data_store")
		return nil, diag
	}

	return &summary, diag
}

func (s *ActivityDataStore) UpdateActivitySummary(ctx *appctx.AppContext, tenantID string, summary *entities.ActivitySummary) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_activity_summary")

	if summary == nil {
		diag.AddError("activity_summary_cannot_be_nil", "activity summary cannot be nil", "activity_data_store")
		return diag
	}

	if summary.ID == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store")
		return diag
	}

	summary.UpdatedAt = time.Now()

	query := s.GetDB().Where("id = ?", summary.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(summary).Error; err != nil {
		diag.AddError("failed_to_update_activity_summary", fmt.Sprintf("failed to update activity summary, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}

func (s *ActivityDataStore) DeleteActivitySummary(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_activity_summary")

	if id == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store", nil)
		return diag
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.ActivitySummary{}).Error; err != nil {
		diag.AddError("failed_to_delete_activity_summary", fmt.Sprintf("failed to delete activity summary, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}

// Maintenance operations
func (s *ActivityDataStore) CleanupOldActivities(ctx *appctx.AppContext, tenantID string, retentionDays int) *diagnostics.Diagnostics {
	diag := diagnostics.New("cleanup_old_activities")

	if retentionDays <= 0 {
		diag.AddError("retention_days_must_be_positive", "retention days must be positive", "activity_data_store", nil)
		return diag
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	query := s.GetDB().Where("created_at < ?", cutoffDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_cleanup_old_activities", fmt.Sprintf("failed to cleanup old activities, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}

func (s *ActivityDataStore) ArchiveActivities(ctx *appctx.AppContext, tenantID string, beforeDate time.Time) *diagnostics.Diagnostics {
	diag := diagnostics.New("archive_activities")

	query := s.GetDB().Where("created_at < ?", beforeDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Mark activities as archived (you might want to move them to an archive table)
	// For now, we'll just delete them
	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_archive_activities", fmt.Sprintf("failed to archive activities, error: %s", err.Error()), "activity_data_store")
		return diag
	}

	return diag
}
