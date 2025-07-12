package utils

import (
	"testing"
	"time"
)

// TestStruct is a simple struct for testing
type TestStruct struct {
	Name    string    `json:"name"`
	Age     int       `json:"age"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

// TestStructWithJSON is a struct with JSON fields for testing
type TestStructWithJSON struct {
	Name            string `json:"name"`
	Parameters      string `json:"parameters" gorm:"type:json"`
	ContainerConfig string `json:"container_config" gorm:"type:json"`
	Files           string `json:"files" gorm:"type:json"`
	Active          bool   `json:"active"`
}

func TestPartialUpdateMap_Basic(t *testing.T) {
	entity := &TestStruct{
		Name:   "John",
		Age:    30,
		Active: true,
	}

	updates := PartialUpdateMap(entity, "updated_at")

	// Check that all non-zero fields are included
	if updates["name"] != "John" {
		t.Errorf("Expected name to be 'John', got %v", updates["name"])
	}
	if updates["age"] != 30 {
		t.Errorf("Expected age to be 30, got %v", updates["age"])
	}
	if updates["active"] != true {
		t.Errorf("Expected active to be true, got %v", updates["active"])
	}

	// Check that updated_at is included
	if _, exists := updates["updated_at"]; !exists {
		t.Error("Expected updated_at to be in updates")
	}

	// Check that created is not included (zero value)
	if _, exists := updates["created"]; exists {
		t.Error("Expected created to not be in updates (zero value)")
	}
}

func TestPartialUpdateMap_EmptyStruct(t *testing.T) {
	entity := &TestStruct{}

	updates := PartialUpdateMap(entity, "updated_at")

	// Should only contain updated_at
	if _, exists := updates["updated_at"]; !exists {
		t.Error("Expected updated_at to be in updates")
	}
	if len(updates) != 1 {
		t.Errorf("Expected only 1 field in updates, got %d", len(updates))
	}
}

func TestPartialUpdateMap_OnlyOneField(t *testing.T) {
	entity := &TestStruct{
		Name: "Jane",
	}

	updates := PartialUpdateMap(entity, "updated_at")

	// Should contain name and updated_at
	if updates["name"] != "Jane" {
		t.Errorf("Expected name to be 'Jane', got %v", updates["name"])
	}
	if _, exists := updates["updated_at"]; !exists {
		t.Error("Expected updated_at to be in updates")
	}

	// Should not contain other fields
	if _, exists := updates["age"]; exists {
		t.Error("Expected age to not be in updates")
	}
	if _, exists := updates["active"]; exists {
		t.Error("Expected active to not be in updates")
	}
}

func TestPartialUpdateMap_JSONFields(t *testing.T) {
	// Test with valid JSON values
	entity := &TestStructWithJSON{
		Name:            "Test",
		Parameters:      `[{"key": "value"}]`,
		ContainerConfig: `{"image": "ubuntu"}`,
		Files:           `[{"name": "test.txt"}]`,
		Active:          true,
	}

	updates := PartialUpdateMap(entity, "updated_at")

	// Should contain all fields
	expectedFields := []string{"name", "parameters", "container_config", "files", "active", "updated_at"}
	for _, field := range expectedFields {
		if _, exists := updates[field]; !exists {
			t.Errorf("Expected '%s' field to be in updates", field)
		}
	}
}

func TestPartialUpdateMap_EmptyJSONFields(t *testing.T) {
	// Test with empty JSON values - these should be skipped
	entity := &TestStructWithJSON{
		Name:            "Test",
		Parameters:      "[]", // Empty array
		ContainerConfig: "{}", // Empty object
		Files:           "",   // Empty string
		Active:          true,
	}

	updates := PartialUpdateMap(entity, "updated_at")

	// Should contain name, active, and updated_at
	expectedFields := []string{"name", "active", "updated_at"}
	for _, field := range expectedFields {
		if _, exists := updates[field]; !exists {
			t.Errorf("Expected '%s' field to be in updates", field)
		}
	}

	// Should not contain empty JSON fields
	emptyJSONFields := []string{"parameters", "container_config", "files"}
	for _, field := range emptyJSONFields {
		if _, exists := updates[field]; exists {
			t.Errorf("Expected '%s' field to not be in updates (empty JSON)", field)
		}
	}
}

func TestPartialUpdateMap_MixedJSONFields(t *testing.T) {
	// Test with mix of valid and empty JSON values
	entity := &TestStructWithJSON{
		Name:            "Test",
		Parameters:      `[{"key": "value"}]`, // Valid JSON
		ContainerConfig: "{}",                 // Empty JSON object
		Files:           "[]",                 // Empty JSON array
		Active:          true,
	}

	updates := PartialUpdateMap(entity, "updated_at")

	// Should contain valid fields
	expectedFields := []string{"name", "parameters", "active", "updated_at"}
	for _, field := range expectedFields {
		if _, exists := updates[field]; !exists {
			t.Errorf("Expected '%s' field to be in updates", field)
		}
	}

	// Should not contain empty JSON fields
	emptyJSONFields := []string{"container_config", "files"}
	for _, field := range emptyJSONFields {
		if _, exists := updates[field]; exists {
			t.Errorf("Expected '%s' field to not be in updates (empty JSON)", field)
		}
	}
}

func TestPartialUpdateMapWithCustomFields(t *testing.T) {
	entity := &TestStruct{
		Name: "John",
		Age:  25,
	}

	fieldMappings := map[string]string{
		"name": "user_name",
		"age":  "user_age",
	}

	updates := PartialUpdateMapWithCustomFields(entity, fieldMappings, "updated_at")

	// Check that custom field names are used
	if updates["user_name"] != "John" {
		t.Errorf("Expected user_name to be 'John', got %v", updates["user_name"])
	}
	if updates["user_age"] != 25 {
		t.Errorf("Expected user_age to be 25, got %v", updates["user_age"])
	}

	// Check that original field names are not used
	if _, exists := updates["name"]; exists {
		t.Error("Expected 'name' to not be in updates (should be 'user_name')")
	}
	if _, exists := updates["age"]; exists {
		t.Error("Expected 'age' to not be in updates (should be 'user_age')")
	}
}
