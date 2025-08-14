package migrations

// // IntermediateCertificateMigrationWorker demonstrates how to create a seed worker
// type IntermediateCertificateMigrationWorker struct {
// 	db                 *gorm.DB
// 	certificateService *certificates.CertificateService
// }

// // NewIntermediateCertificateMigrationWorker creates a new example seed worker
// func NewIntermediateCertificateMigrationWorker(db *gorm.DB, certificateService *certificates.CertificateService) *IntermediateCertificateMigrationWorker {
// 	return &IntermediateCertificateMigrationWorker{
// 		db:                 db,
// 		certificateService: certificateService,
// 	}
// }

// // GetName returns the name of this seed
// func (e *IntermediateCertificateMigrationWorker) GetName() string {
// 	return "intermediate-certificate-migration"
// }

// // GetDescription returns the description of this seed
// func (e *IntermediateCertificateMigrationWorker) GetDescription() string {
// 	return "Intermediate certificate migration"
// }

// // GetVersion returns the version number
// func (e *IntermediateCertificateMigrationWorker) GetVersion() int {
// 	return 1
// }

// // Up applies the seed
// func (e *IntermediateCertificateMigrationWorker) Up(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("intermediate_certificate_migration_up")

// 	var existingRootCertificate entities.RootCertificate
// 	var existingIntermediateCertificate entities.IntermediateCertificate
// 	needToCreate := false

// 	// Getting the root certificate
// 	if dbErr := e.db.Where("slug = ?", config.RootCertificateSlug).First(&existingRootCertificate).Error; dbErr != nil {
// 		diag.AddError("intermediate_certificate_up", "failed to get root certificate", "intermediate_certificate_migration", nil)
// 		return diag
// 	}

// 	if dbErr := e.db.Where("slug = ?", config.IntermediateCertificateSlug).First(&existingIntermediateCertificate).Error; dbErr != nil {
// 		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 			needToCreate = true
// 		} else {
// 			diag.AddError("failed_to_get_intermediate_certificate", "failed to get intermediate certificate", "intermediate_certificate_migration", nil)
// 			return diag
// 		}
// 	}

// 	if existingIntermediateCertificate.ID != "" {
// 		ctx.Log().WithField("intermediate_certificate_id", existingIntermediateCertificate.ID).Info("Intermediate certificate already exists, skipping migration")
// 		return diag
// 	}

// 	if needToCreate {
// 		dtoRootCertificate := mappers.MapRootCertificateToDto(existingRootCertificate)
// 		intermediateCertificate, certDiag := e.certificateService.GenerateIntermediateCertificate(ctx, &dtoRootCertificate)
// 		if certDiag.HasErrors() {
// 			diag.Append(certDiag)
// 			return diag
// 		}
// 		dbIntermediateCertificate := mappers.MapIntermediateCertificateToEntity(*intermediateCertificate)
// 		dbIntermediateCertificate.Slug = config.IntermediateCertificateSlug
// 		dbIntermediateCertificate.RootCertificateID = existingRootCertificate.ID
// 		dbIntermediateCertificate.ID = uuid.New().String()
// 		dbIntermediateCertificate.CreatedAt = time.Now()
// 		dbIntermediateCertificate.UpdatedAt = time.Now()

// 		if dbErr := e.db.Create(&dbIntermediateCertificate).Error; dbErr != nil {
// 			diag.AddError("failed_to_create_intermediate_certificate", fmt.Sprintf("failed to create intermediate certificate: %v", dbErr), "intermediate_certificate_migration", nil)
// 			return diag
// 		}

// 		diag.AddPathEntry("intermediate_certificate_created", "intermediate_certificate_migration", map[string]interface{}{
// 			"intermediate_certificate_id": dbIntermediateCertificate.ID,
// 		})
// 	}

// 	return diag
// }

// // Down rolls back the seed
// func (e *IntermediateCertificateMigrationWorker) Down(ctx *appctx.AppContext) *diagnostics.Diagnostics {
// 	diag := diagnostics.New("intermediate_certificate_migration_down")
// 	defer diag.Complete()

// 	var existingRootCertificate *entities.RootCertificate
// 	if dbErr := e.db.Where("slug = ?", "root").First(&existingRootCertificate).Error; dbErr != nil {
// 		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
// 			return diag
// 		}
// 		diag.AddError("failed_to_get_root_certificate", "failed to get root certificate", "root_certificate_migration", nil)
// 		return diag
// 	}

// 	if existingRootCertificate == nil {
// 		diag.AddError("root_certificate_not_found", "root certificate not found", "root_certificate_migration", nil)
// 		return diag
// 	}

// 	if dbErr := e.db.Delete(&existingRootCertificate).Error; dbErr != nil {
// 		diag.AddError("failed_to_delete_root_certificate", "failed to delete root certificate", "root_certificate_migration", nil)
// 		return diag
// 	}

// 	diag.AddPathEntry("root_certificate_rolled_back", "root_certificate_migration", map[string]interface{}{
// 		"root_certificate_id": existingRootCertificate.ID,
// 	})

// 	return diag
// }

// // Usage example:
// // func main() {
// //     db := // your database connection
// //     seedService := NewSeedService(db)
// //
// //     // Register the example seed
// //     exampleWorker := NewExampleSeedWorker(db)
// //     seedService.Register(exampleWorker)
// //
// //     // Run all seeds
// //     ctx := appctx.NewContext(nil)
// //     diag := seedService.RunAll(ctx)
// //
// //     if diag.HasErrors() {
// //         fmt.Printf("Seeds failed: %s\n", diag.GetSummary())
// //     } else {
// //         fmt.Println("All seeds applied successfully")
// //     }
// // }
