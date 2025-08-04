package models

import (
	"testing"
)

func TestAccessLevel_IsParentOf(t *testing.T) {
	tests := []struct {
		name     string
		parent   AccessLevel
		child    AccessLevel
		expected bool
	}{
		{"All is parent of Write", AccessLevelAll, AccessLevelWrite, true},
		{"All is parent of Read", AccessLevelAll, AccessLevelRead, true},
		{"All is parent of View", AccessLevelAll, AccessLevelView, true},
		{"Write is parent of Read", AccessLevelWrite, AccessLevelRead, true},
		{"Write is parent of Update", AccessLevelWrite, AccessLevelUpdate, true},
		{"Write is parent of Create", AccessLevelWrite, AccessLevelCreate, true},
		{"Update is parent of Read", AccessLevelUpdate, AccessLevelRead, true},
		{"Read is parent of View", AccessLevelRead, AccessLevelView, true},
		{"View is parent of None", AccessLevelView, AccessLevelNone, true},
		{"Read is not parent of Write", AccessLevelRead, AccessLevelWrite, false},
		{"View is not parent of Read", AccessLevelView, AccessLevelRead, false},
		{"None is not parent of anything", AccessLevelNone, AccessLevelView, false},
		{"All is not parent of itself", AccessLevelAll, AccessLevelAll, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parent.IsParentOf(tt.child)
			if result != tt.expected {
				t.Errorf("IsParentOf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_CanAccess(t *testing.T) {
	tests := []struct {
		name     string
		have     AccessLevel
		required AccessLevel
		expected bool
	}{
		{"All can access everything", AccessLevelAll, AccessLevelWrite, true},
		{"All can access itself", AccessLevelAll, AccessLevelAll, true},
		{"Write can access Read", AccessLevelWrite, AccessLevelRead, true},
		{"Write can access Update", AccessLevelWrite, AccessLevelUpdate, true},
		{"Write can access Create", AccessLevelWrite, AccessLevelCreate, true},
		{"Update can access Read", AccessLevelUpdate, AccessLevelRead, true},
		{"Read can access View", AccessLevelRead, AccessLevelView, true},
		{"Same level can access itself", AccessLevelRead, AccessLevelRead, true},
		{"Read cannot access Write", AccessLevelRead, AccessLevelWrite, false},
		{"View cannot access Read", AccessLevelView, AccessLevelRead, false},
		{"None cannot access anything", AccessLevelNone, AccessLevelView, false},
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
