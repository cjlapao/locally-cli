package stores

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	db_utils "github.com/cjlapao/locally-cli/internal/database/utils"
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
	GetRootCertificate(ctx *appctx.AppContext) (*entities.Certificate, *diagnostics.Diagnostics)
	GetTenantIntermediateCertificate(ctx *appctx.AppContext, tenantID string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetCertificates(ctx *appctx.AppContext, tenantID string) ([]entities.Certificate, *diagnostics.Diagnostics)
	GetCertificatesByQuery(ctx *appctx.AppContext, tenantID string, query_builder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics)
	CreateCertificate(ctx *appctx.AppContext, tenantID string, certificate *entities.Certificate) (*entities.Certificate, *diagnostics.Diagnostics)
	DeleteCertificate(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	// new ones
	GetCertificateBy(ctx *appctx.AppContext, tenantId string, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetCertificateByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics)
	GetCertificatesByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, query_builder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics)
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
	if err := s.GetDB().AutoMigrate(&entities.Certificate{}); err != nil {
		diag.AddError("failed_to_migrate_certificates_table", fmt.Sprintf("failed to migrate certificates table: %v", err), "certificates_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.CertificateConfig{}); err != nil {
		diag.AddError("failed_to_migrate_certificate_configs_table", fmt.Sprintf("failed to migrate certificate configs table: %v", err), "certificates_data_store", nil)
		return diag
	}
	return diag
}

func (s *CertificatesDataStore) GetCertificateBy(ctx *appctx.AppContext, tenantId string, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificate_by")
	var certificate entities.Certificate
	db := s.GetDB()
	if err := db.Preload("Config").
		Where("tenant_id = ? AND (slug = ? OR id = ? or name = ?)", tenantId, slugOrId, slugOrId, slugOrId).
		Order("created_at desc").
		First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate_by", fmt.Sprintf("failed to get certificate by: %v", err), "certificates_data_store", nil)
		return nil, diag
	}
	return &certificate, diag
}

func (s *CertificatesDataStore) GetCertificateByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, slugOrId string) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificate_by_type")
	var certificate entities.Certificate
	db := s.GetDB()
	if err := db.Preload("Config").
		Where("tenant_id = ? AND type = ? AND (slug = ? OR id = ? OR name = ?)", tenantId, certType, slugOrId, slugOrId, slugOrId).
		Order("created_at desc").
		First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_certificate_by", fmt.Sprintf("failed to get certificate by: %v", err), "certificates_data_store", nil)
		return nil, diag
	}
	return &certificate, diag
}

func (s *CertificatesDataStore) GetCertificatesByType(ctx *appctx.AppContext, tenantID string, certType pkg_types.CertificateType, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates_by_type")
	db := s.GetDB()
	db = db.Preload("Config")
	db = db.Where("type = ?", certType)

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := db_utils.QueryDatabase[entities.Certificate](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_certificates", fmt.Sprintf("failed to get certificates: %v", err), "certificates_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *CertificatesDataStore) GetRootCertificate(ctx *appctx.AppContext) (*entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_root_certificate")
	tenantID := config.GlobalTenantID
	rootCertificateID := config.GlobalRootCertificateID
	certificate, certificateDiag := s.GetCertificateBy(ctx, tenantID, rootCertificateID)
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
	db := s.GetDB()
	db = db.Preload("Config")
	db = db.Where("tenant_id = ? AND type = ?", tenantID, pkg_types.CertificateTypeIntermediate)
	var certificate *entities.Certificate
	if err := db.First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_intermediate_certificate", fmt.Sprintf("failed to get intermediate certificate: %v", err), "certificates_data_store", nil)
		return nil, diag
	}
	if certificate == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", "certificates_data_store", nil)
		return nil, diag
	}
	return certificate, diag
}

func (s *CertificatesDataStore) GetCertificates(ctx *appctx.AppContext, tenantID string) ([]entities.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "certificates_data_store", nil)
		return nil, diag
	}

	db := s.GetDB()
	db = db.Preload("Config")
	db = db.Where("tenant_id = ?", tenantID)

	var certificates []entities.Certificate
	if err := db.Find(&certificates).Error; err != nil {
		diag.AddError("failed_to_get_certificates", fmt.Sprintf("failed to get certificates: %v", err), "certificates_data_store", nil)
		return nil, diag
	}

	return certificates, diag
}

func (s *CertificatesDataStore) GetCertificatesByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_certificates")
	db := s.GetDB()
	db = db.Preload("Config")

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := db_utils.QueryDatabase[entities.Certificate](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_certificates", fmt.Sprintf("failed to get certificates: %v", err), "certificates_data_store", nil)
		return nil, diag
	}

	return result, diag
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
	if certificate.Config.ID == "" {
		certificate.Config.ID = uuid.New().String()
	}
	certificate.Config.TenantID = tenantID
	certificate.Config.CreatedAt = time.Now()
	certificate.Config.UpdatedAt = time.Now()
	certificate.Config.CertificateID = certificate.ID
	certificate.Config.Slug = utils.Slugify(certificate.Name + "-config")
	certificate.Config.CreatedBy = certificate.CreatedBy
	if certificate.Config.FQDNs == nil {
		certificate.Config.FQDNs = []string{}
	}
	if certificate.Config.IpAddresses == nil {
		certificate.Config.IpAddresses = []string{}
	}

	// checking if the certificate is already in the database
	existingCertificate, existCertificateDiag := s.GetCertificateByType(ctx, tenantID, certificate.Type, certificate.ID)
	if existCertificateDiag.HasErrors() {
		diag.Append(existCertificateDiag)
		return nil, diag
	}
	if existingCertificate != nil {
		diag.AddError("certificate_already_exists", "certificate already exists", "certificates_data_store", map[string]interface{}{
			"certificate_id": certificate.ID,
		})
		return nil, diag
	}

	db := s.GetDB()
	if err := db.Save(certificate).Error; err != nil {
		diag.AddError("failed_to_create_certificate", fmt.Sprintf("failed to create certificate: %v", err), "certificates_data_store", map[string]interface{}{
			"error":          err.Error(),
			"certificate_id": certificate.ID,
		})
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
		diag.AddError("failed_to_delete_certificate", fmt.Sprintf("failed to delete certificate: %v", err), "certificates_data_store", nil)
		return diag
	}
	return diag
}
