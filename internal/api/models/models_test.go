package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationRequest_ToQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		request  *PaginationRequest
		expected struct {
			hasFilters    bool
			hasOrdering   bool
			hasPagination bool
			page          int
			pageSize      int
		}
	}{
		{
			name: "nil request",
			request: nil,
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{false, true, true, 1, 20}, // defaults
		},
		{
			name: "empty request",
			request: &PaginationRequest{},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{false, true, true, 1, 20}, // defaults
		},
		{
			name: "pagination only",
			request: &PaginationRequest{
				Page:     2,
				PageSize: 50,
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{false, true, true, 2, 50},
		},
		{
			name: "with filter",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 10,
				Filter:   "name=john",
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{true, true, true, 1, 10},
		},
		{
			name: "with sort",
			request: &PaginationRequest{
				Page:     3,
				PageSize: 25,
				Sort:     "created_at desc",
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{false, true, true, 3, 25},
		},
		{
			name: "with order fallback",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 15,
				Order:    "name asc",
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{false, true, true, 1, 15},
		},
		{
			name: "complete request",
			request: &PaginationRequest{
				Page:     4,
				PageSize: 30,
				Filter:   "status=active,age>18",
				Sort:     "name desc,created_at asc",
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{true, true, true, 4, 30},
		},
		{
			name: "complex filter",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 20,
				Filter:   "status=active AND (category=tech OR category=science)",
				Sort:     "priority desc",
			},
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
				page          int
				pageSize      int
			}{true, true, true, 1, 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.request.ToQueryBuilder()

			assert.NotNil(t, qb)
			assert.Equal(t, tt.expected.hasFilters, qb.HasFilters())
			assert.Equal(t, tt.expected.hasOrdering, qb.HasOrdering())
			assert.Equal(t, tt.expected.hasPagination, qb.HasPagination())
			assert.Equal(t, tt.expected.page, qb.GetPage())
			assert.Equal(t, tt.expected.pageSize, qb.GetPageSize())
		})
	}
}

func TestPaginationRequest_ToQueryString(t *testing.T) {
	tests := []struct {
		name     string
		request  *PaginationRequest
		expected string
	}{
		{
			name:     "nil request",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty request",
			request:  &PaginationRequest{},
			expected: "",
		},
		{
			name: "pagination only",
			request: &PaginationRequest{
				Page:     2,
				PageSize: 50,
			},
			expected: "?page=2&page_size=50",
		},
		{
			name: "with filter",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 10,
				Filter:   "name=john",
			},
			expected: "?page=1&page_size=10&filter=name=john",
		},
		{
			name: "with sort",
			request: &PaginationRequest{
				Page:     3,
				PageSize: 25,
				Sort:     "created_at desc",
			},
			expected: "?page=3&page_size=25&sort=created_at desc",
		},
		{
			name: "with order",
			request: &PaginationRequest{
				Page:  1,
				Order: "name asc",
			},
			expected: "?page=1&order=name asc",
		},
		{
			name: "complete request",
			request: &PaginationRequest{
				Page:     4,
				PageSize: 30,
				Filter:   "status=active",
				Sort:     "name desc",
				Order:    "priority asc",
			},
			expected: "?page=4&page_size=30&filter=status=active&sort=name desc&order=priority asc",
		},
		{
			name: "zero values ignored",
			request: &PaginationRequest{
				Page:     0, // should be ignored
				PageSize: 0, // should be ignored
				Filter:   "status=active",
			},
			expected: "?filter=status=active",
		},
		{
			name: "complex filter",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 20,
				Filter:   "status=active AND category IN (tech,science)",
			},
			expected: "?page=1&page_size=20&filter=status=active AND category IN (tech,science)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToQueryString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name      string
		parts     []string
		separator string
		expected  string
	}{
		{
			name:      "empty parts",
			parts:     []string{},
			separator: "&",
			expected:  "",
		},
		{
			name:      "single part",
			parts:     []string{"page=1"},
			separator: "&",
			expected:  "page=1",
		},
		{
			name:      "multiple parts",
			parts:     []string{"page=1", "page_size=20", "filter=name=john"},
			separator: "&",
			expected:  "page=1&page_size=20&filter=name=john",
		},
		{
			name:      "different separator",
			parts:     []string{"a", "b", "c"},
			separator: ",",
			expected:  "a,b,c",
		},
		{
			name:      "empty separator",
			parts:     []string{"hello", "world"},
			separator: "",
			expected:  "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.parts, tt.separator)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationRequest_Integration(t *testing.T) {
	// Test the full round trip: PaginationRequest -> QueryBuilder -> Database query simulation
	request := &PaginationRequest{
		Page:     2,
		PageSize: 15,
		Filter:   "status=active,age>25",
		Sort:     "created_at desc,name asc",
	}

	// Convert to QueryBuilder
	qb := request.ToQueryBuilder()

	// Verify QueryBuilder properties
	assert.True(t, qb.HasFilters())
	assert.True(t, qb.HasOrdering())
	assert.True(t, qb.HasPagination())

	// Verify pagination values
	assert.Equal(t, 2, qb.GetPage())
	assert.Equal(t, 15, qb.GetPageSize())
	assert.Equal(t, 15, qb.GetOffset()) // (2-1) * 15 = 15

	// Test round trip with ToQueryString
	queryString := request.ToQueryString()
	assert.Contains(t, queryString, "page=2")
	assert.Contains(t, queryString, "page_size=15")
	assert.Contains(t, queryString, "filter=status=active,age>25")
	assert.Contains(t, queryString, "sort=created_at desc,name asc")
}

func TestPaginationRequest_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		request     *PaginationRequest
		description string
	}{
		{
			name: "negative values",
			request: &PaginationRequest{
				Page:     -1,
				PageSize: -10,
				Filter:   "valid=filter",
			},
			description: "Should handle negative values gracefully",
		},
		{
			name: "very large values",
			request: &PaginationRequest{
				Page:     999999,
				PageSize: 999999,
			},
			description: "Should handle very large values",
		},
		{
			name: "empty strings",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 10,
				Filter:   "",
				Sort:     "",
				Order:    "",
			},
			description: "Should handle empty strings",
		},
		{
			name: "special characters in filter",
			request: &PaginationRequest{
				Page:     1,
				PageSize: 10,
				Filter:   "email=user@example.com,name='John O\\'Brien'",
			},
			description: "Should handle special characters in filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			qb := tt.request.ToQueryBuilder()
			assert.NotNil(t, qb)

			queryString := tt.request.ToQueryString()
			// Should return a string (might be empty)
			assert.IsType(t, "", queryString)
		})
	}
}