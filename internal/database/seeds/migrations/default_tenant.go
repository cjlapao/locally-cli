package migrations

import (
	"errors"
	"fmt"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"gorm.io/gorm"
)

// RootCertificateMigrationWorker demonstrates how to create a seed worker
type DefaultTenantMigrationWorker struct {
	db          *gorm.DB
	tenantStore stores.TenantDataStoreInterface
}

// NewDefaultTenantMigrationWorker creates a new example seed worker
func NewDefaultTenantMigrationWorker(db *gorm.DB, tenantStore stores.TenantDataStoreInterface) *DefaultTenantMigrationWorker {
	return &DefaultTenantMigrationWorker{
		db:          db,
		tenantStore: tenantStore,
	}
}

// GetName returns the name of this seed
func (e *DefaultTenantMigrationWorker) GetName() string {
	return "default-tenant-migration"
}

// GetDescription returns the description of this seed
func (e *DefaultTenantMigrationWorker) GetDescription() string {
	return "Default tenant migration"
}

// GetVersion returns the version number
func (e *DefaultTenantMigrationWorker) GetVersion() int {
	return 1
}

// Up applies the seed
func (e *DefaultTenantMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_tenant_migration_up")

	var existingTenant entities.Tenant
	needToCreate := false
	if _, dbErr := e.tenantStore.GetTenantBySlug(ctx, utils.Slugify(config.GlobalTenantName)); dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			needToCreate = true
		} else {
			diag.AddError("failed_to_get_global_tenant", "failed to get global tenant", "default_tenant_migration", nil)
			return diag
		}
	}

	if existingTenant.ID != "" {
		ctx.Log().WithField("tenant_id", existingTenant.ID).Info("Global tenant already exists, skipping migration")
		return diag
	}

	if needToCreate {
		now := time.Now()
		globalTenant := entities.Tenant{
			Name:         config.GlobalTenantName,
			Status:       "active",
			Description:  "This is the global tenant that is created by default",
			Domain:       "locally.internal",
			ContactEmail: "root@locally.internal",
			ActivatedAt:  &now,
			Require2FA:   false,
		}

		createdTenant, dbErr := e.tenantStore.CreateTenant(ctx, &globalTenant)
		if dbErr != nil {
			diag.AddError("failed_to_create_global_tenant", fmt.Sprintf("failed to create global tenant: %v", dbErr), "default_tenant_migration", map[string]interface{}{
				"error": dbErr.Error(),
			})
			return diag
		}

		diag.AddPathEntry("global_tenant_created", "default_tenant_migration", map[string]interface{}{
			"global_tenant_id": createdTenant.ID,
		})
	}

	return diag
}

// Down rolls back the seed
func (e *DefaultTenantMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_tenant_migration_down")
	defer diag.Complete()

	existingTenant, dbErr := e.tenantStore.GetTenantBySlug(ctx, config.GlobalTenantID)
	if dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return diag
		}
		diag.AddError("failed_to_get_global_tenant", "failed to get global tenant", "default_tenant_migration", map[string]interface{}{
			"error": dbErr.Error(),
		})
		return diag
	}

	if existingTenant == nil {
		diag.AddError("global_tenant_not_found", "global tenant not found", "default_tenant_migration", nil)
		return diag
	}

	if dbErr := e.tenantStore.DeleteTenant(ctx, existingTenant); dbErr != nil {
		diag.AddError("failed_to_delete_global_tenant", "failed to delete global tenant", "default_tenant_migration", nil)
		return diag
	}

	diag.AddPathEntry("global_tenant_rolled_back", "default_tenant_migration", map[string]interface{}{
		"global_tenant_id": existingTenant.ID,
	})

	return diag
}

// Usage example:
// func main() {
//     db := // your database connection
//     seedService := NewSeedService(db)
//
//     // Register the example seed
//     exampleWorker := NewExampleSeedWorker(db)
//     seedService.Register(exampleWorker)
//
//     // Run all seeds
//     ctx := appctx.NewContext(nil)
//     diag := seedService.RunAll(ctx)
//
//     if diag.HasErrors() {
//         fmt.Printf("Seeds failed: %s\n", diag.GetSummary())
//     } else {
//         fmt.Println("All seeds applied successfully")
//     }
// }
