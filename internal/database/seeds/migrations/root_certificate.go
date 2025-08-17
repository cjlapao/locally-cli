package migrations

import (
	"errors"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RootCertificateMigrationWorker demonstrates how to create a seed worker
type RootCertificateMigrationWorker struct {
	db                 *gorm.DB
	certificateService interfaces.CertificateServiceInterface
	certificateStore   stores.CertificatesDataStoreInterface
}

// NewRootCertificateMigrationWorker creates a new example seed worker
func NewRootCertificateMigrationWorker(db *gorm.DB, certificateService interfaces.CertificateServiceInterface, certificateStore stores.CertificatesDataStoreInterface) *RootCertificateMigrationWorker {
	return &RootCertificateMigrationWorker{
		db:                 db,
		certificateService: certificateService,
		certificateStore:   certificateStore,
	}
}

// GetName returns the name of this seed
func (e *RootCertificateMigrationWorker) GetName() string {
	return "root-certificate-migration"
}

// GetDescription returns the description of this seed
func (e *RootCertificateMigrationWorker) GetDescription() string {
	return "Root certificate migration"
}

// GetVersion returns the version number
func (e *RootCertificateMigrationWorker) GetVersion() int {
	return 1
}

// Up applies the seed
func (e *RootCertificateMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("root_certificate_migration_up")

	var existingRootCertificate entities.Certificate
	needToCreate := false
	if dbErr := e.db.Where("id = ?", config.GlobalRootCertificateID).First(&existingRootCertificate).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			needToCreate = true
		} else {
			diag.AddError("failed_to_get_root_certificate", "failed to get root certificate", "root_certificate_migration", nil)
			return diag
		}
	}

	if existingRootCertificate.ID != "" {
		ctx.Log().WithField("root_certificate_id", existingRootCertificate.ID).Info("Root certificate already exists, skipping migration")
		return diag
	}

	if needToCreate {
		rootCertificate, certDiag := e.certificateService.GenerateX509RootCertificate(ctx)
		if certDiag.HasErrors() {
			diag.Append(certDiag)
			return diag
		}

		dbRootCertificate := mappers.MapX509CertificateToEntity(rootCertificate)
		dbRootCertificate.ID = config.GlobalRootCertificateID
		dbRootCertificate.TenantID = config.GlobalTenantID
		dbRootCertificate.CreatedBy = config.DefaultSuperUserUserID
		dbRootCertificate.Type = pkg_types.CertificateTypeRoot
		dbRootCertificate.Config.ID = uuid.New().String()
		dbRootCertificate.Config.TenantID = config.GlobalTenantID
		dbRootCertificate.Config.CreatedBy = config.DefaultSuperUserUserID

		dbRootCertificate.CreatedAt = time.Now()
		dbRootCertificate.UpdatedAt = time.Now()

		createdCertificate, createIntermediateCertDiag := e.certificateStore.CreateCertificate(ctx, config.GlobalTenantID, &dbRootCertificate)
		if createIntermediateCertDiag.HasErrors() {
			diag.Append(createIntermediateCertDiag)
			return diag
		}

		diag.AddPathEntry("root_certificate_created", "root_certificate_migration", map[string]interface{}{
			"root_certificate_id": createdCertificate.ID,
		})
	}

	return diag
}

// Down rolls back the seed
func (e *RootCertificateMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("root_certificate_migration_down")
	defer diag.Complete()

	var existingRootCertificate *entities.RootCertificate
	if dbErr := e.db.Where("slug = ?", config.RootCertificateSlug).First(&existingRootCertificate).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return diag
		}
		diag.AddError("failed_to_get_root_certificate", "failed to get root certificate", "root_certificate_migration", nil)
		return diag
	}

	if existingRootCertificate == nil {
		diag.AddError("root_certificate_not_found", "root certificate not found", "root_certificate_migration", nil)
		return diag
	}

	if dbErr := e.db.Delete(&existingRootCertificate).Error; dbErr != nil {
		diag.AddError("failed_to_delete_root_certificate", "failed to delete root certificate", "root_certificate_migration", nil)
		return diag
	}

	diag.AddPathEntry("root_certificate_rolled_back", "root_certificate_migration", map[string]interface{}{
		"root_certificate_id": existingRootCertificate.ID,
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
