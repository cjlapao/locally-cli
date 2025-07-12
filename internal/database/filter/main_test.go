package filter

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    *Filter
	}{
		{
			name:        "empty string",
			input:       "",
			expectError: false,
			expected:    &Filter{items: []FilterItem{}},
		},
		{
			name:        "single field with equals",
			input:       "name = value",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "name", Operator: FilterOperatorEqual, Value: "value", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with not equals",
			input:       "status != active",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorNotEqual, Value: "active", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with greater than",
			input:       "age > 18",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "age", Operator: FilterOperatorGreaterThan, Value: "18", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with greater than or equal",
			input:       "score >= 100",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "score", Operator: FilterOperatorGreaterThanOrEqual, Value: "100", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with less than",
			input:       "price < 50",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "price", Operator: FilterOperatorLessThan, Value: "50", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with less than or equal",
			input:       "quantity <= 10",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "quantity", Operator: FilterOperatorLessThanOrEqual, Value: "10", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with LIKE",
			input:       "title LIKE %test%",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "title", Operator: FilterOperatorLike, Value: "%test%", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with IN",
			input:       "category IN electronics",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "category", Operator: FilterOperatorIn, Value: "electronics", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with NOT IN",
			input:       "status NOT IN deleted",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorNotIn, Value: "deleted", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with IS NULL",
			input:       "deleted_at IS NULL",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "deleted_at", Operator: FilterOperatorIsNull, Value: "", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with IS NOT NULL",
			input:       "email IS NOT NULL",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "email", Operator: FilterOperatorIsNotNull, Value: "", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with BETWEEN",
			input:       "date BETWEEN 2023-01-01 2023-12-31",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "date", Operator: FilterOperatorBetween, Value: "2023-01-01", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with NOT BETWEEN",
			input:       "price NOT BETWEEN 10 100",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "price", Operator: FilterOperatorNotBetween, Value: "10", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "single field with CONTAINS",
			input:       "description CONTAINS important",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "description", Operator: FilterOperatorContains, Value: "important", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "two fields with AND",
			input:       "status = active AND age > 18",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorEqual, Value: "active", Joiner: FilterJoinerNone},
					{Field: "age", Operator: FilterOperatorGreaterThan, Value: "18", Joiner: FilterJoinerAnd},
				},
			},
		},
		{
			name:        "two fields with OR",
			input:       "category = electronics OR category = books",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "category", Operator: FilterOperatorEqual, Value: "electronics", Joiner: FilterJoinerNone},
					{Field: "category", Operator: FilterOperatorEqual, Value: "books", Joiner: FilterJoinerOr},
				},
			},
		},
		{
			name:        "three fields with mixed joiners",
			input:       "status = active AND age > 18 OR vip = true",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorEqual, Value: "active", Joiner: FilterJoinerNone},
					{Field: "age", Operator: FilterOperatorGreaterThan, Value: "18", Joiner: FilterJoinerAnd},
					{Field: "vip", Operator: FilterOperatorEqual, Value: "true", Joiner: FilterJoinerOr},
				},
			},
		},
		{
			name:        "field with spaces in value",
			input:       "name = John Doe",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "name", Operator: FilterOperatorEqual, Value: "John", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "multiple spaces between parts",
			input:       "name  =  value",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "name", Operator: FilterOperatorEqual, Value: "value", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "invalid operator",
			input:       "name INVALID value",
			expectError: true,
			expected:    nil,
		},
		{
			name:        "invalid joiner",
			input:       "status = active INVALID age > 18",
			expectError: true,
			expected:    nil,
		},
		{
			name:        "missing value after operator",
			input:       "name =",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "name", Operator: FilterOperatorEqual, Value: "", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "missing joiner after value",
			input:       "status = active age > 18",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorEqual, Value: "active", Joiner: FilterJoinerNone},
					{Field: "age", Operator: FilterOperatorGreaterThan, Value: "18", Joiner: FilterJoinerNone},
				},
			},
		},
		{
			name:        "complex nested conditions",
			input:       "status = active AND (age > 18 OR vip = true) AND category = electronics",
			expectError: false,
			expected: &Filter{
				items: []FilterItem{
					{Field: "status", Operator: FilterOperatorEqual, Value: "active", Joiner: FilterJoinerNone},
					{Field: "(age", Operator: FilterOperatorGreaterThan, Value: "18", Joiner: FilterJoinerAnd},
					{Field: "vip", Operator: FilterOperatorEqual, Value: "true)", Joiner: FilterJoinerOr},
					{Field: "category", Operator: FilterOperatorEqual, Value: "electronics", Joiner: FilterJoinerAnd},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			// Compare the filter items
			if len(result.items) != len(tt.expected.items) {
				t.Errorf("Expected %d items, got %d", len(tt.expected.items), len(result.items))
				return
			}

			for i, expectedItem := range tt.expected.items {
				if i >= len(result.items) {
					t.Errorf("Missing item at index %d", i)
					continue
				}

				actualItem := result.items[i]
				if actualItem.Field != expectedItem.Field {
					t.Errorf("Item %d: Expected field '%s', got '%s'", i, expectedItem.Field, actualItem.Field)
				}
				if actualItem.Operator != expectedItem.Operator {
					t.Errorf("Item %d: Expected operator '%s', got '%s'", i, expectedItem.Operator, actualItem.Operator)
				}
				if actualItem.Value != expectedItem.Value {
					t.Errorf("Item %d: Expected value '%s', got '%s'", i, expectedItem.Value, actualItem.Value)
				}
				if actualItem.Joiner != expectedItem.Joiner {
					t.Errorf("Item %d: Expected joiner '%s', got '%s'", i, expectedItem.Joiner, actualItem.Joiner)
				}
			}
		})
	}
}

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "only spaces",
			input:       "   ",
			description: "Should handle strings with only spaces",
		},
		{
			name:        "single word",
			input:       "fieldname",
			description: "Should handle single word input",
		},
		{
			name:        "field and operator only",
			input:       "name =",
			description: "Should handle incomplete filter with missing value",
		},
		{
			name:        "field operator value joiner",
			input:       "status = active AND",
			description: "Should handle incomplete filter ending with joiner",
		},
		{
			name:        "very long field name",
			input:       "very_long_field_name_with_underscores = value",
			description: "Should handle long field names",
		},
		{
			name:        "numeric values",
			input:       "id = 123 AND count > 0",
			description: "Should handle numeric values as strings",
		},
		{
			name:        "special characters in values",
			input:       "email = user@example.com AND path = /api/v1/users",
			description: "Should handle special characters in values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			// These edge cases should not cause panics
			if err != nil {
				t.Logf("Got expected error for %s: %v", tt.description, err)
			}

			if result != nil {
				t.Logf("Successfully parsed: %s", tt.description)
			}
		})
	}
}

func TestParseGenerateIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		args     []interface{}
	}{
		{
			name:     "simple equals",
			input:    "name = value",
			expected: "name = ?",
			args:     []interface{}{"value"},
		},
		{
			name:     "two conditions with AND",
			input:    "status = active AND age > 18",
			expected: "status = ? AND age > ?",
			args:     []interface{}{"active", "18"},
		},
		{
			name:     "three conditions with mixed joiners",
			input:    "category = electronics AND price > 100 OR vip = true",
			expected: "category = ? AND price > ? OR vip = ?",
			args:     []interface{}{"electronics", "100", "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			generated, args := filter.Generate()

			if generated != tt.expected {
				t.Errorf("Expected SQL '%s', got '%s'", tt.expected, generated)
			}

			if len(args) != len(tt.args) {
				t.Errorf("Expected %d args, got %d", len(tt.args), len(args))
				return
			}

			for i, expectedArg := range tt.args {
				if i >= len(args) {
					t.Errorf("Missing arg at index %d", i)
					continue
				}
				if args[i] != expectedArg {
					t.Errorf("Arg %d: Expected %v, got %v", i, expectedArg, args[i])
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkParse(b *testing.B) {
	testCases := []string{
		"",
		"name = value",
		"status = active AND age > 18",
		"category = electronics AND price > 100 OR vip = true",
		"status = active AND (age > 18 OR vip = true) AND category = electronics",
	}

	for _, input := range testCases {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Parse(input)
			}
		})
	}
}

func TestWithField(t *testing.T) {
	f := &Filter{}
	f.WithField("name", FilterOperatorEqual, "john", FilterJoinerNone)
	if len(f.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(f.items))
	}
	item := f.items[0]
	if item.Field != "name" || item.Operator != FilterOperatorEqual || item.Value != "john" || item.Joiner != FilterJoinerNone {
		t.Errorf("unexpected item: %+v", item)
	}

	// Add another with AND joiner
	f.WithField("age", FilterOperatorGreaterThan, "18", FilterJoinerAnd)
	if len(f.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(f.items))
	}
	item2 := f.items[1]
	if item2.Field != "age" || item2.Operator != FilterOperatorGreaterThan || item2.Value != "18" || item2.Joiner != FilterJoinerAnd {
		t.Errorf("unexpected item2: %+v", item2)
	}
}

func TestWithFields(t *testing.T) {
	f := &Filter{}
	items := []FilterItem{
		{Field: "a", Operator: FilterOperatorEqual, Value: "1", Joiner: FilterJoinerNone},
		{Field: "b", Operator: FilterOperatorNotEqual, Value: "2", Joiner: FilterJoinerAnd},
	}
	f.WithFields(items...)
	if len(f.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(f.items))
	}
	if f.items[0] != items[0] || f.items[1] != items[1] {
		t.Errorf("items not added correctly: %+v", f.items)
	}
}

func TestGenerate_Empty(t *testing.T) {
	f := &Filter{}
	s, args := f.Generate()
	if s != "" || args != nil {
		t.Errorf("expected empty string and nil args, got '%s', %v", s, args)
	}
}

func TestGenerate_SingleItem(t *testing.T) {
	f := &Filter{}
	f.WithField("foo", FilterOperatorEqual, "bar", FilterJoinerNone)
	s, args := f.Generate()
	if s != "foo = ?" {
		t.Errorf("expected 'foo = ?', got '%s'", s)
	}
	if len(args) != 1 || args[0] != "bar" {
		t.Errorf("expected args [bar], got %v", args)
	}
}

func TestGenerate_MultipleItems(t *testing.T) {
	f := &Filter{}
	f.WithField("foo", FilterOperatorEqual, "bar", FilterJoinerNone)
	f.WithField("baz", FilterOperatorGreaterThan, "10", FilterJoinerAnd)
	s, args := f.Generate()
	if s != "foo = ? AND baz > ?" {
		t.Errorf("expected 'foo = ? AND baz > ?', got '%s'", s)
	}
	if len(args) != 2 || args[0] != "bar" || args[1] != "10" {
		t.Errorf("expected args [bar 10], got %v", args)
	}
}

func TestGenerate_JoinerNoneBetweenItems(t *testing.T) {
	f := &Filter{}
	f.WithField("foo", FilterOperatorEqual, "bar", FilterJoinerNone)
	f.WithField("baz", FilterOperatorGreaterThan, "10", FilterJoinerNone)
	s, args := f.Generate()
	if s != "foo = ? baz > ?" {
		t.Errorf("expected 'foo = ? baz > ?', got '%s'", s)
	}
	if len(args) != 2 || args[0] != "bar" || args[1] != "10" {
		t.Errorf("expected args [bar 10], got %v", args)
	}
}
