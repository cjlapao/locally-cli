package types

import "testing"

func TestJSONObject_MarshalUnmarshal(t *testing.T) {
	// Test with a single struct
	exampleObject := ExampleStructObject{}
	exampleObject.Set(ExampleStruct{Name: "Alice", Age: 30})

	// Test Value() method (marshaling)
	value, err := exampleObject.Value()
	if err != nil {
		t.Fatalf("Failed to marshal example object: %v", err)
	}

	// Test Scan() method (unmarshaling)
	var newExampleObject ExampleStructObject
	err = newExampleObject.Scan(value)
	if err != nil {
		t.Fatalf("Failed to unmarshal example object: %v", err)
	}

	// Verify the data is preserved
	original := exampleObject.Get()
	unmarshaled := newExampleObject.Get()
	if original.Name != unmarshaled.Name || original.Age != unmarshaled.Age {
		t.Errorf("Expected %+v, got %+v", original, unmarshaled)
	}
}

func TestJSONObject_NilHandling(t *testing.T) {
	// Test nil value
	var nilObject ExampleStructObject

	// Test Scan() with nil
	err := nilObject.Scan(nil)
	if err != nil {
		t.Fatalf("Failed to scan nil value: %v", err)
	}
}

func TestJSONObject_EmptyStruct(t *testing.T) {
	// Test empty struct
	emptyObject := ExampleStructObject{}
	emptyObject.Set(ExampleStruct{})

	// Test Value() with empty struct
	value, err := emptyObject.Value()
	if err != nil {
		t.Fatalf("Failed to marshal empty object: %v", err)
	}

	// Should marshal to {"name":"","age":0} for zero values
	expected := `{"name":"","age":0}`
	if string(value.([]byte)) != expected {
		t.Errorf("Expected %s, got %s", expected, string(value.([]byte)))
	}
}
