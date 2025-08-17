package service

import (
	"fmt"
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
)

const (
	CertificateModuleKey = "certificate_module"
)

var (
	globalCertificateService interfaces.CertificateServiceInterface
	certificateServiceOnce   sync.Once
	certificateServiceMutex  sync.Mutex
)

type CertificateService struct {
	certificatesStore stores.CertificatesDataStoreInterface
	tenantStore       stores.TenantDataStoreInterface
}

func Initialize(certificatesStore stores.CertificatesDataStoreInterface,
	tenantStore stores.TenantDataStoreInterface,
) interfaces.CertificateServiceInterface {
	certificateServiceMutex.Lock()
	defer certificateServiceMutex.Unlock()

	certificateServiceOnce.Do(func() {
		globalCertificateService = newService(certificatesStore, tenantStore)
	})
	return globalCertificateService
}

func GetInstance() interfaces.CertificateServiceInterface {
	if globalCertificateService == nil {
		panic("certificate service not initialized")
	}
	return globalCertificateService
}

// Reset resets the singleton for testing purposes
func Reset() {
	certificateServiceMutex.Lock()
	defer certificateServiceMutex.Unlock()
	globalCertificateService = nil
	certificateServiceOnce = sync.Once{}
}

func newService(certificatesStore stores.CertificatesDataStoreInterface,
	tenantStore stores.TenantDataStoreInterface,
) *CertificateService {
	return &CertificateService{
		certificatesStore: certificatesStore,
		tenantStore:       tenantStore,
	}
}

func (s *CertificateService) GetName() string {
	return "certificate"
}

func (s *CertificateService) GenerateX509RootCertificate(ctx *appctx.AppContext) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")

	rootCA := NewX509RootCertificate("root", models.CertificateConfig{
		CertificateType:    pkg_types.CertificateTypeRoot,
		CertificateID:      config.GlobalRootCertificateID,
		CommonName:         "Locally Root CA",
		Country:            "UK",
		State:              "London",
		City:               "London",
		Organization:       "Locally",
		OrganizationalUnit: "Locally",
		ExpiresInYears:     10,
		FQDNs:              []string{},
		IpAddresses:        []string{},
		AdminEmailAddress:  "admin@locally.com",
		KeySize:            pkg_types.CertificateKeySize2048,
		SignatureAlgorithm: pkg_types.SignatureAlgorithmSHA512,
	})

	certificate, certDiag := rootCA.Generate(ctx)
	if certDiag.HasErrors() {
		ctx.Log().Error("Error generating root certificate", certDiag.Errors)
		diag.Append(certDiag)
		return nil, certDiag
	}

	ctx.Log().Debug("Generated Root Certificate")
	ctx.Log().Debugf("\nPem: %s", certificate.GetPemCertificate())
	ctx.Log().Debugf("\nPrivate Key: %s", certificate.GetPemPrivateKey())
	ctx.Log().Debugf("\nCSR: %s", certificate.GetPemCsr())
	ctx.Log().Debugf("\nConfig: %v", certificate.GetConfiguration())

	return certificate, diag
}

