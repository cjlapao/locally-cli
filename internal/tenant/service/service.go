// Package service contains the service for the tenant service
package service

import (
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	certificates_interfaces "github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	claimsvc "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	claim_models "github.com/cjlapao/locally-cli/internal/claim/models"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	rolesvc "github.com/cjlapao/locally-cli/internal/role/interfaces"
	role_models "github.com/cjlapao/locally-cli/internal/role/models"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	usersvc "github.com/cjlapao/locally-cli/internal/user/interfaces"
	user_models "github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

var (
	globalTenantService *TenantService
	tenantServiceOnce   sync.Once
	tenantServiceMutex  sync.Mutex
)

type TenantService struct {
	tenantStore        stores.TenantDataStoreInterface
	userService        usersvc.UserServiceInterface
	roleService        rolesvc.RoleServiceInterface
	claimService       claimsvc.ClaimServiceInterface
	systemService      system_interfaces.SystemServiceInterface
	certificateService certificates_interfaces.CertificateServiceInterface
}

func Initialize(tenantStore stores.TenantDataStoreInterface,
	userService usersvc.UserServiceInterface,
	roleService rolesvc.RoleServiceInterface,
	systemService system_interfaces.SystemServiceInterface,
	claimService claimsvc.ClaimServiceInterface,
	certificateService certificates_interfaces.CertificateServiceInterface,
) interfaces.TenantServiceInterface {
	tenantServiceMutex.Lock()
	defer tenantServiceMutex.Unlock()

	tenantServiceOnce.Do(func() {
		globalTenantService = new(tenantStore, userService, roleService, systemService, claimService, certificateService)
	})
	return globalTenantService
}

func GetInstance() interfaces.TenantServiceInterface {
	if globalTenantService == nil {
		panic("tenant service not initialized")
	}
	return globalTenantService
}

// Reset resets the singleton for testing purposes
func Reset() {
	tenantServiceMutex.Lock()
	defer tenantServiceMutex.Unlock()
	globalTenantService = nil
	tenantServiceOnce = sync.Once{}
}

func new(tenantStore stores.TenantDataStoreInterface,
	userService usersvc.UserServiceInterface,
	roleService rolesvc.RoleServiceInterface,
	systemService system_interfaces.SystemServiceInterface,
	claimService claimsvc.ClaimServiceInterface,
	certificateService certificates_interfaces.CertificateServiceInterface,
) *TenantService {
	return &TenantService{
		tenantStore:        tenantStore,
		userService:        userService,
		roleService:        roleService,
		systemService:      systemService,
		claimService:       claimService,
		certificateService: certificateService,
	}
}

func (s *TenantService) GetName() string {
	return "tenant"
}

func (s *TenantService) GetTenants(ctx *appctx.AppContext, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[models.Tenant], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenants")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbTenants, getTenantsDiag := s.tenantStore.GetTenantsByQuery(ctx, query)
	if getTenantsDiag.HasErrors() {
		diag.Append(getTenantsDiag)
		return nil, getTenantsDiag
	}

	tenants := mappers.MapTenantsToDto(dbTenants.Items)

	response := api_models.PaginationResponse[models.Tenant]{
		Data:       tenants,
		TotalCount: dbTenants.Total,
		Pagination: api_models.Pagination{
			Page:       dbTenants.Page,
			PageSize:   dbTenants.PageSize,
			TotalPages: dbTenants.TotalPages,
		},
	}

	return &response, diag
}

func (s *TenantService) GetTenantByIDOrSlug(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenant_by_id")
	defer diag.Complete()

	dbTenant, getTenantDiag := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if getTenantDiag.HasErrors() {
		diag.Append(getTenantDiag)
		return nil, getTenantDiag
	}
	if dbTenant == nil {
		return nil, diag
	}

	tenant := mappers.MapTenantToDto(dbTenant)
	return tenant, diag
}

func (s *TenantService) CreateTenant(ctx *appctx.AppContext, request *tenant_models.TenantCreateRequest) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_tenant")
	defer diag.Complete()

	// check if the tenant already exists
	tenantSlug := utils.Slugify(request.Name)
	existingTenant, getDiag := s.GetTenantByIDOrSlug(ctx, tenantSlug)
	if getDiag.HasErrors() {
		diag.Append(getDiag)
		return nil, diag
	}
	if existingTenant != nil {
		diag.AddError("tenant_already_exists", "tenant already exists", "tenant", map[string]interface{}{
			"tenant_slug": tenantSlug,
		})
		return nil, diag
	}

	tenant := MapTenantCreateRequestToTenant(request)
	tenant.Slug = utils.Slugify(tenant.Name)
	dbTenant := mappers.MapTenantToEntity(tenant)

	createdTenant, createDiag := s.tenantStore.CreateTenant(ctx, dbTenant)
	if createDiag.HasErrors() {
		diag.Append(createDiag)
		return nil, diag
	}

	// creating the default claims for the tenant
	claims := s.systemService.GenerateSystemClaims()
	for _, claim := range claims {
		claim.TenantID = createdTenant.ID
		claimRequest := claim_models.CreateClaimRequest{
			Service:       claim.Service,
			Module:        claim.Module,
			Action:        claim.Action,
			SecurityLevel: claim.SecurityLevel,
		}
		_, createDiag := s.claimService.CreateClaim(ctx, createdTenant.ID, &claimRequest)
		if createDiag.HasErrors() {
			diag.Append(createDiag)
			// reverting the tenant creation
			s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
			return nil, diag
		}
	}

	// Creating the roles for the tenant
	roles := s.systemService.GenerateDefaultRoles()
	for _, role := range roles {
		role.TenantID = createdTenant.ID
		roleRequest := role_models.CreateRoleRequest{
			Name:          role.Name,
			Description:   role.Description,
			SecurityLevel: role.SecurityLevel,
		}
		_, createDiag := s.roleService.CreateRole(ctx, createdTenant.ID, &roleRequest)
		if createDiag.HasErrors() {
			diag.Append(createDiag)
			// reverting the tenant creation
			s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
			return nil, diag
		}
	}

	if request.CreateAdminUser {
		// creating the admin user for the tenant
		adminUser := user_models.CreateUserRequest{
			Email:    request.AdminContactEmail,
			Password: request.AdminPassword,
			Name:     request.AdminName,
		}
		adminRole, getRoleDiag := s.systemService.GetRoleBySecurityLevel(models.SecurityLevelAdmin)
		if getRoleDiag.HasErrors() {
			diag.Append(getRoleDiag)
			// reverting the tenant creation
			s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
			return nil, diag
		}

		createdAdminUser, createDiag := s.userService.CreateUser(ctx, createdTenant.ID, adminRole.ID, &adminUser)
		if createDiag.HasErrors() {
			diag.Append(createDiag)
			// reverting the tenant creation
			s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
			return nil, diag
		}

		// Updating tenant owner id
		createdTenant.OwnerID = createdAdminUser.ID
		updateDiag := s.tenantStore.UpdateTenant(ctx, createdTenant)
		if updateDiag.HasErrors() {
			diag.Append(updateDiag)
			// reverting the tenant creation
			s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
			return nil, diag
		}
	}

	// Creating the tenant intermediate certificate
	intermediateCert, generateDiag := s.certificateService.GenerateX509IntermediateCertificate(ctx, createdTenant.ID)
	if generateDiag.HasErrors() {
		diag.Append(generateDiag)
		return nil, diag
	}
	certConfig := *intermediateCert.GetConfiguration()

	_, createCertDiag := s.certificateService.CreateCertificate(ctx, createdTenant.ID, pkg_types.CertificateTypeIntermediate, certConfig)
	if createCertDiag.HasErrors() {
		diag.Append(createCertDiag)
		// reverting the tenant creation
		s.tenantStore.DeleteTenant(ctx, createdTenant.ID)
		return nil, diag
	}

	return mappers.MapTenantToDto(createdTenant), diag
}

func (s *TenantService) UpdateTenant(ctx *appctx.AppContext, tenantRequest *tenant_models.TenantUpdateRequest) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_tenant")
	defer diag.Complete()

	tenant := models.Tenant{
		ID:           tenantRequest.ID,
		Description:  tenantRequest.Description,
		Name:         tenantRequest.Name,
		Domain:       tenantRequest.Domain,
		OwnerID:      tenantRequest.OwnerID,
		ContactEmail: tenantRequest.ContactEmail,
	}

	dbTenant := mappers.MapTenantToEntity(&tenant)

	updateDiag := s.tenantStore.UpdateTenant(ctx, dbTenant)
	if updateDiag.HasErrors() {
		diag.Append(updateDiag)
		return nil, diag
	}

	return &tenant, diag
}

func (s *TenantService) DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_tenant")
	defer diag.Complete()

	dbTenant, getTenantDiag := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if getTenantDiag.HasErrors() {
		diag.Append(getTenantDiag)
		return diag
	}

	deleteDiag := s.tenantStore.DeleteTenant(ctx, dbTenant.ID)
	if deleteDiag.HasErrors() {
		diag.Append(deleteDiag)
		return diag
	}

	return diag
}
