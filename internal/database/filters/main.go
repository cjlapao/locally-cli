// Package filters provides a service for managing filters
package filters

import (
	"fmt"
	"strings"
)

// This package will be user to create filters for the database

type FilterOperator string

const (
	FilterOperatorEqual              FilterOperator = "="
	FilterOperatorNotEqual           FilterOperator = "!="
	FilterOperatorGreaterThan        FilterOperator = ">"
	FilterOperatorGreaterThanOrEqual FilterOperator = ">="
	FilterOperatorLessThan           FilterOperator = "<"
	FilterOperatorLessThanOrEqual    FilterOperator = "<="
	FilterOperatorLike               FilterOperator = "LIKE"
	FilterOperatorIn                 FilterOperator = "IN"
	FilterOperatorNotIn              FilterOperator = "NOT IN"
	FilterOperatorIsNull             FilterOperator = "IS NULL"
	FilterOperatorIsNotNull          FilterOperator = "IS NOT NULL"
	FilterOperatorBetween            FilterOperator = "BETWEEN"
	FilterOperatorNotBetween         FilterOperator = "NOT BETWEEN"
	FilterOperatorContains           FilterOperator = "CONTAINS"
)

type FilterJoiner string

const (
	FilterJoinerNone FilterJoiner = ""
	FilterJoinerAnd  FilterJoiner = "AND"
	FilterJoinerOr   FilterJoiner = "OR"
)

type FilterItem struct {
	Joiner   FilterJoiner
	Field    string
	Operator FilterOperator
	Value    string
}

type Filter struct {
	items    []FilterItem
	Page     int
	PageSize int
}

func NewFilter() *Filter {
	return &Filter{
		items:    make([]FilterItem, 0),
		Page:     -1,
		PageSize: -1,
	}
}

func (f *Filter) WithPage(page int) *Filter {
	f.Page = page
	return f
}

func (f *Filter) WithPageSize(pageSize int) *Filter {
	f.PageSize = pageSize
	return f
}

func (f *Filter) WithField(field string, operator FilterOperator, value string, joiner FilterJoiner) *Filter {
	if f.items == nil {
		f.items = make([]FilterItem, 0)
	}
	// Only set joiner to AND if not provided and not the first item
	if len(f.items) == 0 {
		joiner = FilterJoinerNone
	}
	f.items = append(f.items, FilterItem{
		Field:    field,
		Operator: operator,
		Value:    value,
		Joiner:   joiner,
	})
	return f
}

func (f *Filter) WithFields(fields ...FilterItem) *Filter {
	if f.items == nil {
		f.items = make([]FilterItem, 0)
	}
	f.items = append(f.items, fields...)
	return f
}

func (f *Filter) Generate() (string, []interface{}) {
	if f == nil {
		return "", nil
	}
	filter := ""
	if len(f.items) == 0 {
		return "", nil
	}
	args := make([]interface{}, 0)
	for i, item := range f.items {
		if i > 0 {
			if item.Joiner != FilterJoinerNone {
				filter += " " + string(item.Joiner) + " "
			} else {
				filter += " "
			}
		}
		filter += fmt.Sprintf("%s %s ?", item.Field, item.Operator)
		args = append(args, item.Value)
	}
	return filter, args
}

// Parse method to parse a filter string into a Filter object
func Parse(val string) (*Filter, error) {
	if val == "" {
		return &Filter{}, nil
	}

	filter := &Filter{}
	parts := strings.Fields(val)
	if len(parts) == 0 {
		return filter, nil
	}

	var (
		currentItem *FilterItem
		nextJoiner  = FilterJoinerNone
		i           int
	)

	for i < len(parts) {
		part := parts[i]
		if currentItem == nil {
			currentItem = &FilterItem{
				Field:  part,
				Joiner: nextJoiner,
			}
			nextJoiner = FilterJoinerNone
			i++
			continue
		}

		// Operator
		operator, consumed := checkMultiWordOperator(parts, i)
		if operator != "" {
			currentItem.Operator = FilterOperator(operator)
			i += consumed
		} else if isSingleWordOperator(part) {
			currentItem.Operator = FilterOperator(part)
			i++
		} else {
			return nil, fmt.Errorf("invalid operator: %s", part)
		}

		// Value (if needed)
		if currentItem.Operator == FilterOperatorIsNull || currentItem.Operator == FilterOperatorIsNotNull {
			currentItem.Value = ""
		} else if i < len(parts) && !isJoiner(parts[i]) {
			currentItem.Value = parts[i]
			i++
		} else {
			currentItem.Value = ""
		}

		// Joiner or end
		if i < len(parts) && isJoiner(parts[i]) {
			nextJoiner = FilterJoiner(parts[i])
			i++
		} else {
			nextJoiner = FilterJoinerNone
		}

		filter.WithField(currentItem.Field, currentItem.Operator, currentItem.Value, currentItem.Joiner)
		currentItem = nil
	}

	return filter, nil
}

// checkMultiWordOperator checks if the current position starts a multi-word operator
func checkMultiWordOperator(parts []string, start int) (string, int) {
	if start >= len(parts) {
		return "", 0
	}

	// Check for three-word operators first
	if start+2 < len(parts) {
		threeWord := parts[start] + " " + parts[start+1] + " " + parts[start+2]
		switch threeWord {
		case "IS NOT NULL":
			return "IS NOT NULL", 3
		}
	}

	// Check for two-word operators
	if start+1 < len(parts) {
		twoWord := parts[start] + " " + parts[start+1]
		switch twoWord {
		case "NOT IN":
			return "NOT IN", 2
		case "IS NULL":
			return "IS NULL", 2
		case "NOT BETWEEN":
			return "NOT BETWEEN", 2
		}
	}

	return "", 0
}

// isSingleWordOperator checks if a single word is a valid operator
func isSingleWordOperator(part string) bool {
	operators := []string{"=", "!=", ">", ">=", "<", "<=", "LIKE", "IN", "BETWEEN", "CONTAINS"}
	for _, op := range operators {
		if part == op {
			return true
		}
	}
	return false
}

// isJoiner checks if a word is a valid joiner
func isJoiner(part string) bool {
	return part == "AND" || part == "OR"
}
