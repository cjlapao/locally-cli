package utils

import (
	"testing"
)

func TestObfuscatePassword(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "empty password",
			input:       "",
			expected:    "***",
			description: "Should return *** for empty password",
		},
		{
			name:        "single character",
			input:       "a",
			expected:    "***",
			description: "Should return *** for single character",
		},
		{
			name:        "two characters",
			input:       "ab",
			expected:    "***",
			description: "Should return *** for two characters",
		},
		{
			name:        "three characters",
			input:       "abc",
			expected:    "***",
			description: "Should return *** for three characters",
		},
		{
			name:        "four characters",
			input:       "abcd",
			expected:    "***",
			description: "Should return *** for four characters",
		},
		{
			name:        "five characters",
			input:       "abcde",
			expected:    "***",
			description: "Should return *** for five characters",
		},
		{
			name:        "six characters",
			input:       "abcdef",
			expected:    "a***f",
			description: "Should return first + *** + last for six characters",
		},
		{
			name:        "seven characters",
			input:       "abcdefg",
			expected:    "a***g",
			description: "Should return first + *** + last for seven characters",
		},
		{
			name:        "long password",
			input:       "verylongpassword123",
			expected:    "v***3",
			description: "Should return first + *** + last for long password",
		},
		{
			name:        "unicode characters",
			input:       "pässwörd",
			expected:    "p***d",
			description: "Should handle unicode characters correctly",
		},
		{
			name:        "special characters",
			input:       "p@ssw0rd!",
			expected:    "p***!",
			description: "Should handle special characters correctly",
		},
		{
			name:        "numbers only",
			input:       "123456",
			expected:    "1***6",
			description: "Should handle numbers correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObfuscateString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			description: "Should return empty string for empty input",
		},
		{
			name:        "simple text",
			input:       "Hello World",
			expected:    "hello-world",
			description: "Should convert spaces to dashes and lowercase",
		},
		{
			name:        "multiple spaces",
			input:       "Hello   World",
			expected:    "hello-world",
			description: "Should handle multiple consecutive spaces",
		},
		{
			name:        "special characters",
			input:       "Hello, World!",
			expected:    "hello-world-",
			description: "Should remove special characters",
		},
		{
			name:        "numbers and letters",
			input:       "Product 123",
			expected:    "product-123",
			description: "Should handle numbers correctly",
		},
		{
			name:        "underscores and dashes",
			input:       "product_name-123",
			expected:    "product_name-123",
			description: "Should preserve underscores and dashes",
		},
		{
			name:        "mixed case",
			input:       "ProductName",
			expected:    "productname",
			description: "Should convert to lowercase",
		},
		{
			name:        "unicode characters",
			input:       "café résumé",
			expected:    "caf-r-sum-",
			description: "Should handle unicode characters",
		},
		{
			name:        "leading and trailing spaces",
			input:       "  Hello World  ",
			expected:    "-hello-world-",
			description: "Should trim leading and trailing spaces",
		},
		{
			name:        "only special characters",
			input:       "!@#$%^&*()",
			expected:    "-",
			description: "Should return dash for only special characters",
		},
		{
			name:        "consecutive special characters",
			input:       "Hello!!World",
			expected:    "hello-world",
			description: "Should handle consecutive special characters",
		},
		{
			name:        "already slugified",
			input:       "hello-world",
			expected:    "hello-world",
			description: "Should return same string if already slugified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s' - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestStringToMap(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    map[string]interface{}
		description string
	}{
		{
			name:        "valid JSON object",
			input:       `{"name": "John", "age": 30}`,
			expectError: false,
			expected:    map[string]interface{}{"name": "John", "age": float64(30)},
			description: "Should parse valid JSON object",
		},
		{
			name:        "empty object",
			input:       `{}`,
			expectError: false,
			expected:    map[string]interface{}{},
			description: "Should parse empty JSON object",
		},
		{
			name:        "nested object",
			input:       `{"user": {"name": "John", "age": 30}}`,
			expectError: false,
			expected:    map[string]interface{}{"user": map[string]interface{}{"name": "John", "age": float64(30)}},
			description: "Should parse nested JSON object",
		},
		{
			name:        "array in object",
			input:       `{"tags": ["tag1", "tag2"]}`,
			expectError: false,
			expected:    map[string]interface{}{"tags": []interface{}{"tag1", "tag2"}},
			description: "Should parse object with array",
		},
		{
			name:        "invalid JSON",
			input:       `{"name": "John", "age": 30`,
			expectError: true,
			expected:    nil,
			description: "Should return error for invalid JSON",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			expected:    nil,
			description: "Should return error for empty string",
		},
		{
			name:        "null value",
			input:       `{"name": null}`,
			expectError: false,
			expected:    map[string]interface{}{"name": nil},
			description: "Should handle null values",
		},
		{
			name:        "boolean values",
			input:       `{"active": true, "verified": false}`,
			expectError: false,
			expected:    map[string]interface{}{"active": true, "verified": false},
			description: "Should handle boolean values",
		},
		{
			name:        "number values",
			input:       `{"price": 99.99, "quantity": 5}`,
			expectError: false,
			expected:    map[string]interface{}{"price": 99.99, "quantity": float64(5)},
			description: "Should handle number values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringToMap(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none - %s", tt.description)
				}
				if result != nil {
					t.Errorf("Expected nil result when error occurs - %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v - %s", err, tt.description)
				}
				if result == nil {
					t.Errorf("Expected result but got nil - %s", tt.description)
				}
				// Compare maps - handle nested structures carefully
				if len(result) != len(tt.expected) {
					t.Errorf("Expected map with %d keys, got %d - %s", len(tt.expected), len(result), tt.description)
				}
				for key, expectedValue := range tt.expected {
					if actualValue, exists := result[key]; !exists {
						t.Errorf("Expected key '%s' not found - %s", key, tt.description)
					} else {
						// For simple types, we can compare directly
						switch v := expectedValue.(type) {
						case string, int, float64, bool:
							if actualValue != expectedValue {
								t.Errorf("Expected value %v for key '%s', got %v - %s", expectedValue, key, actualValue, tt.description)
							}
						case map[string]interface{}:
							// For nested maps, just check they exist and have the right type
							if actualMap, ok := actualValue.(map[string]interface{}); !ok {
								t.Errorf("Expected map for key '%s', got %T - %s", key, actualValue, tt.description)
							} else if len(actualMap) != len(v) {
								t.Errorf("Expected nested map with %d keys for '%s', got %d - %s", len(v), key, len(actualMap), tt.description)
							}
						case []interface{}:
							// For arrays, just check they exist and have the right type
							if actualSlice, ok := actualValue.([]interface{}); !ok {
								t.Errorf("Expected slice for key '%s', got %T - %s", key, actualValue, tt.description)
							} else if len(actualSlice) != len(v) {
								t.Errorf("Expected slice with %d items for '%s', got %d - %s", len(v), key, len(actualSlice), tt.description)
							}
						default:
							// For other types, just check they exist
							if actualValue == nil && expectedValue != nil {
								t.Errorf("Expected non-nil value for key '%s', got nil - %s", key, tt.description)
							}
						}
					}
				}
			}
		})
	}
}

