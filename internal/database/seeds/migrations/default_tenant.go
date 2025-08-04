package migrations

import (
	"errors"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	tenant_interfaces "github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"gorm.io/gorm"
)

// DefaultTenantMigrationWorker demonstrates how to create a seed worker
type DefaultTenantMigrationWorker struct {
	db            *gorm.DB
	tenantService tenant_interfaces.TenantServiceInterface
}

// NewDefaultTenantMigrationWorker creates a new example seed worker
func NewDefaultTenantMigrationWorker(db *gorm.DB, tenantService tenant_interfaces.TenantServiceInterface) *DefaultTenantMigrationWorker {
	return &DefaultTenantMigrationWorker{
		db:            db,
		tenantService: tenantService,
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
	if err := e.db.Where("id = ?", config.GlobalTenantID).First(&existingTenant).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_global_tenant", "failed to get global tenant", "default_tenant_migration", nil)
			return diag
		}
	}

	if existingTenant.ID != "" {
		ctx.Log().WithField("tenant_id", existingTenant.ID).Info("Global tenant already exists, skipping migration")
		return diag
	}

	createdTenant, dbErr := e.tenantService.CreateTenant(ctx, &tenant_models.TenantCreateRequest{
		ID:              config.GlobalTenantID,
		Name:            config.GlobalTenantName,
		Description:     "This is the global tenant that is created by default",
		Domain:          "locally.internal",
		ContactEmail:    "root@locally.internal",
		CreateAdminUser: false,
		Metadata: map[string]interface{}{
			"created_by": "default_tenant_migration",
		},
	})
	if dbErr.HasErrors() {
		diag.Append(dbErr)
		return diag
	}

	diag.AddPathEntry("global_tenant_created", "default_tenant_migration", map[string]interface{}{
		"global_tenant_id": createdTenant.ID,
	})

	return diag
}

// Down rolls back the seed
func (e *DefaultTenantMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("default_tenant_migration_down")
	defer diag.Complete()

	existingTenant, dbErr := e.tenantService.GetTenantByIDOrSlug(ctx, config.GlobalTenantID)
	if dbErr.HasErrors() {
		diag.Append(dbErr)
		return diag
	}

	if existingTenant == nil {
		diag.AddError("global_tenant_not_found", "global tenant not found", "default_tenant_migration", nil)
		return diag
	}

	if dbErr := e.tenantService.DeleteTenant(ctx, existingTenant.ID); dbErr.HasErrors() {
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
