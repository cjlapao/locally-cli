package service

import (
	"fmt"
	"sync"

	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	activity_types "github.com/cjlapao/locally-cli/internal/activity/types"
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/errors"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/certificates/models"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
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
	activityService   activity_interfaces.ActivityServiceInterface
	userStore         stores.UserDataStoreInterface
}

func Initialize(certificatesStore stores.CertificatesDataStoreInterface,
	tenantStore stores.TenantDataStoreInterface,
	activityService activity_interfaces.ActivityServiceInterface,
	userStore stores.UserDataStoreInterface,
) interfaces.CertificateServiceInterface {
	certificateServiceMutex.Lock()
	defer certificateServiceMutex.Unlock()

	certificateServiceOnce.Do(func() {
		globalCertificateService = newService(certificatesStore, tenantStore, activityService, userStore)
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
	activityService activity_interfaces.ActivityServiceInterface,
	userStore stores.UserDataStoreInterface,
) *CertificateService {
	return &CertificateService{
		certificatesStore: certificatesStore,
		tenantStore:       tenantStore,
		activityService:   activityService,
		userStore:         userStore,
	}
}

func (s *CertificateService) GetName() string {
	return "certificate"
}

func (s *CertificateService) GenerateX509RootCertificate(ctx *appctx.AppContext) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")

	rootCA := NewX509RootCertificate("root", pkg_models.CertificateConfig{
		CertificateType:    pkg_types.CertificateTypeRoot,
		CertificateID:      config.GlobalRootCertificateID,
		CommonName:         config.DefaultRootCertificateCommonName,
		Country:            config.DefaultCertificateCountry,
		State:              config.DefaultCertificateState,
		City:               config.DefaultCertificateCity,
		Organization:       config.DefaultCertificateOrganization,
		OrganizationalUnit: config.DefaultCertificateOrganizationalUnit,
		ExpiresInYears:     config.DefaultCertificateExpiresInYears,
		FQDNs:              []string{},
		IpAddresses:        []string{},
		AdminEmailAddress:  config.DefaultCertificateAdminEmailAddress,
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
	tenant, getTenantDiag := s.tenantStore.GetTenantByIdOrSlug(ctx, tenantId)
	if getTenantDiag.HasErrors() {
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

	intermediateCA := NewX509IntermediateCertificate("intermediate", rootCA, pkg_models.CertificateConfig{
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

func (s *CertificateService) GenerateX509Certificate(ctx *appctx.AppContext, tenantId string, config pkg_models.CertificateConfig) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")
	ctx.Log().Infof("Generating certificate for tenant %s with config %v", tenantId, config)
	tenant, getTenantDiag := s.tenantStore.GetTenantByIdOrSlug(ctx, tenantId)
	if getTenantDiag.HasErrors() {
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
	if rootCA == nil {
		diag.AddError("root_certificate_not_found", "root certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}

	// Getting the intermediate certificate
	intermediateCertificate, intermediateCertificateDiag := s.GetIntermediateCertificate(ctx, tenantId, "intermediate")
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}
	if intermediateCertificate == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}
	intermediateCA := CertificateFromDto(ctx, intermediateCertificate)
	if intermediateCA == nil {
		diag.AddError("intermediate_certificate_not_found", "intermediate certificate not found", CertificateModuleKey, nil)
		return nil, diag
	}

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

func (s *CertificateService) GetCertificates(ctx *appctx.AppContext, tenantId string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_certificates")
	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}
	certificates, certificatesDiag := s.certificatesStore.GetCertificatesByQuery(ctx, tenantId, query)
	if certificatesDiag.HasErrors() {
		diag.Append(certificatesDiag)
		return nil, certificatesDiag
	}

	certs := make([]pkg_models.Certificate, len(certificates.Items))
	for i, cert := range certificates.Items {
		certs[i] = mappers.MapCertificateToDto(cert)
	}

	return &api_models.PaginationResponse[pkg_models.Certificate]{
		TotalCount: certificates.Total,
		Pagination: api_models.Pagination{
			Page:       certificates.Page,
			PageSize:   certificates.PageSize,
			TotalPages: certificates.TotalPages,
		},
		Data: certs,
	}, diag
}

func (s *CertificateService) GetCertificateBy(ctx *appctx.AppContext, tenantId string, slugOrId string) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
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

func (s *CertificateService) GetCertificatesByType(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Certificate], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_certificates_by_type")
	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	certificates, certificatesDiag := s.certificatesStore.GetCertificatesByType(ctx, tenantId, certType, query)
	if certificatesDiag.HasErrors() {
		diag.Append(certificatesDiag)
		return nil, certificatesDiag
	}

	certs := make([]pkg_models.Certificate, len(certificates.Items))
	for i, cert := range certificates.Items {
		certs[i] = mappers.MapCertificateToDto(cert)
	}

	return &api_models.PaginationResponse[pkg_models.Certificate]{
		TotalCount: certificates.Total,
		Pagination: api_models.Pagination{
			Page:       certificates.Page,
			PageSize:   certificates.PageSize,
			TotalPages: certificates.TotalPages,
		},
		Data: certs,
	}, diag
}

func (s *CertificateService) CreateCertificateFromConfig(ctx *appctx.AppContext, tenantId string, certType pkg_types.CertificateType, certificateConfig pkg_models.CertificateConfig) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
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

	// creating the activity
	activityDiags := s.activityService.RecordSuccessActivity(ctx, "create_certificate", &activity_types.ActivityRecord{
		Module:        CertificateModuleKey,
		Message:       "Certificate created successfully",
		Service:       "certificates",
		Success:       true,
		ActorType:     activity_types.ActorTypeUser,
		ActivityType:  activity_types.ActivityTypeCreate,
		ActivityLevel: activity_types.ActivityLevelInfo,
		Data: &activity_types.ActivityData{
			Metadata: map[string]interface{}{
				"certificate_id": createdCert.ID,
			},
		},
	})
	if activityDiags.HasErrors() {
		diag.Append(activityDiags)
	}

	response := mappers.MapCertificateToDto(*createdCert)

	return &response, diag
}

func (s *CertificateService) GetRootCertificate(ctx *appctx.AppContext) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_root_certificate")
	rootCertificate, rootCertificateDiag := s.certificatesStore.GetRootCertificate(ctx)
	if rootCertificateDiag.HasErrors() {
		diag.Append(rootCertificateDiag)
		return nil, rootCertificateDiag
	}
	cert := mappers.MapCertificateToDto(*rootCertificate)
	return &cert, diag
}

func (s *CertificateService) GetTenantIntermediateCertificate(ctx *appctx.AppContext, tenantId string) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenant_intermediate_certificate")
	intermediateCertificate, intermediateCertificateDiag := s.certificatesStore.GetTenantIntermediateCertificate(ctx, tenantId)
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}
	cert := mappers.MapCertificateToDto(*intermediateCertificate)
	return &cert, diag
}

func (s *CertificateService) GetIntermediateCertificate(ctx *appctx.AppContext, tenantId string, slug string) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_intermediate_certificate")
	if slug == "" {
		slug = "intermediate"
	}

	intermediateCertificate, intermediateCertificateDiag := s.certificatesStore.GetCertificateByType(ctx, tenantId, pkg_types.CertificateTypeIntermediate, slug)
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}

	if intermediateCertificate == nil {
		diag.AddError(errors.ErrorMissingIntermediateCertificate, "Intermediate certificate not found", "certificates_handler")
		return nil, diag
	}

	cert := mappers.MapCertificateToDto(*intermediateCertificate)
	return &cert, diag
}

