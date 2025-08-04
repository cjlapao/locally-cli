package service

import (
	"errors"
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/mocks"
	"github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTenantDataStore implements TenantDataStoreInterface for testing
type MockTenantDataStore struct {
	*mocks.BaseMockStore
}

// Helper function to setup service with mock
func setupTenantServiceWithMock(mockStore *MockTenantDataStore) interfaces.TenantServiceInterface {
	Reset() // Reset singleton for test isolation
	Initialize(mockStore, nil, nil, nil, nil)
	return GetInstance()
}

func TestGetTenantsByFilter(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	filter := &filters.Filter{
		Page:     1,
		PageSize: 10,
	}

	// Mock data
	mockTenants := []entities.Tenant{
		{
			BaseModel: entities.BaseModel{
				ID:        "tenant1",
				CreatedAt: time.Now(),
			},
			Name:         "Test Tenant 1",
			Description:  "Test tenant description",
			Domain:       "test1.example.com",
			OwnerID:      "user1",
			ContactEmail: "admin@test1.example.com",
		},
		{
			BaseModel: entities.BaseModel{
				ID:        "tenant2",
				CreatedAt: time.Now(),
			},
			Name:         "Test Tenant 2",
			Description:  "Another test tenant",
			Domain:       "test2.example.com",
			OwnerID:      "user2",
			ContactEmail: "admin@test2.example.com",
		},
	}

	mockResponse := &filters.FilterResponse[entities.Tenant]{
		Items:      mockTenants,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockStore.On("GetTenantsByFilter", mock.AnythingOfType("*appctx.AppContext"), filter).Return(mockResponse, nil)

	result, diag := service.GetTenantsByFilter(ctx, filter)

	assert.False(t, diag.HasErrors())
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "Test Tenant 1", result.Data[0].Name)
	assert.Equal(t, "Test Tenant 2", result.Data[1].Name)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.PageSize)

	mockStore.AssertExpectations(t)
}

func TestGetTenantsByFilter_Error(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	filter := &filters.Filter{
		Page:     1,
		PageSize: 10,
	}

	mockStore.On("GetTenantsByFilter", mock.AnythingOfType("*appctx.AppContext"), filter).Return(nil, errors.New("database error"))

	result, diag := service.GetTenantsByFilter(ctx, filter)

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

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, nil)

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

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, gorm.ErrRecordNotFound)

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

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, errors.New("database error"))

	result, diag := service.GetTenantByIDOrSlug(ctx, tenantID)

	assert.True(t, diag.HasErrors())
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
}

func TestCreateTenant(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenant := &tenant_models.TenantCreateRequest{
		Name:         "New Tenant",
		Description:  "New tenant description",
		Domain:       "new.example.com",
		ContactEmail: "admin@new.example.com",
	}

	// Mock the created tenant (with ID and timestamps)
	mockCreatedTenant := &entities.Tenant{
		BaseModel: entities.BaseModel{
			ID:        "new-tenant-id",
			CreatedAt: time.Now(),
		},
		Name:         tenant.Name,
		Description:  tenant.Description,
		Domain:       tenant.Domain,
		ContactEmail: tenant.ContactEmail,
	}

	mockStore.On("CreateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(mockCreatedTenant, nil)

	result, diag := service.CreateTenant(ctx, tenant)

	assert.False(t, diag.HasErrors())
	assert.NotNil(t, result)
	assert.Equal(t, "New Tenant", result.Name)
	assert.Equal(t, "new.example.com", result.Domain)

	mockStore.AssertExpectations(t)
}

func TestCreateTenant_Error(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenant := &tenant_models.TenantCreateRequest{
		Name: "New Tenant",
	}

	mockStore.On("CreateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(nil, errors.New("database error"))

	result, diag := service.CreateTenant(ctx, tenant)

	assert.True(t, diag.HasErrors())
	assert.Nil(t, result)

	mockStore.AssertExpectations(t)
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

	mockStore.On("UpdateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(nil)

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

	mockStore.On("UpdateTenant", mock.AnythingOfType("*appctx.AppContext"), mock.AnythingOfType("*entities.Tenant")).Return(errors.New("database error"))

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

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, nil)
	mockStore.On("DeleteTenant", mock.AnythingOfType("*appctx.AppContext"), mockTenant).Return(nil)

	diag := service.DeleteTenant(ctx, tenantID)

	assert.False(t, diag.HasErrors())

	mockStore.AssertExpectations(t)
}

func TestDeleteTenant_NotFound(t *testing.T) {
	mockStore := &MockTenantDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupTenantServiceWithMock(mockStore)

	ctx := appctx.NewContext(nil)
	tenantID := "nonexistent"

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(nil, gorm.ErrRecordNotFound)

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

	mockStore.On("GetTenantByIdOrSlug", mock.AnythingOfType("*appctx.AppContext"), tenantID).Return(mockTenant, nil)
	mockStore.On("DeleteTenant", mock.AnythingOfType("*appctx.AppContext"), mockTenant).Return(errors.New("delete error"))

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
