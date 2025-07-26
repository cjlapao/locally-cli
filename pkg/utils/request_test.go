package utils

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetPaginationFromRequest(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  map[string]string
		expectedPage int
		expectedSize int
		description  string
	}{
		{
			name:         "default values",
			queryParams:  map[string]string{},
			expectedPage: 1,
			expectedSize: 20,
			description:  "Should return default values when no pagination params provided",
		},
		{
			name:         "custom page and size",
			queryParams:  map[string]string{"page": "5", "page_size": "50"},
			expectedPage: 5,
			expectedSize: 50,
			description:  "Should return custom page and page_size values",
		},
		{
			name:         "only page provided",
			queryParams:  map[string]string{"page": "3"},
			expectedPage: 3,
			expectedSize: 20,
			description:  "Should return custom page with default page_size",
		},
		{
			name:         "only page_size provided",
			queryParams:  map[string]string{"page_size": "100"},
			expectedPage: 1,
			expectedSize: 100,
			description:  "Should return default page with custom page_size",
		},
		{
			name:         "invalid page number",
			queryParams:  map[string]string{"page": "invalid", "page_size": "20"},
			expectedPage: 0,
			expectedSize: 0,
			description:  "Should return 0,0 when page is invalid",
		},
		{
			name:         "invalid page_size",
			queryParams:  map[string]string{"page": "1", "page_size": "invalid"},
			expectedPage: 0,
			expectedSize: 0,
			description:  "Should return 0,0 when page_size is invalid",
		},
		{
			name:         "both invalid",
			queryParams:  map[string]string{"page": "invalid", "page_size": "invalid"},
			expectedPage: 0,
			expectedSize: 0,
			description:  "Should return 0,0 when both values are invalid",
		},
		{
			name:         "zero values",
			queryParams:  map[string]string{"page": "0", "page_size": "0"},
			expectedPage: 0,
			expectedSize: 0,
			description:  "Should handle zero values correctly",
		},
		{
			name:         "negative values",
			queryParams:  map[string]string{"page": "-1", "page_size": "-10"},
			expectedPage: -1,
			expectedSize: -10,
			description:  "Should handle negative values correctly",
		},
		{
			name:         "large values",
			queryParams:  map[string]string{"page": "999999", "page_size": "999999"},
			expectedPage: 999999,
			expectedSize: 999999,
			description:  "Should handle large values correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := &http.Request{}
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Set(key, value)
				}
				req.URL = &url.URL{RawQuery: values.Encode()}
			} else {
				req.URL = &url.URL{}
			}

			page, size := GetPaginationFromRequest(req)

			if page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d - %s", tt.expectedPage, page, tt.description)
			}
			if size != tt.expectedSize {
				t.Errorf("Expected size %d, got %d - %s", tt.expectedSize, size, tt.description)
			}
		})
	}
}