func (s *CertificateService) GenerateX509IntermediateCertificate(ctx *appctx.AppContext, tenantId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")
	cfg := config.GetInstance().Get()
	tenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, tenantId)
	if err != nil {
		diag.AddError("tenant_not_found", "tenant not found", CertificateModuleKey, nil)
		return nil, diag
	}
	if tenant == nil {
		diag.AddError("tenant_not_found", "tenant not found", CertificateModuleKey, nil)
		return nil, diag
	}
	if tenant.Status != pkg_types.RecordStatusActive {
		diag.AddError("tenant_not_active", "tenant is not active", CertificateModuleKey, nil)
		return nil, diag
	}
	// getting the root certificate

	rootCADb, rootDiag := s.certificatesStore.GetRootCertificate(ctx)
	if rootDiag.HasErrors() {
		diag.Append(rootDiag)
		return nil, rootDiag
	}
	if rootCADb == nil {
		diag.AddError("root_certificate_not_found", "root certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}
	rootCA := CertificateFromEntity(ctx, rootCADb)

	commonName := fmt.Sprintf("%s Intermediate CA", tenant.Name)
	tenantCountry := tenant.Country
	if tenantCountry == "" {
		tenantCountry = cfg.GetString(config.CertificateCountryKey, config.DefaultCertificateCountry)
	}
	tenantState := tenant.State
	if tenantState == "" {
		tenantState = cfg.GetString(config.CertificateStateKey, config.DefaultCertificateState)
	}
	tenantCity := tenant.City
	if tenantCity == "" {
		tenantCity = cfg.GetString(config.CertificateCityKey, config.DefaultCertificateCity)
	}
	tenantOrganization := tenant.Name
	if tenantOrganization == "" {
		tenantOrganization = cfg.GetString(config.CertificateOrganizationKey, config.DefaultCertificateOrganization)
	}
	tenantOrganizationalUnit := tenant.Name
	if tenantOrganizationalUnit == "" {
		tenantOrganizationalUnit = cfg.GetString(config.CertificateOrganizationalUnitKey, config.DefaultCertificateOrganizationalUnit)
	}
	tenantExpiresInYears := cfg.GetInt(config.CertificateExpiresInYearsKey, config.DefaultCertificateExpiresInYears)
	tenantKeySize := cfg.GetInt(config.CertificateKeySizeKey, config.DefaultCertificateKeySize)
	tenantSignatureAlgorithm := cfg.GetString(config.CertificateSignatureAlgorithmKey, config.DefaultCertificateSignatureAlgorithm)

	keySize := pkg_types.IntToCertificateKeySize(tenantKeySize)
	signatureAlgorithm := pkg_types.StringToSignatureAlgorithm(tenantSignatureAlgorithm)

	intermediateCA := NewX509IntermediateCertificate("intermediate", rootCA, models.CertificateConfig{
		CertificateType:    pkg_types.CertificateTypeIntermediate,
		RootCertificateID:  rootCADb.ID,
		CommonName:         commonName,
		Country:            tenantCountry,
		State:              tenantState,
		City:               tenantCity,
		Organization:       tenantOrganization,
		OrganizationalUnit: tenantOrganizationalUnit,
		ExpiresInYears:     tenantExpiresInYears,
		KeySize:            keySize,
		SignatureAlgorithm: signatureAlgorithm,
	})

	certificate, certDiag := intermediateCA.Generate(ctx)

	if certDiag.HasErrors() {
		ctx.Log().Error("Error generating intermediate certificate", certDiag.Errors)
		diag.Append(certDiag)
		return nil, certDiag
	}

	return certificate, diag
}

func (s *CertificateService) GenerateX509Certificate(ctx *appctx.AppContext, tenantId string, config models.CertificateConfig) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")
	ctx.Log().Infof("Generating certificate for tenant %s with config %v", tenantId, config)
	tenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, tenantId)
	if err != nil {
		diag.AddError("tenant_not_found", "tenant not found", CertificateModuleKey, nil)
		return nil, diag
	}
	if tenant == nil {
		diag.AddError("tenant_not_found", "tenant not found", CertificateModuleKey, nil)
		return nil, diag
	}
	if tenant.Status != pkg_types.RecordStatusActive {
		diag.AddError("tenant_not_active", "tenant is not active", CertificateModuleKey, nil)
		return nil, diag
	}
	if validateDiag := config.Validate(); validateDiag.HasErrors() {
		diag.Append(validateDiag)
		return nil, diag
	}

	// checking if we already have a certificate with this name
	certificate, certificateDiag := s.GetCertificateBy(ctx, tenantId, config.CommonName)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate != nil {
		diag.AddError("certificate_already_exists", "certificate already exists", CertificateModuleKey, nil)
		return nil, diag
	}

	// Getting the root certificate
	rootCertificate, rootCertificateDiag := s.GetRootCertificate(ctx)
	if rootCertificateDiag.HasErrors() {
		diag.Append(rootCertificateDiag)
		return nil, rootCertificateDiag
	}
	if rootCertificate == nil {
		diag.AddError("root_certificate_not_found", "root certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}
	rootCA := CertificateFromDto(ctx, rootCertificate)

	// Getting the intermediate certificate
	intermediateCertificate, intermediateCertificateDiag := s.GetIntermediateCertificate(ctx, tenantId, "default")
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}
	if intermediateCertificate == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}
	intermediateCA := CertificateFromDto(ctx, intermediateCertificate)

	certGenSvc := NewX509ServerCertificate(config.CommonName, rootCA, intermediateCA, config)
	cert, certDiag := certGenSvc.Generate(ctx)
	if certDiag.HasErrors() {
		diag.Append(certDiag)
		return nil, certDiag
	}

	return cert, diag
}

func (s *CertificateService) GetX509Certificate(ctx *appctx.AppContext, tenantId string, certificateId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_x509_certificate")
	certificate, certificateDiag := s.certificatesStore.GetCertificateBy(ctx, tenantId, certificateId)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate == nil {
		return nil, diag
	}
	cert := CertificateFromEntity(ctx, certificate)
	return cert, diag
}

