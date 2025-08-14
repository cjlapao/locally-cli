package stores

import (
	"errors"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	certificatesDataStoreInstance *CertificatesDataStore
	certificatesDataStoreOnce     sync.Once
)

type CertificatesDataStoreInterface interface {
	GetRootCertificate(ctx *appctx.AppContext, id string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetTenantIntermediateCertificate(ctx *appctx.AppContext, tenantID string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetCertificates(ctx *appctx.AppContext, tenantID string, query_builder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics)
	CreateCertificate(ctx *appctx.AppContext, tenantID string, certificate *entities.Certificate) (*entities.Certificate, *diagnostics.Diagnostics)
	DeleteCertificate(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	// new ones
	GetCertificateBy(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetCertificatesByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType) ([]entities.Certificate, *diagnostics.Diagnostics)
}

type CertificatesDataStore struct {
	database.BaseDataStore
}

func GetCertificatesDataStoreInstance() CertificatesDataStoreInterface {
	return certificatesDataStoreInstance
}

func InitializeCertificatesDataStore() (CertificatesDataStoreInterface, *diagnostics.Diagnostics) {
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
	return certificatesDataStoreInstance, diag
}

func (s *CertificatesDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_certificates_data_store")
	if err := s.GetDB().AutoMigrate(&entities.CertificateConfig{}); err != nil {
		diag.AddError("failed_to_migrate_certificate_configs_table", "failed to migrate certificate configs table", "certificates_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.Certificate{}); err != nil {
		diag.AddError("failed_to_migrate_certificates_table", "failed to migrate certificates table", "certificates_data_store", nil)
		return diag
	}

	return diag
}

func (s *CertificatesDataStore) GetCertificateBy(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificate_by")
	var certificate entities.Certificate
	db := s.GetDB()
	if err := db.Preload("Config").
		Where("tenant_id = ? AND type = ? AND (slug = ? OR id = ? or name = ?)", tenantId, certType, slugOrId, slugOrId, slugOrId).
		Order("created_at desc").
		First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate_by", "failed to get certificate by", "certificates_data_store", nil)
		return nil, diag
	}
	return &certificate, diag
}

func (s *CertificatesDataStore) GetCertificatesByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType) ([]entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates_by_type")
	db := s.GetDB()
	var certificates []entities.Certificate
	if err := db.Where("tenant_id = ? AND type = ?", tenantId, certType).Find(&certificates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate_by_type", "failed to get certificate by type", "certificates_data_store", nil)
		return nil, diag
	}
	return certificates, diag
}

func (s *CertificatesDataStore) GetRootCertificate(ctx *appctx.AppContext) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_root_certificate")
	tenantID := config.GlobalTenantID
	rootCertificateID := config.GlobalRootCertificateID
	certificate, certificateDiag := s.GetCertificateBy(ctx, tenantID, pkg_types.CertificateTypeRoot, rootCertificateID)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate == nil {
		diag.AddError("root_certificate_not_found", "root certificate not found", "certificates_data_store", nil)
		return nil, diag
	}

	return certificate, diag
}

func (s *CertificatesDataStore) GetTenantIntermediateCertificate(ctx *appctx.AppContext, tenantID string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_tenant_intermediate_certificate")
	certificate, certificateDiag := s.GetCertificateBy(ctx, tenantID, pkg_types.CertificateTypeIntermediate, tenantID)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", "certificates_data_store", nil)
		return nil, diag
	}
	return certificate, diag
}

func (s *CertificatesDataStore) GetCertificates(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates")
	var certificates []entities.Certificate
	db := s.GetDB()
	db.Preload("Config")

	if queryBuilder != nil {
		db = queryBuilder.Apply(db)
	}

	if err := db.Find(&certificates).Error; err != nil {
		diag.AddError("failed_to_get_certificates", "failed to get certificates", "certificates_data_store", nil)
		return nil, diag
	}

	return certificates, diag
}

func (s *CertificatesDataStore) CreateCertificate(ctx *appctx.AppContext, tenantID string, certificate *entities.Certificate) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_certificate")
	if certificate.ID == "" {
		certificate.ID = uuid.New().String()
	}
	if certificate.TenantID == "" {
		certificate.TenantID = tenantID
	}
	certificate.Slug = utils.Slugify(certificate.Name)
	certificate.CreatedAt = time.Now()
	certificate.UpdatedAt = time.Now()

	// checking if the certificate is already in the database
	existingCertificate, err := s.GetCertificateBy(ctx, tenantID, certificate.Type, certificate.Slug)
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

func (s *CertificatesDataStore) DeleteCertificate(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_certificate")
	db := s.GetDB()
	if err := db.
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&entities.Certificate{}).Error; err != nil {
		diag.AddError("failed_to_delete_certificate", "failed to delete certificate", "certificates_data_store", nil)
		return diag
	}
	return diag
}