func TestHasPaginationRequest(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		expected    bool
		description string
	}{
		{
			name:        "no pagination",
			queryParams: map[string]string{},
			expected:    false,
			description: "Should return false when no page parameter is present",
		},
		{
			name:        "with page parameter",
			queryParams: map[string]string{"page": "1"},
			expected:    true,
			description: "Should return true when page parameter is present",
		},
		{
			name:        "with page_size but no page",
			queryParams: map[string]string{"page_size": "20"},
			expected:    false,
			description: "Should return false when only page_size is present",
		},
		{
			name:        "with both page and page_size",
			queryParams: map[string]string{"page": "5", "page_size": "50"},
			expected:    true,
			description: "Should return true when both parameters are present",
		},
		{
			name:        "empty page parameter",
			queryParams: map[string]string{"page": ""},
			expected:    false,
			description: "Should return false when page parameter is empty",
		},
		{
			name:        "other parameters present",
			queryParams: map[string]string{"filter": "name = test", "sort": "name"},
			expected:    false,
			description: "Should return false when other parameters are present but no page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := &http.Request{}
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Set(key, value)
				}
				req.URL = &url.URL{RawQuery: values.Encode()}
			} else {
				req.URL = &url.URL{}
			}

			result := HasPaginationRequest(req)

			if result != tt.expected {
				t.Errorf("Expected %t, got %t - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestGetFilterFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		expectError bool
		description string
	}{
		{
			name:        "no filter parameter",
			queryParams: map[string]string{},
			expectError: false,
			description: "Should return nil filter when no filter parameter is present",
		},
		{
			name:        "empty filter parameter",
			queryParams: map[string]string{"filter": ""},
			expectError: false,
			description: "Should return nil filter when filter parameter is empty",
		},
		{
			name:        "valid simple filter",
			queryParams: map[string]string{"filter": "name = test"},
			expectError: false,
			description: "Should return valid filter for simple condition",
		},
		{
			name:        "valid complex filter",
			queryParams: map[string]string{"filter": "status = active AND age > 18"},
			expectError: false,
			description: "Should return valid filter for complex condition",
		},
		{
			name:        "invalid filter syntax",
			queryParams: map[string]string{"filter": "name INVALID test"},
			expectError: true,
			description: "Should return error for invalid filter syntax",
		},
		{
			name:        "filter with special characters",
			queryParams: map[string]string{"filter": "email = user@example.com"},
			expectError: false,
			description: "Should handle filter with special characters",
		},
		{
			name:        "filter with spaces",
			queryParams: map[string]string{"filter": "name = John Doe"},
			expectError: false,
			description: "Should handle filter with spaces in values",
		},
		{
			name:        "other parameters present",
			queryParams: map[string]string{"filter": "name = test", "page": "1", "sort": "name"},
			expectError: false,
			description: "Should extract filter when other parameters are present",
		},
		{
			name:        "filter with quotes",
			queryParams: map[string]string{"filter": `"name = test"`},
			expectError: false,
			description: "Should handle filter wrapped in quotes",
		},
		{
			name:        "filter with single quote",
			queryParams: map[string]string{"filter": `"name = test`},
			expectError: false,
			description: "Should handle filter with only opening quote",
		},
		{
			name:        "filter with trailing quote",
			queryParams: map[string]string{"filter": `name = test"`},
			expectError: false,
			description: "Should handle filter with only closing quote",
		},
		{
			name:        "filter with multiple quotes",
			queryParams: map[string]string{"filter": `""name = test""`},
			expectError: false,
			description: "Should handle filter with multiple quotes",
		},
		{
			name:        "filter with quotes and spaces",
			queryParams: map[string]string{"filter": `" name = test "`},
			expectError: false,
			description: "Should handle filter with quotes and spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := &http.Request{}
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Set(key, value)
				}
				req.URL = &url.URL{RawQuery: values.Encode()}
			} else {
				req.URL = &url.URL{}
			}

			result, err := GetFilterFromRequest(req)

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
				// For valid filters, we can't easily test the exact content without exposing internal structure
				// But we can verify it's not nil when we expect a filter
				if tt.queryParams["filter"] != "" && result == nil {
					t.Errorf("Expected filter result but got nil - %s", tt.description)
				}
			}
		})
	}
}

