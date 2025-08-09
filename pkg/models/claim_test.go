package models

import (
	"testing"
)

func TestParseClaim(t *testing.T) {
	tests := []struct {
		name        string
		slug        string
		expectError bool
		expected    *Claim
	}{
		{
			name:        "Valid claim",
			slug:        "users::api::read",
			expectError: false,
			expected: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
				Slug:    "users::api::read",
			},
		},
		{
			name:        "Valid claim with write",
			slug:        "users::api::write",
			expectError: false,
			expected: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
				Slug:    "users::api::write",
			},
		},
		{
			name:        "Valid claim with all",
			slug:        "users::api::*",
			expectError: false,
			expected: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelAll,
				Slug:    "users::api::*",
			},
		},
		{
			name:        "Invalid format - missing parts",
			slug:        "users::api",
			expectError: true,
		},
		{
			name:        "Invalid format - too many parts",
			slug:        "users::api::read::extra",
			expectError: true,
		},
		{
			name:        "Invalid action",
			slug:        "users::api::invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseClaim(tt.slug)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseClaim() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseClaim() unexpected error: %v", err)
				return
			}

			if result.Service != tt.expected.Service {
				t.Errorf("Service = %v, want %v", result.Service, tt.expected.Service)
			}
			if result.Module != tt.expected.Module {
				t.Errorf("Module = %v, want %v", result.Module, tt.expected.Module)
			}
			if result.Action != tt.expected.Action {
				t.Errorf("Action = %v, want %v", result.Action, tt.expected.Action)
			}
			if result.Slug != tt.expected.Slug {
				t.Errorf("Slug = %v, want %v", result.Slug, tt.expected.Slug)
			}
		})
	}
}

func TestClaim_CanAccess(t *testing.T) {
	tests := []struct {
		name     string
		have     *Claim
		required *Claim
		expected bool
	}{
		{
			name: "Same service and module, higher action",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
		{
			name: "Same service and module, same action",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			required: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
		{
			name: "Same service and module, lower action",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			required: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			expected: false,
		},
		{
			name: "Different service",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "admin",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: false,
		},
		{
			name: "Different module",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "users",
				Module:  "view",
				Action:  AccessLevelRead,
			},
			expected: false,
		},
		{
			name: "Wildcard service in required",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "*",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
		{
			name: "Wildcard module in required",
			have: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "users",
				Module:  "*",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
		{
			name: "Wildcard service in have",
			have: &Claim{
				Service: "*",
				Module:  "api",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
		{
			name: "Wildcard module in have",
			have: &Claim{
				Service: "users",
				Module:  "*",
				Action:  AccessLevelWrite,
			},
			required: &Claim{
				Service: "users",
				Module:  "api",
				Action:  AccessLevelRead,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.have.CanAccess(tt.required)
			if result != tt.expected {
				t.Errorf("CanAccess() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClaim_GetSlug(t *testing.T) {
	claim := &Claim{
		Service: "users",
		Module:  "api",
		Action:  AccessLevelRead,
	}

	expected := "users::api::read"
	result := claim.GetSlug()

	if result != expected {
		t.Errorf("GetSlug() = %v, want %v", result, expected)
	}
}

func TestClaim_GetService(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		expected string
	}{
		{"Normal service", "users", "users"},
		{"Wildcard service", "*", "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &Claim{Service: tt.service}
			result := claim.GetService()
			if result != tt.expected {
				t.Errorf("GetService() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClaim_GetModule(t *testing.T) {
	tests := []struct {
		name     string
		module   string
		expected string
	}{
		{"Normal module", "api", "api"},
		{"Wildcard module", "*", "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &Claim{Module: tt.module}
			result := claim.GetModule()
			if result != tt.expected {
				t.Errorf("GetModule() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClaim_GetAction(t *testing.T) {
	tests := []struct {
		name     string
		action   AccessLevel
		expected string
	}{
		{"Normal action", AccessLevelRead, "read"},
		{"All action", AccessLevelAll, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &Claim{Action: tt.action}
			result := claim.GetAction()
			if result != tt.expected {
				t.Errorf("GetAction() = %v, want %v", result, tt.expected)
			}
		})
	}
}
