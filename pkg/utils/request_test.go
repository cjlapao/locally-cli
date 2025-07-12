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
