package service

import (
	"testing"

	"github.com/cjlapao/locally-cli/pkg/models"
)

func TestSystemService_GetName(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface to ensure we're using the interface
	systemService := service

	// Test GetName
	if systemService.GetName() != "system" {
		t.Errorf("Expected service name to be 'system', got '%s'", systemService.GetName())
	}
}

func TestSystemService_Initialize(t *testing.T) {
	ResetForTesting()

	// Initialize the service
	service := Initialize()

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	// Test singleton behavior
	service2 := Initialize()

	if service != service2 {
		t.Fatal("Singleton instances should be the same")
	}
}

func TestSystemService_GetService(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface
	systemServiceTest := service

	usersService := &models.ServiceDefinition{
		Name:        "users",
		Description: "User management service",
		Modules:     make(map[string]*models.ModuleDefinition),
	}

	systemServiceTest.AddService(usersService)

	// Get existing service
	retrievedService, exists := systemServiceTest.GetService("users")
	if !exists {
		t.Error("Existing service should be found")
	}

	if retrievedService == nil {
		t.Error("Retrieved service should not be nil")
	}

	// Get non-existent service
	_, exists = systemServiceTest.GetService("nonexistent")
	if exists {
		t.Error("Non-existent service should not be found")
	}
}

func TestSystemService_GenerateSystemClaims(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface
	systemServiceTest := service

	claims := systemServiceTest.GenerateSystemClaims()

	// Count the actual claims: wildcard + default system claims
	// Default system has: tenant(3), user(3), role(3), claim(3) = 12 claims
	// Plus wildcard = 13
	expectedCount := 17
	if len(claims) != expectedCount {
		t.Errorf("Expected %d claims, got %d", expectedCount, len(claims))
	}

	// Check for wildcard claim
	foundWildcard := false
	for _, claim := range claims {
		if claim.Slug == "*::*::*" {
			foundWildcard = true
			break
		}
	}

	if !foundWildcard {
		t.Error("Wildcard claim should be present")
	}
}

func TestSystemService_GenerateDefaultClaimsForSecurityLevel(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface
	systemServiceTest := service

	usersAPI := &models.ModuleDefinition{
		Name:        "api",
		Description: "API module",
		Actions:     []models.AccessLevel{models.AccessLevelRead, models.AccessLevelWrite},
	}

	usersView := &models.ModuleDefinition{
		Name:        "view",
		Description: "View module",
		Actions:     []models.AccessLevel{models.AccessLevelRead, models.AccessLevelView},
	}

	usersService := &models.ServiceDefinition{
		Name:        "users",
		Description: "User management service",
		Modules: map[string]*models.ModuleDefinition{
			"api":  usersAPI,
			"view": usersView,
		},
	}

	systemServiceTest.AddService(usersService)

	tests := []struct {
		name          string
		securityLevel models.SecurityLevel
		expectedCount int
	}{
		{
			name:          "SuperUser claims",
			securityLevel: models.SecurityLevelSuperUser,
			expectedCount: 1, // Single wildcard claim
		},
		{
			name:          "Manager claims",
			securityLevel: models.SecurityLevelManager,
			expectedCount: 10, // Default system (8) + users::api::write, users::view::read
		},
		{
			name:          "User claims",
			securityLevel: models.SecurityLevelUser,
			expectedCount: 10, // Default system (8) + users::api::read, users::view::read
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := systemServiceTest.GenerateDefaultClaimsForSecurityLevel(tt.securityLevel)
			if len(claims) != tt.expectedCount {
				t.Errorf("Expected %d claims, got %d", tt.expectedCount, len(claims))
			}
		})
	}
}

func TestSystemService_ValidateClaim(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface
	systemServiceTest := service

	usersAPI := &models.ModuleDefinition{
		Name:        "api",
		Description: "API module",
		Actions:     []models.AccessLevel{models.AccessLevelRead, models.AccessLevelWrite},
	}

	usersService := &models.ServiceDefinition{
		Name:        "users",
		Description: "User management service",
		Modules: map[string]*models.ModuleDefinition{
			"api": usersAPI,
		},
	}

	systemServiceTest.AddService(usersService)

	// Valid claim
	validClaim := &models.Claim{
		Service: "users",
		Module:  "api",
		Action:  models.AccessLevelRead,
	}

	err := systemServiceTest.ValidateClaim(validClaim)
	if err != nil {
		t.Errorf("Valid claim should not return error: %v", err)
	}

	// Invalid claim - non-existent service
	invalidClaim := &models.Claim{
		Service: "nonexistent",
		Module:  "api",
		Action:  models.AccessLevelRead,
	}

	err = systemServiceTest.ValidateClaim(invalidClaim)
	if err == nil {
		t.Error("Invalid claim should return error")
	}
}

func TestSystemService_ParseClaim(t *testing.T) {
	ResetForTesting()
	service := Initialize()

	// Explicitly cast to interface
	systemServiceTest := service

	usersAPI := &models.ModuleDefinition{
		Name:        "api",
		Description: "API module",
		Actions:     []models.AccessLevel{models.AccessLevelRead, models.AccessLevelWrite},
	}

	usersService := &models.ServiceDefinition{
		Name:        "users",
		Description: "User management service",
		Modules: map[string]*models.ModuleDefinition{
			"api": usersAPI,
		},
	}

	systemServiceTest.AddService(usersService)

	// Valid claim
	claim, err := systemServiceTest.ParseClaim("users::api::read")
	if err != nil {
		t.Errorf("Valid claim should not return error: %v", err)
	}

	if claim.Service != "users" || claim.Module != "api" || claim.Action != models.AccessLevelRead {
		t.Error("Parsed claim should have correct values")
	}

	// Invalid claim - non-existent action
	_, err = systemServiceTest.ParseClaim("users::api::delete")
	if err == nil {
		t.Error("Invalid claim should return error")
	}
}