func (s *CertificateService) GetCertificates(ctx *appctx.AppContext, tenantId string, pagination *api_models.PaginationRequest) (*api_models.PaginatedResponse[models.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_certificates")
	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}
	certificates, certificatesDiag := s.certificatesStore.GetCertificates(ctx, tenantId, query)
	if certificatesDiag.HasErrors() {
		diag.Append(certificatesDiag)
		return nil, certificatesDiag
	}

	certs := make([]models.Certificate, len(certificates.Items))
	for i, cert := range certificates.Items {
		certs[i] = mappers.MapCertificateToDto(cert)
	}

	return &api_models.PaginatedResponse[models.Certificate]{
		TotalCount: certificates.Total,
		Pagination: api_models.Pagination{
			Page:       certificates.Page,
			PageSize:   certificates.PageSize,
			TotalPages: certificates.TotalPages,
		},
		Data: certs,
	}, diag
}

func (s *CertificateService) GetCertificateBy(ctx *appctx.AppContext, tenantId string, slugOrId string) (*models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_certificate_by")
	certificate, certificateDiag := s.certificatesStore.GetCertificateBy(ctx, tenantId, slugOrId)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate == nil {
		return nil, diag
	}

	cert := mappers.MapCertificateToDto(*certificate)
	return &cert, diag
}

func (s *CertificateService) CreateCertificate(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, certificateConfig models.CertificateConfig) (*models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_certificate")
	validationDiags := certificateConfig.Validate()
	if validationDiags.HasErrors() {
		diag.Append(validationDiags)
		return nil, diag
	}
	if tenantId == "" {
		diag.AddError("tenant_id_empty", "tenant ID is required", CertificateModuleKey, nil)
		return nil, diag
	}
	userID := ctx.GetUserID()
	if userID == "" {
		userID = config.DefaultSuperUserUserID
	}

	var x509Certificate interfaces.X509Certificate
	switch certType {
	case pkg_types.CertificateTypeRoot:
		rootCertificate, rootCertificateDiag := s.GenerateX509RootCertificate(ctx)
		if rootCertificateDiag.HasErrors() {
			diag.Append(rootCertificateDiag)
			return nil, rootCertificateDiag
		}
		x509Certificate = rootCertificate
	case pkg_types.CertificateTypeIntermediate:
		intermediateCertificate, intermediateCertificateDiag := s.GenerateX509IntermediateCertificate(ctx, tenantId)
		if intermediateCertificateDiag.HasErrors() {
			diag.Append(intermediateCertificateDiag)
			return nil, intermediateCertificateDiag
		}
		x509Certificate = intermediateCertificate
	case pkg_types.CertificateTypeCertificate:
		serverCertificate, serverCertificateDiag := s.GenerateX509Certificate(ctx, tenantId, certificateConfig)
		if serverCertificateDiag.HasErrors() {
			diag.Append(serverCertificateDiag)
			return nil, serverCertificateDiag
		}
		x509Certificate = serverCertificate
	}

	// Saving the certificate to the database
	dbCert := mappers.MapX509CertificateToEntity(x509Certificate)
	dbCert.TenantID = tenantId
	dbCert.CreatedBy = userID
	dbCert.Type = certType

	createdCert, createDiags := s.certificatesStore.CreateCertificate(ctx, tenantId, &dbCert)
	if createDiags.HasErrors() {
		diag.Append(createDiags)
		return nil, diag
	}

	response := mappers.MapCertificateToDto(*createdCert)

	return &response, diag
}

func (s *CertificateService) GetRootCertificate(ctx *appctx.AppContext) (*models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_root_certificate")
	rootCertificate, rootCertificateDiag := s.certificatesStore.GetRootCertificate(ctx)
	if rootCertificateDiag.HasErrors() {
		diag.Append(rootCertificateDiag)
		return nil, rootCertificateDiag
	}
	cert := mappers.MapCertificateToDto(*rootCertificate)
	return &cert, diag
}

func (s *CertificateService) GetIntermediateCertificate(ctx *appctx.AppContext, tenantId string, slug string) (*models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_intermediate_certificate")
	if slug == "" {
		slug = "default"
	}

	intermediateCertificates, intermediateCertificateDiag := s.certificatesStore.GetCertificateByType(ctx, tenantId, pkg_types.CertificateTypeIntermediate, slug)
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}

	cert := mappers.MapCertificateToDto(*intermediateCertificates)
	return &cert, diag
}

// ******************* fixing this later
func (s *CertificateService) GetCertificate(ctx *appctx.AppContext, tenantId string, slugOrId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	return nil, nil
}

func (s *CertificateService) DeleteCertificate(ctx *appctx.AppContext, tenantId string, slugOrId string) *diagnostics.Diagnostics {
	return nil
}
