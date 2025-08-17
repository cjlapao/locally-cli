// Package service contains the service for the tenant service
package service

import (
	"errors"
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
	"gorm.io/gorm"
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

func (s *TenantService) GetTenantsByFilter(ctx *appctx.AppContext, filter *filters.Filter) (*api_models.PaginatedResponse[models.Tenant], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenants")
	defer diag.Complete()

	dbTenants, err := s.tenantStore.GetTenantsByFilter(ctx, filter)
	if err != nil {
		diag.AddError("failed_to_get_tenants", "failed to get tenants", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	tenants := mappers.MapTenantsToDto(dbTenants.Items)
	pagination := api_models.Pagination{
		Page:       dbTenants.Page,
		PageSize:   dbTenants.PageSize,
		TotalPages: dbTenants.TotalPages,
	}

	response := api_models.PaginatedResponse[models.Tenant]{
		Data:       tenants,
		TotalCount: dbTenants.Total,
		Pagination: pagination,
	}

	return &response, diag
}

func (s *TenantService) GetTenantByIDOrSlug(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenant_by_id")
	defer diag.Complete()

	dbTenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Log().WithField("id_or_slug", idOrSlug).Infof("Tenant with id or slug %v not found", idOrSlug)
			return nil, diag
		}
		diag.AddError("failed_to_get_tenant_by_id_or_slug", "failed to get tenant by id or slug", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
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

	createdTenant, err := s.tenantStore.CreateTenant(ctx, dbTenant)
	if err != nil {
		diag.AddError("failed_to_create_tenant", "failed to create tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
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
		adminRole, err := s.systemService.GetRoleBySecurityLevel(models.SecurityLevelAdmin)
		if err != nil {
			diag.AddError("failed_to_get_admin_role", "failed to get admin role", "tenant", map[string]interface{}{
				"error": err.Error(),
			})
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
		err = s.tenantStore.UpdateTenant(ctx, createdTenant)
		if err != nil {
			diag.AddError("failed_to_update_tenant_owner_id", "failed to update tenant owner id", "tenant", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Creating the tenant intermediate certificate
	intermediateCert, generateDiag := s.certificateService.GenerateX509IntermediateCertificate(ctx, createdTenant.ID)
	if generateDiag.HasErrors() {
		diag.Append(generateDiag)
		return nil, diag
	}
	certConfig := *intermediateCert.GetConfiguration()

	_, createDiag := s.certificateService.CreateCertificate(ctx, createdTenant.ID, pkg_types.CertificateTypeIntermediate, certConfig)
	if createDiag.HasErrors() {
		diag.Append(createDiag)
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

	err := s.tenantStore.UpdateTenant(ctx, dbTenant)
	if err != nil {
		diag.AddError("failed_to_update_tenant", "failed to update tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return &tenant, diag
}

func (s *TenantService) DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_tenant")
	defer diag.Complete()

	dbTenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if err != nil {
		diag.AddError("failed_to_delete_tenant", "failed to delete tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	err = s.tenantStore.DeleteTenant(ctx, dbTenant.ID)
	if err != nil {
		diag.AddError("failed_to_delete_tenant", "failed to delete tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}