func TestGetFilterFromRequest_Integration(t *testing.T) {
	tests := []struct {
		name           string
		filterString   string
		expectedFields []string
		description    string
	}{
		{
			name:           "simple equals filter",
			filterString:   "name = test",
			expectedFields: []string{"name"},
			description:    "Should parse simple equals filter",
		},
		{
			name:           "multiple conditions",
			filterString:   "status = active AND age > 18",
			expectedFields: []string{"status", "age"},
			description:    "Should parse multiple conditions",
		},
		{
			name:           "OR condition",
			filterString:   "category = electronics OR category = books",
			expectedFields: []string{"category", "category"},
			description:    "Should parse OR conditions",
		},
		{
			name:           "IS NULL condition",
			filterString:   "deleted_at IS NULL",
			expectedFields: []string{"deleted_at"},
			description:    "Should parse IS NULL condition",
		},
		{
			name:           "LIKE condition",
			filterString:   "title LIKE %test%",
			expectedFields: []string{"title"},
			description:    "Should parse LIKE condition",
		},
		{
			name:           "filter with quotes",
			filterString:   `"name = test"`,
			expectedFields: []string{"name"},
			description:    "Should parse filter wrapped in quotes",
		},
		{
			name:           "filter with single quote",
			filterString:   `"name = test`,
			expectedFields: []string{"name"},
			description:    "Should parse filter with only opening quote",
		},
		{
			name:           "filter with trailing quote",
			filterString:   `name = test"`,
			expectedFields: []string{"name"},
			description:    "Should parse filter with only closing quote",
		},
		{
			name:           "filter with multiple quotes",
			filterString:   `""name = test""`,
			expectedFields: []string{"name"},
			description:    "Should parse filter with multiple quotes",
		},
		{
			name:           "complex filter with quotes",
			filterString:   `"status = active AND age > 18"`,
			expectedFields: []string{"status", "age"},
			description:    "Should parse complex filter wrapped in quotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with filter parameter
			values := url.Values{}
			values.Set("filter", tt.filterString)
			req := &http.Request{
				URL: &url.URL{RawQuery: values.Encode()},
			}

			result, err := GetFilterFromRequest(req)
			if err != nil {
				t.Fatalf("Unexpected error: %v - %s", err, tt.description)
			}

			if result == nil {
				t.Fatalf("Expected filter result but got nil - %s", tt.description)
			}

			// Generate SQL to verify the filter was parsed correctly
			sql, args := result.Generate()
			if sql == "" {
				t.Errorf("Expected SQL but got empty string - %s", tt.description)
			}

			// Verify we have the expected number of arguments
			if len(args) != len(tt.expectedFields) {
				t.Errorf("Expected %d arguments, got %d - %s", len(tt.expectedFields), len(args), tt.description)
			}
		})
	}
}

func TestHasFilterRequest(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		expected    bool
		description string
	}{
		{
			name:        "no filter parameter",
			queryParams: map[string]string{},
			expected:    false,
			description: "Should return false when no filter parameter is present",
		},
		{
			name:        "with filter parameter",
			queryParams: map[string]string{"filter": "name = test"},
			expected:    true,
			description: "Should return true when filter parameter is present",
		},
		{
			name:        "empty filter parameter",
			queryParams: map[string]string{"filter": ""},
			expected:    false,
			description: "Should return false when filter parameter is empty",
		},
		{
			name:        "other parameters present",
			queryParams: map[string]string{"page": "1", "sort": "name"},
			expected:    false,
			description: "Should return false when other parameters are present but no filter",
		},
		{
			name:        "filter with other parameters",
			queryParams: map[string]string{"filter": "name = test", "page": "1"},
			expected:    true,
			description: "Should return true when filter is present with other parameters",
		},
		{
			name:        "filter with quotes",
			queryParams: map[string]string{"filter": `"name = test"`},
			expected:    true,
			description: "Should return true when filter with quotes is present",
		},
		{
			name:        "filter with single quote",
			queryParams: map[string]string{"filter": `"name = test`},
			expected:    true,
			description: "Should return true when filter with single quote is present",
		},
		{
			name:        "filter with trailing quote",
			queryParams: map[string]string{"filter": `name = test"`},
			expected:    true,
			description: "Should return true when filter with trailing quote is present",
		},
		{
			name:        "filter with multiple quotes",
			queryParams: map[string]string{"filter": `""name = test""`},
			expected:    true,
			description: "Should return true when filter with multiple quotes is present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := &http.Request{}
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Set(key, value)
				}
				req.URL = &url.URL{RawQuery: values.Encode()}
			} else {
				req.URL = &url.URL{}
			}

			result := HasFilterRequest(req)

			if result != tt.expected {
				t.Errorf("Expected %t, got %t - %s", tt.expected, result, tt.description)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkGetPaginationFromRequest(b *testing.B) {
	values := url.Values{}
	values.Set("page", "5")
	values.Set("page_size", "50")
	req := &http.Request{
		URL: &url.URL{RawQuery: values.Encode()},
	}

	for i := 0; i < b.N; i++ {
		GetPaginationFromRequest(req)
	}
}

func BenchmarkGetFilterFromRequest(b *testing.B) {
	values := url.Values{}
	values.Set("filter", "status = active AND age > 18 OR category = electronics")
	req := &http.Request{
		URL: &url.URL{RawQuery: values.Encode()},
	}

	for i := 0; i < b.N; i++ {
		GetFilterFromRequest(req)
	}
}