func (s *CertificateService) CreateCertificate(ctx *appctx.AppContext, tenantId string, request *models.CreateCertificateRequest) (*pkg_models.Certificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_certificate")

	userId := ctx.GetUserID()
	if userId == "" {
		diag.AddError(errors.ErrorMissingUserID, "User ID is missing", "certificates_handler")
		return nil, diag
	}

	user, userDiag := s.userStore.GetUserByID(ctx, tenantId, userId)
	if userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, userDiag
	}
	if user == nil {
		diag.AddError(errors.ErrorMissingUser, "User is missing", "certificates_handler")
		return nil, diag
	}

	userDto := mappers.MapUserToDto(user)

	if len(request.FQDNs) == 0 {
		if request.SubDomain == "" {
			diag.AddError(errors.ErrorMissingSubDomain, "Sub domain is missing", "certificates_handler", map[string]interface{}{
				"sub_domain": request.SubDomain,
			})
			return nil, diag
		}
		request.FQDNs = []string{request.SubDomain + "." + config.DefaultLocallyDomain}
	}

	// getting the root certificate
	rootCertificate, rootCertificateDiag := s.GetRootCertificate(ctx)
	if rootCertificateDiag.HasErrors() {
		diag.Append(rootCertificateDiag)
		return nil, rootCertificateDiag
	}
	if rootCertificate == nil {
		diag.AddError(errors.ErrorMissingRootCertificate, "Root certificate is missing", "certificates_handler", map[string]interface{}{
			"root_certificate": rootCertificate,
		})
		return nil, diag
	}

	// getting the intermediate certificate
	intermediateCertificate, intermediateCertificateDiag := s.GetIntermediateCertificate(ctx, tenantId, "intermediate")
	if intermediateCertificateDiag.HasErrors() {
		diag.Append(intermediateCertificateDiag)
		return nil, intermediateCertificateDiag
	}
	if intermediateCertificate == nil {
		diag.AddError(errors.ErrorMissingIntermediateCertificate, "Intermediate certificate is missing", "certificates_handler", map[string]interface{}{
			"intermediate_certificate": intermediateCertificate,
		})
		return nil, diag
	}

	certConfig := pkg_models.CertificateConfig{
		AdminEmailAddress:         request.AdminEmailAddress,
		ExpiresInYears:            request.ExpiresInYears,
		KeySize:                   request.KeySize,
		SignatureAlgorithm:        request.SignatureAlgorithm,
		Password:                  request.Password,
		RootCertificateID:         rootCertificate.ID,
		IntermediateCertificateID: intermediateCertificate.ID,
	}
	if userDto.IsSuperUser() {
		certConfig.CertificateType = request.CertificateType
		certConfig.Country = request.Country
		if request.Country == "" {
			certConfig.Country = config.DefaultCertificateCountry
		}
		certConfig.State = request.State
		if request.State == "" {
			certConfig.State = config.DefaultCertificateState
		}
		certConfig.Organization = request.Organization
		if request.Organization == "" {
			certConfig.Organization = config.DefaultCertificateOrganization
		}
		certConfig.CommonName = request.CommonName
		if request.City == "" {
			certConfig.City = config.DefaultCertificateCity
		}
		if len(request.FQDNs) == 0 {
			certConfig.FQDNs = []string{request.SubDomain + "." + config.DefaultLocallyDomain}
		} else {
			certConfig.FQDNs = request.FQDNs
		}
		if len(request.IpAddresses) == 0 {
			certConfig.IpAddresses = []string{}
		} else {
			certConfig.IpAddresses = request.IpAddresses
		}
		if request.OrganizationalUnit == "" {
			certConfig.OrganizationalUnit = config.DefaultCertificateOrganizationalUnit
		} else {
			certConfig.OrganizationalUnit = request.OrganizationalUnit
		}
	} else {
		if request.SubDomain == "" {
			diag.AddError(errors.ErrorMissingSubDomain, "Sub domain is missing", "certificates_handler", map[string]interface{}{
				"sub_domain": request.SubDomain,
			})
			return nil, diag
		}
		certConfig.CertificateType = pkg_types.CertificateTypeCertificate
		certConfig.Country = config.DefaultCertificateCountry
		certConfig.State = config.DefaultCertificateState
		certConfig.Organization = config.DefaultCertificateOrganization
		certConfig.CommonName = request.SubDomain
		certConfig.City = config.DefaultCertificateCity
		certConfig.FQDNs = []string{request.SubDomain + "." + config.DefaultLocallyDomain}
	}

	certificate, createDiag := s.CreateCertificateFromConfig(ctx, tenantId, certConfig.CertificateType, certConfig)
	if createDiag.HasErrors() {
		diag.Append(createDiag)
		return nil, createDiag
	}

	return certificate, diag
}

