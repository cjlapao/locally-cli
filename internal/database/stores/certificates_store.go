package stores

import (
	"errors"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	certificatesDataStoreInstance *CertificatesDataStore
	certificatesDataStoreOnce     sync.Once
)

type CertificatesDataStore struct {
	database.BaseDataStore
}

func GetCertificatesDataStoreInstance() *CertificatesDataStore {
	return certificatesDataStoreInstance
}

func InitializeCertificatesDataStore() *diagnostics.Diagnostics {
	diag := diagnostics.New("initialize_certificates_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing certificates store...")

	certificatesDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "certificates_data_store", nil)
			return
		}

		store := &CertificatesDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running certificates migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Certificates migrations completed")
		}

		certificatesDataStoreInstance = store
	})

	logging.Info("Certificates store initialized successfully")
	return diag
}

func (s *CertificatesDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_certificates_data_store")
	if err := s.GetDB().AutoMigrate(&entities.RootCertificate{}); err != nil {
		diag.AddError("failed_to_migrate_certificates_table", "failed to migrate certificates table", "certificates_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.IntermediateCertificate{}); err != nil {
		diag.AddError("failed_to_migrate_intermediate_certificates_table", "failed to migrate intermediate certificates table", "certificates_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.Certificate{}); err != nil {
		diag.AddError("failed_to_migrate_certificates_table", "failed to migrate certificates table", "certificates_data_store", nil)
		return diag
	}
	return diag
}

func (s *CertificatesDataStore) GetRootCertificates(ctx *appctx.AppContext) ([]entities.RootCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_root_certificates")
	var rootCertificates []entities.RootCertificate
	db := s.GetDB()
	if err := db.Preload("IntermediateCertificates").
		Preload("IntermediateCertificates.Certificates").
		Find(&rootCertificates).Error; err != nil {
		diag.AddError("failed_to_get_root_certificates", "failed to get root certificates", "certificates_data_store", nil)
		return nil, diag
	}

	return rootCertificates, diag
}

func (s *CertificatesDataStore) GetRootCertificate(ctx *appctx.AppContext, id string) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_root_certificate")
	var rootCertificate entities.RootCertificate
	db := s.GetDB()
	if err := db.Where("id = ?", id).
		Preload("IntermediateCertificates").
		Preload("IntermediateCertificates.Certificates").
		First(&rootCertificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_root_certificate", "failed to get root certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return &rootCertificate, diag
}

func (s *CertificatesDataStore) GetRootCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_root_certificate_by_slug")
	var rootCertificate entities.RootCertificate
	db := s.GetDB()
	if err := db.Where("slug = ?", slug).
		Preload("IntermediateCertificates").
		Preload("IntermediateCertificates.Certificates").
		First(&rootCertificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_root_certificate_by_slug", "failed to get root certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}

	return &rootCertificate, diag
}

func (s *CertificatesDataStore) GetIntermediateCertificates(ctx *appctx.AppContext) ([]entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_intermediate_certificates")
	var intermediateCertificates []entities.IntermediateCertificate
	db := s.GetDB()
	if err := db.Preload("Certificates").
		Find(&intermediateCertificates).Error; err != nil {
		diag.AddError("failed_to_get_intermediate_certificates", "failed to get intermediate certificates", "certificates_data_store", nil)
		return nil, diag
	}
	return intermediateCertificates, diag
}

func (s *CertificatesDataStore) GetIntermediateCertificate(ctx *appctx.AppContext, id string) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_intermediate_certificate")
	var intermediateCertificate entities.IntermediateCertificate
	db := s.GetDB()
	if err := db.Where("id = ?", id).
		Preload("Certificates").
		First(&intermediateCertificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_intermediate_certificate", "failed to get intermediate certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return &intermediateCertificate, diag
}

func (s *CertificatesDataStore) GetIntermediateCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_intermediate_certificate_by_slug")
	var intermediateCertificate entities.IntermediateCertificate
	db := s.GetDB()
	if err := db.Where("slug = ?", slug).
		Preload("Certificates").
		First(&intermediateCertificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_intermediate_certificate_by_slug", "failed to get intermediate certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}

	return &intermediateCertificate, diag
}

func (s *CertificatesDataStore) GetCertificates(ctx *appctx.AppContext, rootCertificateID string) ([]entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates")
	var certificates []entities.Certificate
	db := s.GetDB()
	if err := db.Where("root_certificate_id = ?", rootCertificateID).Find(&certificates).Error; err != nil {
		diag.AddError("failed_to_get_certificates", "failed to get certificates", "certificates_data_store", nil)
		return nil, diag
	}
	return certificates, diag
}

