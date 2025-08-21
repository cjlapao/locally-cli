package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTenantService implements TenantServiceInterface for testing
type MockTenantService struct {
	mock.Mock
}

// Helper function to match any TenantCreateRequest regardless of import path
func anyTenantCreateRequest() interface{} {
	return mock.MatchedBy(func(req interface{}) bool {
		// Accept any struct that has the required fields for TenantCreateRequest
		v := reflect.ValueOf(req)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return false
		}
		// Check if it has the Name field which is required
		nameField := v.FieldByName("Name")
		return nameField.IsValid() && nameField.Kind() == reflect.String
	})
}

// Helper function to match any TenantUpdateRequest regardless of import path
func anyTenantUpdateRequest() interface{} {
	return mock.MatchedBy(func(req interface{}) bool {
		// Accept any struct that has the required fields for TenantUpdateRequest
		v := reflect.ValueOf(req)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return false
		}
		// Check if it has the ID field which is expected for updates
		idField := v.FieldByName("ID")
		return idField.IsValid() && idField.Kind() == reflect.String
	})
}

// Implement all TenantServiceInterface methods
func (m *MockTenantService) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTenantService) GetTenants(ctx *appctx.AppContext, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[models.Tenant], *diagnostics.Diagnostics) {
	args := m.Called(ctx, pagination)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*api_models.PaginationResponse[models.Tenant]), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantService) GetTenantByIDOrSlug(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx, idOrSlug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*models.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantService) CreateTenant(ctx *appctx.AppContext, request *tenant_models.TenantCreateRequest) (*models.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*models.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantService) UpdateTenant(ctx *appctx.AppContext, tenantRequest *tenant_models.TenantUpdateRequest) (*models.Tenant, *diagnostics.Diagnostics) {
	args := m.Called(ctx, tenantRequest)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*diagnostics.Diagnostics)
	}
	return args.Get(0).(*models.Tenant), args.Get(1).(*diagnostics.Diagnostics)
}

func (m *MockTenantService) DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics {
	args := m.Called(ctx, idOrSlug)
	return args.Get(0).(*diagnostics.Diagnostics)
}

// Helper functions
func setupTestHandler() (*APIHandler, *MockTenantService) {
	// Initialize config service for tests
	_, err := config.Initialize()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config: %v", err))
	}

	mockService := &MockTenantService{}
	handler := NewApiHandler(mockService)
	return handler, mockService
}

func createTestRequest(method, path string, body interface{}) *http.Request {
	var req *http.Request

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// Add context with appctx
	ctx := appctx.NewContext(nil)
	req = req.WithContext(ctx)

	return req
}

func createTestRequestWithVars(method, path string, vars map[string]string, body interface{}) *http.Request {
	req := createTestRequest(method, path, body)
	req = mux.SetURLVars(req, vars)
	return req
}

