package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthDataStore implements AuthDataStoreInterface for testing
type MockAuthDataStore struct {
	*mocks.BaseMockStore
}

var authConfig = AuthServiceConfig{
	SecretKey: []byte("test-secret"),
	Issuer:    "test-issuer",
}

func setupAuthServiceWithMockStore(mockStore *MockAuthDataStore) *AuthService {
	Reset() // Reset singleton for test isolation
	Initialize(authConfig, mockStore, mockStore, mockStore)
	return GetInstance()
}

func TestGenerateSecureAPIKey(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	key1, errDiag := service.GenerateSecureAPIKey()
	assert.NotEmpty(t, key1)
	assert.False(t, errDiag.HasErrors())
	assert.True(t, len(key1) > 32)
	assert.True(t, len(key1) > len("sk-locally-"))
	assert.Equal(t, "sk-locally-", key1[:len("sk-locally-")])

	key2, errDiag := service.GenerateSecureAPIKey()
	assert.NotEmpty(t, key2)
	assert.False(t, errDiag.HasErrors())
	assert.NotEqual(t, key1, key2)
}

func TestCreateAPIKey(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	userID := "test-user-id"
	req := CreateAPIKeyRequest{
		Name: "Test API Key",
		Permissions: entities.APIKeyPermissions{
			Read:   true,
			Write:  false,
			Delete: false,
			Admin:  false,
			Scopes: []string{"users"},
		},
		TenantID:  "test-tenant",
		ExpiresAt: &time.Time{},
	}

	ctx := appctx.NewContext(context.Background())
	mockStore.On("CreateAPIKey", ctx, mock.AnythingOfType("*entities.APIKey")).Return(
		&entities.APIKey{
			BaseModel: entities.BaseModel{
				ID:        "test-id",
				CreatedAt: time.Now(),
			},
			UserID:      userID,
			Name:        req.Name,
			KeyPrefix:   "sk-locally-test",
			Permissions: `{"read":true,"write":false,"delete":false,"admin":false,"scopes":["users"]}`,
			TenantID:    req.TenantID,
			ExpiresAt:   req.ExpiresAt,
			IsActive:    true,
			CreatedBy:   userID,
		}, nil)

	response, err := service.CreateAPIKey(ctx, userID, req, userID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Name, response.Name)
	assert.NotEmpty(t, response.APIKey)
	assert.True(t, len(response.APIKey) > len("sk-locally-"))
	assert.Equal(t, "sk-locally-", response.APIKey[:len("sk-locally-")])

	mockStore.AssertExpectations(t)
}

func TestListAPIKeys(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	userID := "test-user-id"
	now := time.Now()

	testAPIKeys := []entities.APIKey{
		{
			BaseModel: entities.BaseModel{
				ID:        "key1",
				CreatedAt: now,
			},
			UserID:      userID,
			Name:        "Test Key 1",
			KeyPrefix:   "sk-locally-abc123",
			Permissions: `{"read":true,"write":false,"delete":false,"admin":false,"scopes":["users"]}`,
			TenantID:    "test-tenant",
			IsActive:    true,
			CreatedBy:   userID,
		},
		{
			BaseModel: entities.BaseModel{
				ID:        "key2",
				CreatedAt: now,
			},
			UserID:      userID,
			Name:        "Test Key 2",
			KeyPrefix:   "sk-locally-def456",
			Permissions: `{"read":true,"write":true,"delete":false,"admin":false,"scopes":["admin"]}`,
			TenantID:    "test-tenant",
			IsActive:    true,
			CreatedBy:   userID,
		},
	}

	ctx := appctx.NewContext(context.Background())
	mockStore.On("ListAPIKeysByUserID", ctx, userID).Return(testAPIKeys, nil)

	response, err := service.ListAPIKeys(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(2), response.Total)
	assert.Len(t, response.APIKeys, 2)
	assert.Equal(t, "Test Key 1", response.APIKeys[0].Name)
	assert.Equal(t, "Test Key 2", response.APIKeys[1].Name)

	mockStore.AssertExpectations(t)
}

func TestRevokeAPIKey(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	apiKeyID := "test-api-key-id"
	revokedBy := "test-user"
	reason := "Test revocation"

	ctx := appctx.NewContext(context.Background())
	mockStore.On("RevokeAPIKey", ctx, apiKeyID, revokedBy, reason).Return(nil)

	err := service.RevokeAPIKey(ctx, apiKeyID, revokedBy, reason)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

func TestDeleteAPIKey(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	apiKeyID := "test-api-key-id"

	ctx := appctx.NewContext(context.Background())
	mockStore.On("DeleteAPIKey", ctx, apiKeyID).Return(nil)

	err := service.DeleteAPIKey(ctx, apiKeyID)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

func TestAuthService_GenerateSecureAPIKey(t *testing.T) {
	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
	service := setupAuthServiceWithMockStore(mockStore)

	// Test multiple key generations to ensure uniqueness
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, errDiag := service.GenerateSecureAPIKey()
		assert.False(t, errDiag.HasErrors())
		assert.NotEmpty(t, key)
		assert.True(t, len(key) > 32)
		assert.True(t, len(key) > len("sk-locally-"))
		assert.Equal(t, "sk-locally-", key[:len("sk-locally-")])

		// Check for uniqueness
		assert.False(t, keys[key], "Generated duplicate key: %s", key)
		keys[key] = true
	}
}

func TestAPIKeyPermissions_MarshalUnmarshal(t *testing.T) {
	permissions := entities.APIKeyPermissions{
		Read:   true,
		Write:  false,
		Delete: true,
		Admin:  false,
		Scopes: []string{"users", "admin"},
	}

	// Test marshaling
	jsonData, err := json.Marshal(permissions)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test unmarshaling
	var unmarshaledPermissions entities.APIKeyPermissions
	err = json.Unmarshal(jsonData, &unmarshaledPermissions)
	assert.NoError(t, err)

	// Verify the data is preserved
	assert.Equal(t, permissions.Read, unmarshaledPermissions.Read)
	assert.Equal(t, permissions.Write, unmarshaledPermissions.Write)
	assert.Equal(t, permissions.Delete, unmarshaledPermissions.Delete)
	assert.Equal(t, permissions.Admin, unmarshaledPermissions.Admin)
	assert.Equal(t, permissions.Scopes, unmarshaledPermissions.Scopes)
}
