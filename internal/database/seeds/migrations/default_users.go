package migrations

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"gorm.io/gorm"
)

var (
	defaultRoles       = mappers.MapRolesToEntities(auth.DefaultRoles)
	defaultAdminClaims = mappers.MapClaimsToEntities(auth.DefaultRootClaims)

	usersToCreate = []entities.User{}
)

// DefaultUsersMigrationWorker demonstrates how to create a seed worker
type DefaultUsersMigrationWorker struct {
	db          *gorm.DB
	config      *config.Config
	userStore   stores.UserDataStoreInterface
	tenantStore stores.TenantDataStoreInterface
}

// NewDefaultUsersMigrationWorker creates a new example seed worker
func NewDefaultUsersMigrationWorker(db *gorm.DB, config *config.Config, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) *DefaultUsersMigrationWorker {
	return &DefaultUsersMigrationWorker{
		db:          db,
		config:      config,
		userStore:   userStore,
		tenantStore: tenantStore,
	}
}

// GetName returns the name of this seed
func (e *DefaultUsersMigrationWorker) GetName() string {
	return "default-users-migration"
}

// GetDescription returns the description of this seed
func (e *DefaultUsersMigrationWorker) GetDescription() string {
	return "Default users migration"
}

// GetVersion returns the version number
func (e *DefaultUsersMigrationWorker) GetVersion() int {
	return 1
}

// Up applies the seed
func (e *DefaultUsersMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_users_migration_up")
	cfg := config.GetInstance().Get()
	for i, role := range defaultRoles {
		var roleEntity entities.Role
		if dbErr := e.db.Where("slug = ?", role.Slug).First(&roleEntity).Error; dbErr != nil {
			diag.AddError(fmt.Sprintf("failed_to_get_role_%s", role.Name), fmt.Sprintf("failed to get role: %v", dbErr), "role_migration", nil)
			return diag
		}
		defaultRoles[i] = roleEntity

	}

	defaultAdminDBClaims := make([]entities.Claim, len(defaultAdminClaims))
	for i, claim := range defaultAdminClaims {
		var claimEntity entities.Claim
		if dbErr := e.db.Where("slug = ?", claim.Slug).First(&claimEntity).Error; dbErr != nil {
			diag.AddError(fmt.Sprintf("failed_to_get_claim_%s", claim.Slug), fmt.Sprintf("failed to get claim: %v", dbErr), "claim_migration", nil)
			return diag
		}
		defaultAdminDBClaims[i] = claimEntity
	}

	globalTenant, dbErr := e.tenantStore.GetTenantBySlug(ctx, utils.Slugify(config.GlobalTenantName))
	if dbErr != nil {
		diag.AddError("failed_to_get_global_tenant", "failed to get global tenant", "default_users_migration", nil)
		return diag
	}

	var superUserRole entities.Role
	for _, role := range defaultRoles {
		if role.Slug == config.SuperUserRole {
			superUserRole = role
			break
		}
	}

	// appending the root user
	usersToCreate = append(usersToCreate, entities.User{
		TenantID:          globalTenant.ID,
		Username:          cfg.GetString(config.RootUserUsernameKey, "root"),
		Password:          cfg.GetString(config.RootUserPasswordKey, "root"),
		Email:             "root@locally.internal",
		Name:              "Locally Root User",
		Status:            "active",
		TwoFactorEnabled:  false,
		TwoFactorSecret:   "",
		TwoFactorVerified: false,
		Roles:             []entities.Role{superUserRole},
		Claims:            defaultAdminDBClaims,
	})

	for _, user := range usersToCreate {
		var existingUser entities.User
		needToCreate := false
		if dbErr := e.db.Where("username = ?", user.Username).First(&existingUser).Error; dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				needToCreate = true
			} else {
				diag.AddError(fmt.Sprintf("failed_to_get_user_%s", user.Username), "failed to get user", "user_migration", nil)
				return diag
			}
		}

		if existingUser.ID != "" {
			ctx.Log().WithField("user_id", existingUser.ID).Info("User already exists, skipping migration")
			return diag
		}

		if needToCreate {
			createdUser, err := e.userStore.CreateUser(ctx, globalTenant.ID, &user)
			if err != nil {
				diag.AddError(fmt.Sprintf("failed_to_create_user_%s", user.Username), fmt.Sprintf("failed to create user: %v", err), "user_migration", nil)
				return diag
			}

			user.ID = createdUser.ID
			diag.AddPathEntry(fmt.Sprintf("user_%s_created", createdUser.Username), "user_migration", map[string]interface{}{
				"user_id": createdUser.ID,
			})
		}
		// if it is the root user we need to update the tenant owner
		if user.Username == cfg.GetString(config.RootUserUsernameKey, "root") {
			tenant, dbErr := e.tenantStore.GetTenantBySlug(ctx, utils.Slugify(config.GlobalTenantName))
			if dbErr != nil {
				diag.AddError("failed_to_get_global_tenant", "failed to get global tenant", "default_users_migration", nil)
				return diag
			}
			tenant.OwnerID = user.ID
			if dbErr := e.tenantStore.UpdateTenant(ctx, tenant); dbErr != nil {
				diag.AddError("failed_to_update_global_tenant", "failed to update global tenant", "default_users_migration", nil)
				return diag
			}
		}
	}

	return diag
}

// Down rolls back the seed
func (e *DefaultUsersMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_users_migration_down")
	defer diag.Complete()

	for _, user := range usersToCreate {
		var existingUser *entities.User
		if dbErr := e.db.Where("slug = ?", user.Slug).First(&existingUser).Error; dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				return diag
			}
			diag.AddError(fmt.Sprintf("failed_to_get_user_%s", user.Username), "failed to get user", "user_migration", nil)
			return diag
		}

		if existingUser == nil {
			diag.AddError(fmt.Sprintf("user_%s_not_found", user.Username), "user not found", "user_migration", nil)
			return diag
		}

		if dbErr := e.db.Delete(&existingUser).Error; dbErr != nil {
			diag.AddError(fmt.Sprintf("failed_to_delete_user_%s", user.Username), "failed to delete user", "user_migration", nil)
			return diag
		}

		diag.AddPathEntry(fmt.Sprintf("user_%s_rolled_back", user.Username), "user_migration", map[string]interface{}{
			"user_id": existingUser.ID,
		})
	}

	return diag
}