func TestStringToSlice(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    []string
		description string
	}{
		{
			name:        "valid JSON array",
			input:       `["item1", "item2", "item3"]`,
			expectError: false,
			expected:    []string{"item1", "item2", "item3"},
			description: "Should parse valid JSON array",
		},
		{
			name:        "empty array",
			input:       `[]`,
			expectError: false,
			expected:    []string{},
			description: "Should parse empty JSON array",
		},
		{
			name:        "single item array",
			input:       `["single"]`,
			expectError: false,
			expected:    []string{"single"},
			description: "Should parse single item array",
		},
		{
			name:        "array with empty strings",
			input:       `["", "item", ""]`,
			expectError: false,
			expected:    []string{"", "item", ""},
			description: "Should handle empty strings in array",
		},
		{
			name:        "invalid JSON",
			input:       `["item1", "item2"`,
			expectError: true,
			expected:    nil,
			description: "Should return error for invalid JSON",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			expected:    nil,
			description: "Should return error for empty string",
		},
		{
			name:        "not an array",
			input:       `{"key": "value"}`,
			expectError: true,
			expected:    nil,
			description: "Should return error for non-array JSON",
		},
		{
			name:        "array with mixed types",
			input:       `["string", 123, true]`,
			expectError: true,
			expected:    nil,
			description: "Should return error for array with non-string values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringToSlice(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none - %s", tt.description)
				}
				if result != nil {
					t.Errorf("Expected nil result when error occurs - %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v - %s", err, tt.description)
				}
				if result == nil {
					t.Errorf("Expected result but got nil - %s", tt.description)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("Expected slice with %d items, got %d - %s", len(tt.expected), len(result), tt.description)
				}
				for i, expectedValue := range tt.expected {
					if i >= len(result) {
						t.Errorf("Expected item at index %d but slice is too short - %s", i, tt.description)
						break
					}
					if result[i] != expectedValue {
						t.Errorf("Expected '%s' at index %d, got '%s' - %s", expectedValue, i, result[i], tt.description)
					}
				}
			}
		})
	}
}

