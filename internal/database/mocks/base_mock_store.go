package mocks

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/stretchr/testify/mock"
)

// BaseMockStore provides a reusable mock for any store interface
// It embeds mock.Mock and provides common method implementations
type BaseMockStore struct {
	mock.Mock
}

// NewBaseMockStore creates a new BaseMockStore instance
func NewBaseMockStore() *BaseMockStore {
	return &BaseMockStore{}
}

// ============================================================================
// Tenant Store Methods
// ============================================================================

func (m *BaseMockStore) GetTenantBySlug(ctx *appctx.AppContext, slug string) (*entities.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantByID(ctx *appctx.AppContext, id string) (*entities.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, error) {
	args := m.Called(ctx, idOrSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
	args := m.Called(ctx, filterObj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filters.FilterResponse[entities.Tenant]), args.Error(1)
}

func (m *BaseMockStore) CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, error) {
	args := m.Called(ctx, tenant)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteTenant(ctx *appctx.AppContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) Migrate() *diagnostics.Diagnostics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*diagnostics.Diagnostics)
}

// ============================================================================
// User Store Methods
// ============================================================================

func (m *BaseMockStore) GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
	args := m.Called(ctx, tenantID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filters.FilterResponse[entities.User]), args.Error(1)
}

func (m *BaseMockStore) CreateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) (*entities.User, error) {
	args := m.Called(ctx, tenantID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) UpdateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) error {
	args := m.Called(ctx, tenantID, user)
	return args.Error(0)
}

func (m *BaseMockStore) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, password string) error {
	args := m.Called(ctx, tenantID, id, password)
	return args.Error(0)
}

func (m *BaseMockStore) BlockUser(ctx *appctx.AppContext, tenantID string, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *BaseMockStore) SetRefreshToken(ctx *appctx.AppContext, tenantID string, id string, refreshToken string) error {
	args := m.Called(ctx, tenantID, id, refreshToken)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteUser(ctx *appctx.AppContext, tenantID string, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *BaseMockStore) GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Role], error) {
	args := m.Called(ctx, tenantID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filters.FilterResponse[entities.Role]), args.Error(1)
}

func (m *BaseMockStore) GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Claim], error) {
	args := m.Called(ctx, tenantID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filters.FilterResponse[entities.Claim]), args.Error(1)
}

// ============================================================================
// Auth Store Methods
// ============================================================================

func (m *BaseMockStore) CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.APIKey) (*entities.APIKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.APIKey, error) {
	args := m.Called(ctx, keyPrefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.APIKey), args.Error(1)
}

func (m *BaseMockStore) ListAPIKeysByUserID(ctx *appctx.AppContext, userID string) ([]entities.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.APIKey), args.Error(1)
}

func (m *BaseMockStore) ListAPIKeysByUserIDWithFilter(ctx *appctx.AppContext, userID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.APIKey], error) {
	args := m.Called(ctx, userID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filters.FilterResponse[entities.APIKey]), args.Error(1)
}

func (m *BaseMockStore) UpdateAPIKeyLastUsed(ctx *appctx.AppContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error {
	args := m.Called(ctx, id, revokedBy, reason)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteAPIKey(ctx *appctx.AppContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) CreateAPIKeyUsage(ctx *appctx.AppContext, usage *entities.APIKeyUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *BaseMockStore) GetAPIKeyUsageStats(ctx *appctx.AppContext, apiKeyID string, days int) ([]entities.APIKeyUsage, error) {
	args := m.Called(ctx, apiKeyID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.APIKeyUsage), args.Error(1)
}

func (m *BaseMockStore) CleanupExpiredAPIKeys(ctx *appctx.AppContext) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ============================================================================
// Configuration Store Methods
// ============================================================================

func (m *BaseMockStore) GetConfigurationValue(ctx interface{}, key string, value interface{}) (interface{}, error) {
	args := m.Called(ctx, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// ============================================================================
// Certificates Store Methods
// ============================================================================

func (m *BaseMockStore) GetRootCertificates(ctx *appctx.AppContext) ([]entities.RootCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).([]entities.RootCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetRootCertificate(ctx *appctx.AppContext, id string) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.RootCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetRootCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.RootCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetIntermediateCertificates(ctx *appctx.AppContext) ([]entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).([]entities.IntermediateCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetIntermediateCertificate(ctx *appctx.AppContext, id string) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.IntermediateCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetIntermediateCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.IntermediateCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetCertificates(ctx *appctx.AppContext, rootCertificateID string) ([]entities.Certificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, rootCertificateID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).([]entities.Certificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetCertificate(ctx *appctx.AppContext, id string) (*entities.Certificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.Certificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) GetCertificateBySlug(ctx *appctx.AppContext, slug string) (*entities.Certificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.Certificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) CreateRootCertificate(ctx *appctx.AppContext, rootCertificate *entities.RootCertificate) (*entities.RootCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, rootCertificate)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.RootCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) CreateIntermediateCertificate(ctx *appctx.AppContext, intermediateCertificate *entities.IntermediateCertificate) (*entities.IntermediateCertificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, intermediateCertificate)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.IntermediateCertificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) CreateCertificate(ctx *appctx.AppContext, certificate *entities.Certificate) (*entities.Certificate, *diagnostics.Diagnostics) {
	args := m.Called(ctx, certificate)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.Certificate), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) DeleteRootCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) DeleteIntermediateCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*diagnostics.Diagnostics)
}

func (m *BaseMockStore) DeleteCertificate(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*diagnostics.Diagnostics)
}
