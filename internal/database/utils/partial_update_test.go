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

func TestPartialUpdateMap_ReproduceIssue(t *testing.T) {
	// Test case to reproduce the issue where a changed field is not detected
	original := &TestUser{
		ID:        "123",
		Name:      "test tenant",
		Email:     "test@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updated := &TestUser{
		ID:        "123",
		Name:      "test tenant changed", // Changed from "test tenant"
		Email:     "test@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	t.Logf("Updates map: %+v", updates)

	// Should include the name field since it changed
	if updates["name"] != "test tenant changed" {
		t.Errorf("Expected name to be 'test tenant changed', got %v", updates["name"])
	}

	// Should have at least 2 fields (name + updated_at)
	if len(updates) < 2 {
		t.Errorf("Expected at least 2 updates, got %d", len(updates))
	}
}

func TestPartialUpdateMap_NoJsonTag(t *testing.T) {
	// Test with a field that has no json tag (like the Name field in User entity)
	type BaseModel struct {
		ID        string    `json:"id"`
		Slug      string    `json:"slug"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	type User struct {
		BaseModel
		Name     string `gorm:"not null;type:text"` // No json tag!
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	original := &User{
		BaseModel: BaseModel{
			ID:        "123",
			Slug:      "test-tenant",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "test tenant",
		Username: "testuser",
		Email:    "test@example.com",
	}

	updated := &User{
		BaseModel: BaseModel{
			ID:        "123",
			Slug:      "test-tenant",
			CreatedAt: original.CreatedAt,
			UpdatedAt: original.UpdatedAt,
		},
		Name:     "test tenant changed", // Changed from "test tenant"
		Username: "testuser",
		Email:    "test@example.com",
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	// The Name field should now be included using the field name as fallback
	if updates["Name"] != "test tenant changed" {
		t.Errorf("Expected Name to be 'test tenant changed', got %v", updates["Name"])
	}

	// Should have at least 2 fields (Name + updated_at)
	if len(updates) < 2 {
		t.Errorf("Expected at least 2 updates, got %d", len(updates))
	}
}

func TestPartialUpdateMap_ManyToManyRelationships(t *testing.T) {
	// Test that many-to-many relationships are handled correctly
	type BaseModel struct {
		ID        string    `json:"id"`
		Slug      string    `json:"slug"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	type Role struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type Claim struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type User struct {
		BaseModel
		Name     string  `json:"name"`
		Username string  `json:"username"`
		Email    string  `json:"email"`
		Roles    []Role  `json:"roles"`
		Claims   []Claim `json:"claims"`
	}

	// Create original user with roles and claims
	original := &User{
		BaseModel: BaseModel{
			ID:        "123",
			Slug:      "test-user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Test User",
		Username: "testuser",
		Email:    "test@example.com",
		Roles: []Role{
			{ID: "role1", Name: "admin"},
			{ID: "role2", Name: "user"},
		},
		Claims: []Claim{
			{ID: "claim1", Name: "read"},
			{ID: "claim2", Name: "write"},
		},
	}

	// Create updated user with only password change (no role/claim changes)
	updated := &User{
		BaseModel: BaseModel{
			ID:        "123",
			Slug:      "test-user",
			CreatedAt: original.CreatedAt,
			UpdatedAt: original.UpdatedAt,
		},
		Name:     "Test User Updated", // Only this field changed
		Username: "testuser",
		Email:    "test@example.com",
		// Roles and Claims are not set, so they won't be included in updates
	}

	updates := PartialUpdateMap(original, updated, "updated_at")

	// Should only include the name field and updated_at, not roles or claims
	if updates["name"] != "Test User Updated" {
		t.Errorf("Expected name to be 'Test User Updated', got %v", updates["name"])
	}

	// Should NOT include roles or claims in updates
	if updates["roles"] != nil {
		t.Errorf("Expected roles to NOT be in updates, but got %v", updates["roles"])
	}

	if updates["claims"] != nil {
		t.Errorf("Expected claims to NOT be in updates, but got %v", updates["claims"])
	}

	// Should have at least 2 fields (name + updated_at)
	if len(updates) < 2 {
		t.Errorf("Expected at least 2 updates, got %d", len(updates))
	}
}
