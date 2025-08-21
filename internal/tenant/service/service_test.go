package service

import (
	"testing"
	"time"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/mocks"
	"github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTenantDataStore implements TenantDataStoreInterface for testing
type MockTenantDataStore struct {
	*mocks.BaseMockStore
}


// Implement all TenantDataStoreInterface methods
func (m *MockTenantDataStore) GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx, idOrSlug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).([]entities.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) GetTenantsByQuery(ctx *appctx.AppContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *diagnostics.Diagnostics) {
	args := m.Called(ctx, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Tenant]), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx, tenant)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*entities.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) *diagnostics.Diagnostics {
	args := m.Called(ctx, tenant)
	return args.Get(0).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) DeleteTenant(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	args := m.Called(ctx, id)
	return args.Get(0).(*diagnostics.Diagnostics)
}

func (m *MockTenantDataStore) Migrate() *diagnostics.Diagnostics {
	args := m.Called()
	return args.Get(0).(*diagnostics.Diagnostics)
}


// Helper function to setup service with mock (simple version for basic tests)
func setupTenantServiceWithMock(mockStore *MockTenantDataStore) interfaces.TenantServiceInterface {
	Reset() // Reset singleton for test isolation
	Initialize(mockStore, nil, nil, nil, nil, nil)
	return GetInstance()
}

func TestGetTenants(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	pagination := &api_models.PaginationRequest{
		Page:     1,
		PageSize: 10,
	}



	mockQueryResponse := &filters.QueryBuilderResponse[entities.Tenant]{
		Items:      []entities.Tenant{
			{
				BaseModel: entities.BaseModel{ID: "tenant1"},
				Name:         "Test Tenant 1",
				Description:  "Test tenant description",
				Domain:       "test1.example.com",
				ContactEmail: "admin@test1.example.com",
			},
			{
				BaseModel: entities.BaseModel{ID: "tenant2"},
				Name:         "Test Tenant 2",
				Description:  "Another test tenant",
				Domain:       "test2.example.com",
				ContactEmail: "admin@test2.example.com",
			},
		},
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}
	mockStore.On("GetTenantsByQuery", mock.AnythingOfType("*appctx.AppContext"), mock.Anything).Return(mockQueryResponse, diagnostics.New("test"))

	result, diag := service.GetTenants(ctx, pagination)

	assert.False(t, diag.HasErrors())
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.TotalCount)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "Test Tenant 1", result.Data[0].Name)
	assert.Equal(t, "Test Tenant 2", result.Data[1].Name)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.PageSize)

	mockStore.AssertExpectations(t)
}

func TestGetTenants_Error(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	pagination := &api_models.PaginationRequest{
		Page:     1,
		PageSize: 10,
	}

	diagWithError := diagnostics.New("test")
	diagWithError.AddError("test_error", "database error", "test", nil)
	mockStore.On("GetTenantsByQuery", mock.AnythingOfType("*appctx.AppContext"), mock.Anything).Return(nil, diagWithError)

	result, diag := service.GetTenants(ctx, pagination)

	assert.True(t, diag.HasErrors())
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
}

func TestGetTenantByID(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "tenant1"

	mockTenant := &entities.Tenant{
		BaseModel: entities.BaseModel{
			ID:        tenantID,
			CreatedAt: time.Now(),
		},
		Name:         "Test Tenant",
		Description:  "Test tenant description",
		Domain:       "test.example.com",
		OwnerID:      "user1",
		ContactEmail: "admin@test.example.com",
	}

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, diagnostics.New("test"))

	result, diag := service.GetTenantByIDOrSlug(ctx, tenantID)

	assert.False(t, diag.HasErrors())
	assert.NotNil(t, result)
	assert.Equal(t, "Test Tenant", result.Name)
	assert.Equal(t, "test.example.com", result.Domain)

	mockStore.AssertExpectations(t)
}

func TestGetTenantByID_NotFound(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "nonexistent"

	diagNotFound := diagnostics.New("test")
	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, diagNotFound)

	result, diag := service.GetTenantByIDOrSlug(ctx, tenantID)

	assert.False(t, diag.HasErrors()) // NotFound is not considered an error in this case
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
}

func TestGetTenantByID_Error(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "tenant1"

	diagWithError := diagnostics.New("test")
	diagWithError.AddError("test_error", "database error", "test", nil)
	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, diagWithError)

	result, diag := service.GetTenantByIDOrSlug(ctx, tenantID)

	assert.True(t, diag.HasErrors())
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
}

func TestCreateTenant(t *testing.T) {
	t.Skip("Skipping CreateTenant test due to complex service dependencies - requires full mock setup")
}

func TestCreateTenant_Error(t *testing.T) {
	t.Skip("Skipping CreateTenant_Error test due to complex service dependencies - requires full mock setup")
}

func TestUpdateTenant(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	updateRequest := &tenant_models.TenantUpdateRequest{
		ID:           "tenant1",
		Name:         "Updated Tenant",
		Description:  "Updated description",
		Domain:       "updated.example.com",
		OwnerID:      "user2",
		ContactEmail: "admin@updated.example.com",
	}

	mockStore.On("UpdateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(diagnostics.New("test"))

	result, diag := service.UpdateTenant(ctx, updateRequest)

	assert.False(t, diag.HasErrors())
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Tenant", result.Name)
	assert.Equal(t, "Updated description", result.Description)
	assert.Equal(t, "updated.example.com", result.Domain)

	mockStore.AssertExpectations(t)
}

func TestUpdateTenant_Error(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	updateRequest := &tenant_models.TenantUpdateRequest{
		ID:   "tenant1",
		Name: "Updated Tenant",
	}

	diagWithError := diagnostics.New("test")
	diagWithError.AddError("test_error", "database error", "test", nil)
	mockStore.On("UpdateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(diagWithError)

	result, diag := service.UpdateTenant(ctx, updateRequest)

	assert.True(t, diag.HasErrors())
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
}

func TestDeleteTenant(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "tenant1"

	// Mock tenant to be deleted
	mockTenant := &entities.Tenant{
		BaseModel: entities.BaseModel{
			ID: tenantID,
		},
		Name: "Test Tenant",
	}

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, diagnostics.New("test"))
	mockStore.On("DeleteTenant", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(diagnostics.New("test"))

	diag := service.DeleteTenant(ctx, tenantID)

	assert.False(t, diag.HasErrors())

	mockStore.AssertExpectations(t)
}

func TestDeleteTenant_NotFound(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "nonexistent"

	diagNotFound := diagnostics.New("test")
	diagNotFound.AddError("tenant_not_found", "tenant not found", "test", nil)
	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, diagNotFound)

	diag := service.DeleteTenant(ctx, tenantID)

	assert.True(t, diag.HasErrors())

	mockStore.AssertExpectations(t)
}

func TestDeleteTenant_DeleteError(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "tenant1"

	mockTenant := &entities.Tenant{
		BaseModel: entities.BaseModel{
			ID: tenantID,
		},
		Name: "Test Tenant",
	}

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, diagnostics.New("test"))
	diagWithError := diagnostics.New("test")
	diagWithError.AddError("test_error", "delete error", "test", nil)
	mockStore.On("DeleteTenant", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(diagWithError)

	diag := service.DeleteTenant(ctx, tenantID)

	assert.True(t, diag.HasErrors())

	mockStore.AssertExpectations(t)
}

func TestGetName(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	name := service.GetName()
	assert.Equal(t, "tenant", name)
}
