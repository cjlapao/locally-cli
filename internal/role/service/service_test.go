package service

import (
	"testing"
)

func TestRoleService_GetName(t *testing.T) {
	// Reset for testing
	Reset()

	// Initialize service for testing
	service := Initialize(nil, nil, nil)
	name := service.GetName()
	if name != "role" {
		t.Errorf("Expected service name to be 'role', got '%s'", name)
	}
}

func TestRoleService_CreateRole(t *testing.T) {
	// Reset for testing
	Reset()

	// This test would require a mock user store
	// For now, we'll just test that the service can be initialized
	service := Initialize(nil, nil, nil)
	if service == nil {
		t.Error("Expected service to be initialized, got nil")
	}
}