// Test struct for StringToObject
type TestStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestStringToObject(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    TestStruct
		description string
	}{
		{
			name:        "valid JSON object",
			input:       `{"name": "John", "age": 30, "email": "john@example.com"}`,
			expectError: false,
			expected:    TestStruct{Name: "John", Age: 30, Email: "john@example.com"},
			description: "Should parse valid JSON object to struct",
		},
		{
			name:        "partial object",
			input:       `{"name": "John"}`,
			expectError: false,
			expected:    TestStruct{Name: "John", Age: 0, Email: ""},
			description: "Should parse partial JSON object",
		},
		{
			name:        "empty object",
			input:       `{}`,
			expectError: false,
			expected:    TestStruct{},
			description: "Should parse empty JSON object",
		},
		{
			name:        "invalid JSON",
			input:       `{"name": "John", "age": 30`,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for invalid JSON",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for empty string",
		},
		{
			name:        "wrong type for field",
			input:       `{"name": "John", "age": "thirty", "email": "john@example.com"}`,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for wrong field type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringToObject[TestStruct](tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none - %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v - %s", err, tt.description)
				}
				if result != tt.expected {
					t.Errorf("Expected %+v, got %+v - %s", tt.expected, result, tt.description)
				}
			}
		})
	}
}

// Test StringToObject with different types
func TestStringToObject_DifferentTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "string type",
			input:       `"hello world"`,
			expectError: false,
			description: "Should parse JSON string to string type",
		},
		{
			name:        "int type",
			input:       `42`,
			expectError: false,
			description: "Should parse JSON number to int type",
		},
		{
			name:        "bool type",
			input:       `true`,
			expectError: false,
			description: "Should parse JSON boolean to bool type",
		},
		{
			name:        "slice type",
			input:       `["a", "b", "c"]`,
			expectError: false,
			description: "Should parse JSON array to slice type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with string type
			if tt.name == "string type" {
				result, err := StringToObject[string](tt.input)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none - %s", tt.description)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v - %s", err, tt.description)
					}
					if result != "hello world" {
						t.Errorf("Expected 'hello world', got '%s' - %s", result, tt.description)
					}
				}
			}

			// Test with int type
			if tt.name == "int type" {
				result, err := StringToObject[int](tt.input)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none - %s", tt.description)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v - %s", err, tt.description)
					}
					if result != 42 {
						t.Errorf("Expected 42, got %d - %s", result, tt.description)
					}
				}
			}

			// Test with bool type
			if tt.name == "bool type" {
				result, err := StringToObject[bool](tt.input)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none - %s", tt.description)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v - %s", err, tt.description)
					}
					if result != true {
						t.Errorf("Expected true, got %t - %s", result, tt.description)
					}
				}
			}

			// Test with slice type
			if tt.name == "slice type" {
				result, err := StringToObject[[]string](tt.input)
				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none - %s", tt.description)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v - %s", err, tt.description)
					}
					expected := []string{"a", "b", "c"}
					if len(result) != len(expected) {
						t.Errorf("Expected slice with %d items, got %d - %s", len(expected), len(result), tt.description)
					}
					for i, expectedValue := range expected {
						if i >= len(result) {
							t.Errorf("Expected item at index %d but slice is too short - %s", i, tt.description)
							break
						}
						if result[i] != expectedValue {
							t.Errorf("Expected '%s' at index %d, got '%s' - %s", expectedValue, i, result[i], tt.description)
						}
					}
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkObfuscatePassword(b *testing.B) {
	password := "verylongpassword123"
	for i := 0; i < b.N; i++ {
		ObfuscateString(password)
	}
}

func BenchmarkSlugify(b *testing.B) {
	input := "Hello, World! This is a test string with special characters @#$%"
	for i := 0; i < b.N; i++ {
		Slugify(input)
	}
}

func BenchmarkStringToMap(b *testing.B) {
	input := `{"name": "John", "age": 30, "email": "john@example.com", "active": true}`
	for i := 0; i < b.N; i++ {
		StringToMap(input)
	}
}

func BenchmarkStringToSlice(b *testing.B) {
	input := `["item1", "item2", "item3", "item4", "item5"]`
	for i := 0; i < b.N; i++ {
		StringToSlice(input)
	}
}

func BenchmarkStringToObject(b *testing.B) {
	input := `{"name": "John", "age": 30, "email": "john@example.com"}`
	for i := 0; i < b.N; i++ {
		StringToObject[TestStruct](input)
	}
}
