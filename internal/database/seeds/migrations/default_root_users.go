package migrations

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	user_interfaces "github.com/cjlapao/locally-cli/internal/user/interfaces"
	user_models "github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"gorm.io/gorm"
)

// DefaultRootUsersMigrationWorker demonstrates how to create a seed worker
type DefaultRootUsersMigrationWorker struct {
	db            *gorm.DB
	config        *config.Config
	systemService system_interfaces.SystemServiceInterface
	userService   user_interfaces.UserServiceInterface
}

// NewDefaultUsersMigrationWorker creates a new example seed worker
func NewDefaultUsersMigrationWorker(db *gorm.DB, config *config.Config, systemService system_interfaces.SystemServiceInterface, userService user_interfaces.UserServiceInterface) *DefaultRootUsersMigrationWorker {
	return &DefaultRootUsersMigrationWorker{
		db:            db,
		config:        config,
		systemService: systemService,
		userService:   userService,
	}
}

// GetName returns the name of this seed
func (e *DefaultRootUsersMigrationWorker) GetName() string {
	return "default-root-users-migration"
}

// GetDescription returns the description of this seed
func (e *DefaultRootUsersMigrationWorker) GetDescription() string {
	return "Default root users migration"
}

// GetVersion returns the version number
func (e *DefaultRootUsersMigrationWorker) GetVersion() int {
	return 1
}

// Up applies the seed
func (e *DefaultRootUsersMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_root_users_migration_up")
	cfg := config.GetInstance().Get()
	username := cfg.GetString(config.RootUserUsernameKey, "root")
	role, err := e.systemService.GetRoleBySecurityLevel(models.SecurityLevelSuperUser)
	if err != nil {
		diag.AddError("failed_to_get_super_user_role", "failed to get super user role", "default_root_users_migration", nil)
		return diag
	}

	var existingUser entities.User
	if err := e.db.Where("username = ?", username).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user", "failed to get user", "default_root_users_migration", nil)
			return diag
		}
	}

	if existingUser.ID != "" {
		ctx.Log().WithField("user_id", existingUser.ID).Info("User already exists, skipping migration")
		return diag
	}

	createdUser, createDiag := e.userService.CreateUser(ctx, config.GlobalTenantID, role.Slug, &user_models.CreateUserRequest{
		ID:       config.DefaultSuperUserUserID,
		Username: username,
		Password: cfg.GetString(config.RootUserPasswordKey, "root"),
		Email:    "root@locally.internal",
		Name:     "Locally Root User",
		Role:     role.Slug,
	})
	if createDiag.HasErrors() {
		diag.Append(createDiag)
		return diag
	}

	diag.AddPathEntry(fmt.Sprintf("user_%s_created", username), "user_migration", map[string]interface{}{
		"user_id": createdUser.ID,
	})

	return diag
}

// Down rolls back the seed
func (e *DefaultRootUsersMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_users_migration_down")

	cfg := config.GetInstance().Get()
	superUserUsername := cfg.GetString(config.RootUserUsernameKey, "root")
	defer diag.Complete()

	var existingUser *entities.User
	if dbErr := e.db.Where("username = ?", superUserUsername).First(&existingUser).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return diag
		}
		diag.AddError(fmt.Sprintf("failed_to_get_user_%s", superUserUsername), "failed to get user", "user_migration", nil)
		return diag
	}

	if existingUser == nil {
		diag.AddError(fmt.Sprintf("user_%s_not_found", superUserUsername), "user not found", "user_migration", nil)
		return diag
	}

	if dbErr := e.db.Delete(&existingUser).Error; dbErr != nil {
		diag.AddError(fmt.Sprintf("failed_to_delete_user_%s", superUserUsername), "failed to delete user", "user_migration", nil)
		return diag
	}

	diag.AddPathEntry(fmt.Sprintf("user_%s_rolled_back", superUserUsername), "user_migration", map[string]interface{}{
		"user_id": existingUser.ID,
	})

	return diag
}
