package validation

import (
	"testing"
)

// Test structs to demonstrate nested validation
type TestService struct {
	ServiceName     string `json:"service_name" validate:"required"`
	ContainerConfig string `json:"container_config" validate:"required"`
	Parameters      string `json:"parameters" validate:"required"`
	PortMappings    string `json:"port_mappings" validate:"required"`
	Volumes         string `json:"volumes" validate:"required"`
}

type TestFile struct {
	FileName string `json:"file_name" validate:"required"`
	Path     string `json:"path" validate:"required"`
	Content  []byte `json:"content" validate:"required"`
	UID      int    `json:"uid" validate:"required"`
	GID      int    `json:"gid" validate:"required"`
	Mode     int    `json:"mode" validate:"required"`
}

type TestBlueprint struct {
	Name     string        `json:"name" validate:"required"`
	Type     string        `json:"type" validate:"required"`
	Version  string        `json:"version" validate:"required,versionformat"`
	Services []TestService `json:"services" validate:"required"`
	Files    []TestFile    `json:"files" validate:"required"`
}

func TestValidateNestedStructs(t *testing.T) {
	// Test with valid data
	validBlueprint := TestBlueprint{
		Name:    "Test Blueprint",
		Type:    "docker-container",
		Version: "1.0.0",
		Services: []TestService{
			{
				ServiceName:     "web-service",
				ContainerConfig: `{"image": "nginx:latest"}`,
				Parameters:      `[{"key": "port", "value": "80"}]`,
				PortMappings:    `[{"port": 80, "map_to": 8080}]`,
				Volumes:         `[{"volume_path": "/data", "mount_at": "/app/data"}]`,
			},
		},
		Files: []TestFile{
			{
				FileName: "index.html",
				Path:     "/app/index.html",
				Content:  []byte("<html></html>"),
				UID:      1000,
				GID:      1000,
				Mode:     0o644,
			},
		},
	}

	errors := Validate(validBlueprint)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateNestedStructsWithErrors(t *testing.T) {
	// Test with invalid data - missing required fields in nested structs
	invalidBlueprint := TestBlueprint{
		Name:    "Test Blueprint",
		Type:    "docker-container",
		Version: "invalid-version", // Invalid version format
		Services: []TestService{
			{
				ServiceName: "", // Missing required field
				// Missing other required fields
			},
		},
		Files: []TestFile{
			{
				FileName: "", // Missing required field
				Path:     "/app/index.html",
				// Missing other required fields
			},
		},
	}

	errors := Validate(invalidBlueprint)
	if len(errors) == 0 {
		t.Error("Expected validation errors, got none")
		return
	}

	// Check for specific error paths
	expectedPaths := map[string]bool{
		"version":                      false, // Invalid version format
		"services[0].service_name":     false,
		"services[0].container_config": false,
		"services[0].parameters":       false,
		"services[0].port_mappings":    false,
		"services[0].volumes":          false,
		"files[0].file_name":           false,
		"files[0].content":             false,
		"files[0].uid":                 false,
		"files[0].gid":                 false,
		"files[0].mode":                false,
	}

	for _, err := range errors {
		if _, exists := expectedPaths[err.Field]; exists {
			expectedPaths[err.Field] = true
		}
		t.Logf("Validation error: %s - %s", err.Field, err.Message)
	}

	// Check that we got errors for the expected paths
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected validation error for field: %s", path)
		}
	}
}

func TestValidateEmptySlices(t *testing.T) {
	// Test with empty slices
	emptyBlueprint := TestBlueprint{
		Name:     "Test Blueprint",
		Type:     "docker-container",
		Version:  "1.0.0",
		Services: []TestService{}, // Empty slice
		Files:    []TestFile{},    // Empty slice
	}

	errors := Validate(emptyBlueprint)
	if len(errors) == 0 {
		t.Error("Expected validation errors for empty required slices, got none")
		return
	}

	// Should have errors for required slices
	foundServicesError := false
	foundFilesError := false

	for _, err := range errors {
		if err.Field == "services" {
			foundServicesError = true
		}
		if err.Field == "files" {
			foundFilesError = true
		}
	}

	if !foundServicesError {
		t.Error("Expected validation error for empty services slice")
	}
	if !foundFilesError {
		t.Error("Expected validation error for empty files slice")
	}
}

func TestValidateVersionFormat(t *testing.T) {
	// Test version format validation
	testCases := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"2.1.3", true},
		{"0.0.1", true},
		{"invalid", false},
		{"1.0", false},
		{"1.0.0.0", false},
		{"v1.0.0", false},
	}

	for _, tc := range testCases {
		blueprint := TestBlueprint{
			Name:     "Test Blueprint",
			Type:     "docker-container",
			Version:  tc.version,
			Services: []TestService{{ServiceName: "test", ContainerConfig: "{}", Parameters: "[]", PortMappings: "[]", Volumes: "[]"}},
			Files:    []TestFile{{FileName: "test", Path: "/test", Content: []byte("test"), UID: 1, GID: 1, Mode: 0o644}},
		}

		errors := Validate(blueprint)
		hasVersionError := false

		for _, err := range errors {
			if err.Field == "version" {
				hasVersionError = true
				break
			}
		}

		if tc.valid && hasVersionError {
			t.Errorf("Version '%s' should be valid but got validation error", tc.version)
		}
		if !tc.valid && !hasVersionError {
			t.Errorf("Version '%s' should be invalid but no validation error", tc.version)
		}
	}
}
