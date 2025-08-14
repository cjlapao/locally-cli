package service

import (
	"fmt"
	"sync"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	tenant_interfaces "github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
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
	tenantService     tenant_interfaces.TenantServiceInterface
}

func Initialize(certificatesStore stores.CertificatesDataStoreInterface, tenantService tenant_interfaces.TenantServiceInterface) interfaces.CertificateServiceInterface {
	certificateServiceMutex.Lock()
	defer certificateServiceMutex.Unlock()

	certificateServiceOnce.Do(func() {
		globalCertificateService = newService(certificatesStore, tenantService)
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

func newService(certificatesStore stores.CertificatesDataStoreInterface, tenantService tenant_interfaces.TenantServiceInterface) *CertificateService {
	return &CertificateService{
		certificatesStore: certificatesStore,
		tenantService:     tenantService,
	}
}

func (s *CertificateService) GetName() string {
	return "certificate"
}

func (s *CertificateService) GenerateRootCertificate(ctx *appctx.AppContext) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")

	rootCA := NewX509RootCertificate(ctx, "root", models.CertificateConfig{
		CommonName:         "Locally Root CA",
		Country:            "UK",
		State:              "London",
		City:               "London",
		Organization:       "Locally",
		OrganizationalUnit: "Locally",
		ExpiresInYears:     10,
		KeySize:            types.CertificateKeySize2048,
		SignatureAlgorithm: types.SignatureAlgorithmSHA512,
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

func (s *CertificateService) GenerateIntermediateCertificate(ctx *appctx.AppContext, tenantId string, rootCA interfaces.X509Certificate) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")
	cfg := config.GetInstance().Get()
	tenant, tenantDiag := s.tenantService.GetTenantByIDOrSlug(ctx, tenantId)
	if tenantDiag.HasErrors() {
		diag.Append(tenantDiag)
		return nil, tenantDiag
	}
	if tenant == nil {
		diag.AddError("tenant_not_found", "tenant not found", CertificateModuleKey, nil)
		return nil, diag
	}
	if tenant.Status != pkg_types.RecordStatusActive {
		diag.AddError("tenant_not_active", "tenant is not active", CertificateModuleKey, nil)
		return nil, diag
	}

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

	keySize := types.IntToCertificateKeySize(tenantKeySize)
	signatureAlgorithm := types.StringToSignatureAlgorithm(tenantSignatureAlgorithm)

	intermediateCA := NewX509IntermediateCertificate(ctx, "intermediate", rootCA, models.CertificateConfig{
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

func (s *CertificateService) GenerateCertificate(ctx *appctx.AppContext, tenantId string, config models.CertificateConfig) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")
	ctx.Log().Infof("Generating certificate for tenant %s with config %v", tenantId, config)
	tenant, tenantDiag := s.tenantService.GetTenantByIDOrSlug(ctx, tenantId)
	if tenantDiag.HasErrors() {
		diag.Append(tenantDiag)
		return nil, tenantDiag
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
	// Getting the intermediate certificate
	intermediateCertificate, intermediateCertificateDiag := s.GetCertificateBy(ctx, tenantId, config.IntermediateCertificateID)
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}
	if intermediateCertificate == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}

	certGenSvc := NewX509ServerCertificate(ctx, config.CommonName, rootCertificate, intermediateCertificate, config)
	cert, certDiag := certGenSvc.Generate(ctx)
	if certDiag.HasErrors() {
		diag.Append(certDiag)
		return nil, certDiag
	}

	return cert, diag
}

func (s *CertificateService) GetCertificateBy(ctx *appctx.AppContext, tenantId string, slugOrId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_certificate_by")
	certificate, certificateDiag := s.certificatesStore.GetCertificateBy(ctx, tenantId, slugOrId)
	if certificateDiag.HasErrors() {
		diag.Append(certificateDiag)
		return nil, certificateDiag
	}
	if certificate == nil {
		return nil, diag
	}
	configDto := mappers.MapCertificateConfigToDto(certificate.Config)
	switch certificate.Type {
	case pkg_types.CertificateTypeRoot:
		cert := NewX509RootCertificate(ctx, certificate.Name, configDto)
		cert.FromDatabase(ctx, certificate)
		return cert, diag
	case pkg_types.CertificateTypeIntermediate:
		rootCertificate, rootCertificateDiag := s.GetRootCertificate(ctx)
		if rootCertificateDiag.HasErrors() {
			diag.Append(rootCertificateDiag)
			return nil, rootCertificateDiag
		}
		cert := NewX509IntermediateCertificate(ctx, certificate.Name, rootCertificate, configDto)
		cert.FromDatabase(ctx, certificate)
		return cert, diag
	case pkg_types.CertificateTypeCertificate:
		rootCertificate, rootCertificateDiag := s.GetRootCertificate(ctx)
		if rootCertificateDiag.HasErrors() {
			diag.Append(rootCertificateDiag)
			return nil, rootCertificateDiag
		}
		intermediateCertificate, intermediateCertificateDiag := s.GetCertificateBy(ctx, tenantId, certificate.Config.IntermediateCertificateID)
		if intermediateCertificateDiag.HasErrors() {
			diag.Append(intermediateCertificateDiag)
			return nil, intermediateCertificateDiag
		}
		cert := NewX509ServerCertificate(ctx, certificate.Name, rootCertificate, intermediateCertificate, configDto)
		cert.FromDatabase(ctx, certificate)
		return cert, diag
	default:
		diag.AddError("invalid_certificate_type", "invalid certificate type", CertificateModuleKey, nil)
		return nil, diag
	}
}

func (s *CertificateService) GetRootCertificate(ctx *appctx.AppContext) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_root_certificate")
	rootCertificate, rootCertificateDiag := s.certificatesStore.GetRootCertificate(ctx)
	if rootCertificateDiag.HasErrors() {
		diag.Append(rootCertificateDiag)
		return nil, rootCertificateDiag
	}
	configDto := mappers.MapCertificateConfigToDto(rootCertificate.Config)
	cert := NewX509RootCertificate(ctx, rootCertificate.Name, configDto)
	cert.FromDatabase(ctx, rootCertificate)
	return cert, diag
}

func (s *CertificateService) GetIntermediateCertificate(ctx *appctx.AppContext, tenantId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	return nil, nil
}

func (s *CertificateService) GetCertificate(ctx *appctx.AppContext, tenantId string, slugOrId string) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	return nil, nil
}

func (s *CertificateService) CreateCertificate(ctx *appctx.AppContext, tenantId string, certificate interfaces.X509Certificate) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	return nil, nil
}

func (s *CertificateService) DeleteCertificate(ctx *appctx.AppContext, tenantId string, slugOrId string) *diagnostics.Diagnostics {
	return nil
}