func (s *CertificateService) DeleteCertificate(ctx *appctx.AppContext, tenantId string, certificateId string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_certificate")

	// checking if the certificate exists
	certificate, getCertificateDiag := s.certificatesStore.GetCertificateBy(ctx, tenantId, certificateId)
	if getCertificateDiag.HasErrors() {
		diag.Append(getCertificateDiag)
		return diag
	}
	if certificate == nil {
		diag.AddError(errors.ErrorGettingCertificate, "Certificate not found", "certificates_handler", map[string]interface{}{
			"tenant_id":      tenantId,
			"certificate_id": certificateId,
		})
		return diag
	}

	if certificate.Type == pkg_types.CertificateTypeRoot {
		diag.AddError(errors.ErrorDeletingRootCertificate, "Cannot delete root certificate", "certificates_handler", map[string]interface{}{
			"tenant_id":      tenantId,
			"certificate_id": certificateId,
		})
		return diag
	}

	if certificate.Type == pkg_types.CertificateTypeIntermediate {
		diag.AddError(errors.ErrorDeletingIntermediateCertificate, "Cannot delete intermediate certificate", "certificates_handler", map[string]interface{}{
			"tenant_id":      tenantId,
			"certificate_id": certificateId,
		})
		return diag
	}

	// deleting the certificate
	deleteDiag := s.certificatesStore.DeleteCertificate(ctx, tenantId, certificateId)
	if deleteDiag.HasErrors() {
		diag.Append(deleteDiag)
		return diag
	}

	// creating the activity
	activityDiags := s.activityService.RecordSuccessActivity(ctx, "delete_certificate", &activity_types.ActivityRecord{
		Module:        CertificateModuleKey,
		Message:       "Certificate deleted successfully",
		Service:       "certificates",
		Success:       true,
		ActorType:     activity_types.ActorTypeUser,
		ActivityType:  activity_types.ActivityTypeDelete,
		ActivityLevel: activity_types.ActivityLevelInfo,
		Data: &activity_types.ActivityData{
			Metadata: map[string]interface{}{
				"certificate_id": certificateId,
			},
		},
	})
	if activityDiags.HasErrors() {
		diag.Append(activityDiags)
	}

	return diag
}
