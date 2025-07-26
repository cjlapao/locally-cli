package utils

import (
	"reflect"
	"testing"
	"time"
)

// TestUser represents a user entity for testing
type TestUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Metadata  string    `json:"metadata"`
}

func TestPartialUpdateMap_StringFields(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  `{"key": "value"}`,
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "Jane Doe",         // Changed
		Email:     "john@example.com", // Same
		Age:       30,                 // Same
		Active:    true,               // Same
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
		Metadata:  `{"key": "value"}`, // Same
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	// Should only include changed fields
	if len(updates) != 2 { // name + updated_at
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}

	if updates["name"] != "Jane Doe" {
		t.Errorf("Expected name to be 'Jane Doe', got %v", updates["name"])
	}

	if _, exists := updates["email"]; exists {
		t.Error("Email should not be in updates as it didn't change")
	}
}

func TestPartialUpdateMap_EmptyFields(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  `{"key": "value"}`,
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "", // Empty - should not update
		Email:     "", // Empty - should not update
		Age:       0,  // Zero - should not update
		Active:    true,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
		Metadata:  "", // Empty - should not update
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	// Should only include updated_at since all other fields are empty/zero
	if len(updates) != 1 {
		t.Errorf("Expected 1 update (updated_at), got %d", len(updates))
	}

	if _, exists := updates["name"]; exists {
		t.Error("Name should not be in updates as it's empty")
	}

	if _, exists := updates["email"]; exists {
		t.Error("Email should not be in updates as it's empty")
	}

	if _, exists := updates["age"]; exists {
		t.Error("Age should not be in updates as it's zero")
	}
}

func TestPartialUpdateMap_IntFields(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       35, // Changed
		Active:    true,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	if updates["age"] != 35 {
		t.Errorf("Expected age to be 35, got %v", updates["age"])
	}
}

func TestPartialUpdateMap_BoolFields(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    false, // Changed
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	if updates["active"] != false {
		t.Errorf("Expected active to be false, got %v", updates["active"])
	}
}

func TestPartialUpdateMap_AlwaysUpdateFields(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "John Doe", // No changes
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	// Should include updated_at even though no other fields changed
	if len(updates) != 1 {
		t.Errorf("Expected 1 update (updated_at), got %d", len(updates))
	}

	if _, exists := updates["updated_at"]; !exists {
		t.Error("updated_at should be in updates")
	}
}

func TestPartialUpdateMap_NoChanges(t *testing.T) {
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	updates := PartialUpdateMap(original, updated)

	// Should be empty since no fields changed
	if len(updates) != 0 {
		t.Errorf("Expected 0 updates, got %d", len(updates))
	}
}

func TestPartialUpdateMap_DifferentTypes(t *testing.T) {
	// Test with different types - should fall back to zero-value approach
	original := &TestUser{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Pass a different type - should fall back to zero-value approach
	updates := PartialUpdateMap(original, "not a struct")

	// Should be empty since "not a struct" has no fields
	if len(updates) != 0 {
		t.Errorf("Expected 0 updates for different types, got %d", len(updates))
	}
}

func TestPartialUpdateMap_PointerFields(t *testing.T) {
	type TestStruct struct {
		ID   string  `json:"id"`
		Name string  `json:"name"`
		Ptr  *string `json:"ptr"`
	}

	ptr1 := "value1"
	ptr2 := "value2"

	original := &TestStruct{
		ID:   "123",
		Name: "John",
		Ptr:  &ptr1,
	}

	updated := &TestStruct{
		ID:   "123",
		Name: "John",
		Ptr:  &ptr2, // Different pointer value
	}

	updates := PartialUpdateMap(original, updated)

	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got %d", len(updates))
	}

	if updates["ptr"] != &ptr2 {
		t.Errorf("Expected ptr to be updated")
	}
}

func TestPartialUpdateMap_NilPointers(t *testing.T) {
	type TestStruct struct {
		ID   string  `json:"id"`
		Name string  `json:"name"`
		Ptr  *string `json:"ptr"`
	}

	ptr1 := "value1"

	original := &TestStruct{
		ID:   "123",
		Name: "John",
		Ptr:  &ptr1,
	}

	updated := &TestStruct{
		ID:   "123",
		Name: "John",
		Ptr:  nil, // Changed from pointer to nil
	}

	updates := PartialUpdateMap(original, updated)

	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got %d", len(updates))
	}

	ptrValue := updates["ptr"]
	if ptrValue != nil && !reflect.ValueOf(ptrValue).IsNil() {
		t.Errorf("Expected ptr to be nil, got %v (type: %T)", ptrValue, ptrValue)
	}
}