func (s *CertificatesDataStore) GetCertificate(ctx *appctx.AppContext, id string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificate")
	var certificate entities.Certificate
	db := s.GetDB()
	if err := db.Where("id = ?", id).First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate", "failed to get certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return &certificate, diag
}

func (s *CertificatesDataStore) GetCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificate_by_slug")
	var certificate entities.Certificate
	db := s.GetDB()
	if err := db.Where("slug = ?", slug).First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate_by_slug", "failed to get certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}

	return &certificate, diag
}

func (s *CertificatesDataStore) CreateRootCertificate(ctx *appctx.AppContext, rootCertificate *entities.RootCertificate) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_root_certificate")
	rootCertificate.ID = uuid.New().String()
	rootCertificate.Slug = config.RootCertificateSlug
	rootCertificate.CreatedAt = time.Now()
	rootCertificate.UpdatedAt = time.Now()

	// checking if the root certificate is already in the database
	existingRootCertificate, err := s.GetRootCertificateBySlug(ctx, rootCertificate.Slug)
	if err != nil {
		diag.AddError("failed_to_get_root_certificate_by_slug", "failed to get root certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}
	if existingRootCertificate != nil {
		diag.AddError("root_certificate_already_exists", "root certificate already exists", "certificates_data_store", nil)
		return nil, diag
	}

	db := s.GetDB()
	if err := db.Create(rootCertificate).Error; err != nil {
		diag.AddError("failed_to_create_root_certificate", "failed to create root certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return rootCertificate, diag
}

func (s *CertificatesDataStore) CreateIntermediateCertificate(ctx *appctx.AppContext, intermediateCertificate *entities.IntermediateCertificate) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_intermediate_certificate")
	intermediateCertificate.ID = uuid.New().String()
	intermediateCertificate.Slug = config.IntermediateCertificateSlug
	intermediateCertificate.CreatedAt = time.Now()
	intermediateCertificate.UpdatedAt = time.Now()

	// checking if the intermediate certificate is already in the database
	existingIntermediateCertificate, err := s.GetIntermediateCertificateBySlug(ctx, intermediateCertificate.Slug)
	if err != nil {
		diag.AddError("failed_to_get_intermediate_certificate_by_slug", "failed to get intermediate certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}
	if existingIntermediateCertificate != nil {
		diag.AddError("intermediate_certificate_already_exists", "intermediate certificate already exists", "certificates_data_store", nil)
		return nil, diag
	}

	db := s.GetDB()
	if err := db.Create(intermediateCertificate).Error; err != nil {
		diag.AddError("failed_to_create_intermediate_certificate", "failed to create intermediate certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return intermediateCertificate, diag
}

func (s *CertificatesDataStore) CreateCertificate(ctx *appctx.AppContext, certificate *entities.Certificate) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_certificate")
	certificate.ID = uuid.New().String()
	certificate.Slug = utils.Slugify(certificate.Name)
	certificate.CreatedAt = time.Now()
	certificate.UpdatedAt = time.Now()
	// checking if the certificate is already in the database
	existingCertificate, err := s.GetCertificateBySlug(ctx, certificate.Slug)
	if err != nil {
		diag.AddError("failed_to_get_certificate_by_slug", "failed to get certificate by slug", "certificates_data_store", nil)
		return nil, diag
	}
	if existingCertificate != nil {
		diag.AddError("certificate_already_exists", "certificate already exists", "certificates_data_store", nil)
		return nil, diag
	}

	db := s.GetDB()
	if err := db.Create(certificate).Error; err != nil {
		diag.AddError("failed_to_create_certificate", "failed to create certificate", "certificates_data_store", nil)
		return nil, diag
	}

	return certificate, diag
}

func (s *CertificatesDataStore) DeleteRootCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_root_certificate")
	db := s.GetDB()
	if err := db.Delete(&entities.RootCertificate{}, id).Error; err != nil {
		diag.AddError("failed_to_delete_root_certificate", "failed to delete root certificate", "certificates_data_store", nil)
		return diag
	}
	return diag
}

func (s *CertificatesDataStore) DeleteIntermediateCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_intermediate_certificate")
	db := s.GetDB()
	if err := db.Delete(&entities.IntermediateCertificate{}, id).Error; err != nil {
		diag.AddError("failed_to_delete_intermediate_certificate", "failed to delete intermediate certificate", "certificates_data_store", nil)
		return diag
	}
	return diag
}

func (s *CertificatesDataStore) DeleteCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_certificate")
	db := s.GetDB()
	if err := db.Delete(&entities.Certificate{}, id).Error; err != nil {
		diag.AddError("failed_to_delete_certificate", "failed to delete certificate", "certificates_data_store", nil)
		return diag
	}
	return diag
}
