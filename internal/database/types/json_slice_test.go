package types

import (
	"testing"
)

// Example custom type to demonstrate the generic JSONSlice and JSONObject
type ExampleStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Type aliases using the generic JSONSlice and JSONObject
type (
	ExampleStructSlice  = JSONSlice[ExampleStruct]
	ExampleStructObject = JSONObject[ExampleStruct]
)

func TestJSONSlice_MarshalUnmarshal(t *testing.T) {
	// Test with string slice
	stringSlice := JSONSlice[string]{"hello", "world"}

	// Test Value() method (marshaling)
	value, err := stringSlice.Value()
	if err != nil {
		t.Fatalf("Failed to marshal string slice: %v", err)
	}

	// Test Scan() method (unmarshaling)
	var newStringSlice JSONSlice[string]
	err = newStringSlice.Scan(value)
	if err != nil {
		t.Fatalf("Failed to unmarshal string slice: %v", err)
	}

	// Verify the data is preserved
	if len(newStringSlice) != len(stringSlice) {
		t.Errorf("Expected length %d, got %d", len(stringSlice), len(newStringSlice))
	}

	for i, v := range stringSlice {
		if newStringSlice[i] != v {
			t.Errorf("Expected %s at index %d, got %s", v, i, newStringSlice[i])
		}
	}
}

func TestJSONSlice_CustomStruct(t *testing.T) {
	// Test with custom struct slice
	exampleSlice := ExampleStructSlice{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	// Test Value() method (marshaling)
	value, err := exampleSlice.Value()
	if err != nil {
		t.Fatalf("Failed to marshal example struct slice: %v", err)
	}

	// Test Scan() method (unmarshaling)
	var newExampleSlice ExampleStructSlice
	err = newExampleSlice.Scan(value)
	if err != nil {
		t.Fatalf("Failed to unmarshal example struct slice: %v", err)
	}

	// Verify the data is preserved
	if len(newExampleSlice) != len(exampleSlice) {
		t.Errorf("Expected length %d, got %d", len(exampleSlice), len(newExampleSlice))
	}

	for i, v := range exampleSlice {
		if newExampleSlice[i].Name != v.Name || newExampleSlice[i].Age != v.Age {
			t.Errorf("Expected %+v at index %d, got %+v", v, i, newExampleSlice[i])
		}
	}
}

func TestJSONSlice_NilHandling(t *testing.T) {
	// Test nil slice
	var nilSlice JSONSlice[string]

	// Test Value() with nil
	value, err := nilSlice.Value()
	if err != nil {
		t.Fatalf("Failed to marshal nil slice: %v", err)
	}
	if value != nil {
		t.Errorf("Expected nil value for nil slice, got %v", value)
	}

	// Test Scan() with nil
	err = nilSlice.Scan(nil)
	if err != nil {
		t.Fatalf("Failed to scan nil value: %v", err)
	}
	if nilSlice != nil {
		t.Errorf("Expected nil slice after scanning nil, got %v", nilSlice)
	}
}

func TestJSONSlice_EmptySlice(t *testing.T) {
	// Test empty slice
	emptySlice := JSONSlice[string]{}

	// Test Value() with empty slice
	value, err := emptySlice.Value()
	if err != nil {
		t.Fatalf("Failed to marshal empty slice: %v", err)
	}

	// Should marshal to "[]"
	expected := "[]"
	if string(value.([]byte)) != expected {
		t.Errorf("Expected %s, got %s", expected, string(value.([]byte)))
	}

	// Test Scan() with empty slice JSON
	var newEmptySlice JSONSlice[string]
	err = newEmptySlice.Scan([]byte("[]"))
	if err != nil {
		t.Fatalf("Failed to unmarshal empty slice: %v", err)
	}

	if len(newEmptySlice) != 0 {
		t.Errorf("Expected empty slice, got slice with length %d", len(newEmptySlice))
	}
}

// Demonstrate how easy it is to add new JSON slice types
func TestJSONSlice_AddingNewTypes(t *testing.T) {
	// Just add a type alias - no need to implement Value() and Scan() methods!
	type (
		IntSlice          = JSONSlice[int]
		FloatSlice        = JSONSlice[float64]
		BoolSlice         = JSONSlice[bool]
		CustomStructSlice = JSONSlice[ExampleStruct]
	)

	// These types automatically have JSON marshaling/unmarshaling capabilities
	intSlice := IntSlice{1, 2, 3}
	floatSlice := FloatSlice{1.1, 2.2, 3.3}
	boolSlice := BoolSlice{true, false, true}
	customSlice := CustomStructSlice{{Name: "Test", Age: 42}}

	// Test that they can be marshaled
	_, err1 := intSlice.Value()
	if err1 != nil {
		t.Errorf("Failed to marshal int slice: %v", err1)
	}

	_, err2 := floatSlice.Value()
	if err2 != nil {
		t.Errorf("Failed to marshal float slice: %v", err2)
	}

	_, err3 := boolSlice.Value()
	if err3 != nil {
		t.Errorf("Failed to marshal bool slice: %v", err3)
	}

	_, err4 := customSlice.Value()
	if err4 != nil {
		t.Errorf("Failed to marshal custom struct slice: %v", err4)
	}
}