func TestHandleGetTenants(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockTenantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Get all tenants",
			queryParams: "",
			mockSetup: func(mockService *MockTenantService) {
				expectedResponse := &api_models.PaginationResponse[models.Tenant]{
					Data: []models.Tenant{
						{ID: "1", Name: "Tenant 1", Domain: "tenant1.com"},
						{ID: "2", Name: "Tenant 2", Domain: "tenant2.com"},
					},
					Pagination: api_models.Pagination{Page: 1, PageSize: 10, TotalPages: 1},
					TotalCount: 2,
				}
				diag := diagnostics.New("test")
				mockService.On("GetTenants", mock.Anything, mock.Anything).
					Return(expectedResponse, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total_count":2`,
		},
		{
			name:        "Success - With pagination",
			queryParams: "?page=2&page_size=5",
			mockSetup: func(mockService *MockTenantService) {
				expectedResponse := &api_models.PaginationResponse[models.Tenant]{
					Data:       []models.Tenant{},
					Pagination: api_models.Pagination{Page: 2, PageSize: 5, TotalPages: 1},
					TotalCount: 0,
				}
				diag := diagnostics.New("test")
				mockService.On("GetTenants", mock.Anything, mock.Anything).
					Return(expectedResponse, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"page":2`,
		},
		{
			name:        "Error - Service error",
			queryParams: "",
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				diag.AddError("test_error", "test error", "test", nil)
				mockService.On("GetTenants", mock.Anything, mock.Anything).
					Return(nil, diag)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			req := createTestRequest(http.MethodGet, "/v1/tenants"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.HandleGetTenants(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleGetTenant(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		mockSetup      func(*MockTenantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success - Get tenant by ID",
			tenantID: "123",
			mockSetup: func(mockService *MockTenantService) {
				expectedTenant := &models.Tenant{
					ID:     "123",
					Name:   "Test Tenant",
					Domain: "test.com",
				}
				diag := diagnostics.New("test")
				mockService.On("GetTenantByIDOrSlug", mock.AnythingOfType("*appctx.AppContext"), "123").
					Return(expectedTenant, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"id":"123"`,
		},
		{
			name:     "Success - Get tenant by slug",
			tenantID: "test-tenant",
			mockSetup: func(mockService *MockTenantService) {
				expectedTenant := &models.Tenant{
					ID:     "456",
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Domain: "test.com",
				}
				diag := diagnostics.New("test")
				mockService.On("GetTenantByIDOrSlug", mock.Anything, "test-tenant").
					Return(expectedTenant, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"slug":"test-tenant"`,
		},
		{
			name:     "Error - Tenant not found",
			tenantID: "nonexistent",
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				mockService.On("GetTenantByIDOrSlug", mock.Anything, "nonexistent").
					Return(nil, diag)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"error"`,
		},
		{
			name:     "Error - Service error",
			tenantID: "123",
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				diag.AddError("test_error", "test error", "test", nil)
				mockService.On("GetTenantByIDOrSlug", mock.AnythingOfType("*appctx.AppContext"), "123").
					Return(nil, diag)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error"`,
		},
		{
			name:           "Error - Missing tenant ID",
			tenantID:       "",
			mockSetup:      func(mockService *MockTenantService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request with URL variables
			vars := map[string]string{}
			if tt.tenantID != "" {
				vars["id"] = tt.tenantID
			}
			req := createTestRequestWithVars(http.MethodGet, "/v1/tenants/{id}", vars, nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.HandleGetTenant(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleCreateTenant(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    tenant_models.TenantCreateRequest
		mockSetup      func(*MockTenantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Create tenant",
			requestBody: tenant_models.TenantCreateRequest{
				Name:              "New Tenant",
				Domain:            "newtenant.com",
				ContactEmail:      "contact@newtenant.com",
				AdminUser:         "admin",
				AdminPassword:     "Password123!",
				AdminName:         "Admin User",
				AdminContactEmail: "admin@newtenant.com",
			},
			mockSetup: func(mockService *MockTenantService) {
				expectedTenant := &models.Tenant{
					ID:     "123",
					Name:   "New Tenant",
					Domain: "newtenant.com",
					Slug:   "new-tenant",
				}
				diag := diagnostics.New("test")
				mockService.On("CreateTenant", mock.Anything, anyTenantCreateRequest()).
					Return(expectedTenant, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"id":"123"`,
		},
		{
			name: "Error - Invalid request body",
			requestBody: tenant_models.TenantCreateRequest{
				Name: "", // Invalid: empty name
			},
			mockSetup:      func(mockService *MockTenantService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
		},
		{
			name: "Error - Service error",
			requestBody: tenant_models.TenantCreateRequest{
				Name:              "New Tenant",
				Domain:            "newtenant.com",
				ContactEmail:      "contact@newtenant.com",
				AdminUser:         "admin",
				AdminPassword:     "Password123!",
				AdminName:         "Admin User",
				AdminContactEmail: "admin@newtenant.com",
			},
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				diag.AddError("test_error", "test error", "test", nil)
				mockService.On("CreateTenant", mock.Anything, anyTenantCreateRequest()).
					Return(nil, diag)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request
			req := createTestRequest(http.MethodPost, "/v1/tenants", tt.requestBody)
			w := httptest.NewRecorder()

			// Execute request
			handler.HandleCreateTenant(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleUpdateTenant(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		requestBody    tenant_models.TenantUpdateRequest
		mockSetup      func(*MockTenantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success - Update tenant",
			tenantID: "123",
			requestBody: tenant_models.TenantUpdateRequest{
				Name:   "Updated Tenant",
				Domain: "updated.com",
			},
			mockSetup: func(mockService *MockTenantService) {
				expectedTenant := &models.Tenant{
					ID:     "123",
					Name:   "Updated Tenant",
					Domain: "updated.com",
					Slug:   "updated-tenant",
				}
				diag := diagnostics.New("test")
				mockService.On("UpdateTenant", mock.Anything, anyTenantUpdateRequest()).
					Return(expectedTenant, diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Tenant updated successfully"`,
		},
		{
			name:           "Error - Missing tenant ID",
			tenantID:       "",
			requestBody:    tenant_models.TenantUpdateRequest{},
			mockSetup:      func(mockService *MockTenantService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
		},
		{
			name:     "Error - Service error",
			tenantID: "123",
			requestBody: tenant_models.TenantUpdateRequest{
				Name:   "Updated Tenant",
				Domain: "updated.com",
			},
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				diag.AddError("test_error", "test error", "test", nil)
				mockService.On("UpdateTenant", mock.Anything, anyTenantUpdateRequest()).
					Return(nil, diag)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request with URL variables
			vars := map[string]string{}
			if tt.tenantID != "" {
				vars["id"] = tt.tenantID
			}
			req := createTestRequestWithVars(http.MethodPut, "/v1/tenants/{id}", vars, tt.requestBody)
			w := httptest.NewRecorder()

			// Execute request
			handler.HandleUpdateTenant(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleDeleteTenant(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		mockSetup      func(*MockTenantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success - Delete tenant",
			tenantID: "123",
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				mockService.On("DeleteTenant", mock.Anything, "123").
					Return(diag)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Tenant deleted successfully"`,
		},
		{
			name:           "Error - Missing tenant ID",
			tenantID:       "",
			mockSetup:      func(mockService *MockTenantService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"error"`,
		},
		{
			name:     "Error - Service error",
			tenantID: "123",
			mockSetup: func(mockService *MockTenantService) {
				diag := diagnostics.New("test")
				diag.AddError("test_error", "test error", "test", nil)
				mockService.On("DeleteTenant", mock.Anything, "123").
					Return(diag)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			// Setup mock expectations
			tt.mockSetup(mockService)

			// Create request with URL variables
			vars := map[string]string{}
			if tt.tenantID != "" {
				vars["id"] = tt.tenantID
			}
			req := createTestRequestWithVars(http.MethodDelete, "/v1/tenants/{id}", vars, nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.HandleDeleteTenant(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
			mockService.AssertExpectations(t)
		})
	}
}

// TestHandleUpdateTenantInvalidJSON tests the update tenant handler with invalid JSON
func TestHandleUpdateTenantInvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/123", bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "123"})

	// Add context with appctx
	ctx := appctx.NewContext(nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute request
	handler.HandleUpdateTenant(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
}

// TestRoutes tests that all routes are properly registered
func TestRoutes(t *testing.T) {
	handler, _ := setupTestHandler()
	routes := handler.Routes()

	expectedRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/tenants"},
		{http.MethodGet, "/v1/tenants/{id}"},
		{http.MethodPost, "/v1/tenants"},
		{http.MethodPut, "/v1/tenants/{id}"},
		{http.MethodDelete, "/v1/tenants/{id}"},
	}

	assert.Equal(t, len(expectedRoutes), len(routes))

	for i, expected := range expectedRoutes {
		assert.Equal(t, expected.method, routes[i].Method)
		assert.Equal(t, expected.path, routes[i].Path)

	}
}

// TestIntegrationWithRouter tests the handler with a real router
func TestIntegrationWithRouter(t *testing.T) {
	handler, mockService := setupTestHandler()

	// Setup mock expectations
	expectedTenant := &models.Tenant{
		ID:     "123",
		Name:   "Test Tenant",
		Domain: "test.com",
	}
	diag := diagnostics.New("test")
	mockService.On("GetTenantByIDOrSlug", mock.Anything, "123").
		Return(expectedTenant, diag)

	// Create router and register routes
	router := mux.NewRouter()
	for _, route := range handler.Routes() {
		router.HandleFunc(route.Path, route.Handler).Methods(route.Method)
	}

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Make request
	req := createTestRequest(http.MethodGet, "/v1/tenants/123", nil)
	req.URL.Host = server.URL[7:] // Remove "http://" prefix
	req.URL.Scheme = "http"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"123"`)
	mockService.AssertExpectations(t)
}
