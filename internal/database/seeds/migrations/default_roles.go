package migrations

// import (
// 	"errors"
// 	"time"

// 	"github.com/cjlapao/locally-cli/internal/appctx"
// 	"github.com/cjlapao/locally-cli/internal/config"
// 	"github.com/cjlapao/locally-cli/internal/database/entities"
// 	"github.com/cjlapao/locally-cli/internal/mappers"
// 	"github.com/cjlapao/locally-cli/internal/user/defaults"
// 	"github.com/cjlapao/locally-cli/pkg/diagnostics"
// 	"gorm.io/gorm"
// )

// // DefaultRolesMigrationWorker demonstrates how to create a seed worker
// type DefaultRolesMigrationWorker struct {
// 	db     *gorm.DB
// 	config *config.Config
// }

// // NewDefaultRolesMigrationWorker creates a new example seed worker
// func NewDefaultRolesMigrationWorker(db *gorm.DB, config *config.Config) *DefaultRolesMigrationWorker {
// 	return &DefaultRolesMigrationWorker{
// 		db:     db,
// 		config: config,
// 	}
// }

// // GetName returns the name of this seed
// func (e *DefaultRolesMigrationWorker) GetName() string {
// 	return "default-roles-migration"
// }

// // GetDescription returns the description of this seed
// func (e *DefaultRolesMigrationWorker) GetDescription() string {
// 	return "Default roles migration"
// }

// // GetVersion returns the version number
// func (e *DefaultRolesMigrationWorker) GetVersion() int {
// 	return 1
// }

// // Up applies the seed
// func (e *DefaultRolesMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("default_roles_migration_up")

// 	for _, role := range defaults.DefaultRoles {
// 		var existingRole entities.Role
// 		needToCreate := false
// 		if dbErr := e.db.Where("slug = ?", role.Slug).First(&existingRole).Error; dbErr != nil {
// 			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 				needToCreate = true
// 			} else {
// 				diag.AddError("failed_to_get_default_roles", "failed to get default roles", "default_roles_migration", nil)
// 				return diag
// 			}
// 		}

// 		if existingRole.ID != "" {
// 			ctx.Log().WithField("default_role_id", existingRole.ID).Info("Default role already exists, skipping migration")
// 			return diag
// 		}

// 		if needToCreate {
// 			var claims []entities.Claim
// 			if err := e.db.Where("security_level = ?", role.SecurityLevel).Find(&claims).Error; err != nil {
// 				diag.AddError("failed_to_get_default_claims", "failed to get default claims", "default_roles_migration", nil)
// 				return diag
// 			}

// 			dbRole := mappers.MapRoleToEntity(&role)
// 			dbRole.Slug = role.Slug
// 			dbRole.ID = role.ID
// 			dbRole.CreatedAt = time.Now()
// 			dbRole.UpdatedAt = time.Now()
// 			dbRole.SecurityLevel = role.SecurityLevel
// 			dbRole.Claims = claims

// 			if err := e.db.Create(&dbRole).Error; err != nil {
// 				diag.AddError("failed_to_create_default_roles", "failed to create default roles", "default_roles_migration", nil)
// 				return diag
// 			}

// 			diag.AddPathEntry("default_role_created", "default_roles_migration", map[string]interface{}{
// 				"default_role_id": dbRole.ID,
// 			})
// 		}
// 	}

// 	return diag
// }

// // Down rolls back the seed
// func (e *DefaultRolesMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("default_roles_migration_down")
// 	defer diag.Complete()

// 	for _, role := range defaults.DefaultRoles {
// 		var existingRole entities.Role
// 		if dbErr := e.db.Where("slug = ?", role.Slug).First(&existingRole).Error; dbErr != nil {
// 			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 				diag.AddError("default_role_not_found", "default role not found", "default_roles_migration", nil)
// 				return diag
// 			}
// 		}

// 		if existingRole.ID != "" {
// 			if dbErr := e.db.Delete(&existingRole).Error; dbErr != nil {
// 				diag.AddError("failed_to_delete_default_role", "failed to delete default role", "default_roles_migration", nil)
// 				return diag
// 			}

// 			diag.AddPathEntry("default_role_deleted", "default_roles_migration", map[string]interface{}{
// 				"default_role_id": existingRole.ID,
// 			})
// 		}
// 	}

// 	return diag
// }
