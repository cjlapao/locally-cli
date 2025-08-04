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
// 	"github.com/google/uuid"
// 	"gorm.io/gorm"
// )

// // DefaultClaimsMigrationWorker demonstrates how to create a seed worker
// type DefaultClaimsMigrationWorker struct {
// 	db     *gorm.DB
// 	config *config.Config
// }

// // NewDefaultClaimsMigrationWorker creates a new example seed worker
// func NewDefaultClaimsMigrationWorker(db *gorm.DB, config *config.Config) *DefaultClaimsMigrationWorker {
// 	return &DefaultClaimsMigrationWorker{
// 		db:     db,
// 		config: config,
// 	}
// }

// // GetName returns the name of this seed
// func (e *DefaultClaimsMigrationWorker) GetName() string {
// 	return "default-claims-migration"
// }

// // GetDescription returns the description of this seed
// func (e *DefaultClaimsMigrationWorker) GetDescription() string {
// 	return "Default claims migration"
// }

// // GetVersion returns the version number
// func (e *DefaultClaimsMigrationWorker) GetVersion() int {
// 	return 1
// }

// // Up applies the seed
// func (e *DefaultClaimsMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("default_claims_migration_up")

// 	for _, claim := range defaults.DefaultClaims {
// 		var existingClaim entities.Claim
// 		needToCreate := false
// 		if dbErr := e.db.Where("slug = ?", claim.Slug).First(&existingClaim).Error; dbErr != nil {
// 			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 				needToCreate = true
// 			} else {
// 				diag.AddError("failed_to_get_default_claims", "failed to get default claims", "default_claims_migration", nil)
// 				return diag
// 			}
// 		}

// 		if existingClaim.ID != "" {
// 			ctx.Log().WithField("default_claim_id", existingClaim.ID).Info("Default claim already exists, skipping migration")
// 			return diag
// 		}

// 		if needToCreate {
// 			dbClaim := mappers.MapClaimToEntity(&claim)
// 			dbClaim.Slug = claim.GetSlug()
// 			dbClaim.ID = uuid.New().String()
// 			dbClaim.SecurityLevel = claim.SecurityLevel
// 			dbClaim.CreatedAt = time.Now()
// 			dbClaim.UpdatedAt = time.Now()

// 			if err := e.db.Create(&dbClaim).Error; err != nil {
// 				diag.AddError("failed_to_create_default_claims", "failed to create default claims", "default_claims_migration", nil)
// 				return diag
// 			}

// 			diag.AddPathEntry("default_claim_created", "default_claims_migration", map[string]interface{}{
// 				"default_claim_id": dbClaim.ID,
// 			})
// 		}
// 	}

// 	return diag
// }

// // Down rolls back the seed
// func (e *DefaultClaimsMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("default_claims_migration_down")
// 	defer diag.Complete()

// 	for _, claim := range defaults.DefaultClaims {
// 		var existingClaim entities.Claim
// 		if dbErr := e.db.Where("slug = ?", claim.Slug).First(&existingClaim).Error; dbErr != nil {
// 			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 				diag.AddError("default_claim_not_found", "default claim not found", "default_claims_migration", nil)
// 				return diag
// 			}
// 		}

// 		if existingClaim.ID != "" {
// 			if dbErr := e.db.Delete(&existingClaim).Error; dbErr != nil {
// 				diag.AddError("failed_to_delete_default_claim", "failed to delete default claim", "default_claims_migration", nil)
// 				return diag
// 			}

// 			diag.AddPathEntry("default_claim_deleted", "default_claims_migration", map[string]interface{}{
// 				"default_claim_id": existingClaim.ID,
// 			})
// 		}
// 	}

// 	return diag
// }
