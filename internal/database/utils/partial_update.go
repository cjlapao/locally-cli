// Package utils provides utility functions for the database layer of the application
package utils

import (
	"reflect"
	"strings"
	"time"
)

// isJSONField checks if a field is stored as JSON in the database
func isJSONField(field reflect.StructField) bool {
	gormTag := field.Tag.Get("gorm")
	return strings.Contains(gormTag, "type:json")
}

// isEmptyJSONValue checks if a value represents an empty JSON array or object
func isEmptyJSONValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		// Check for empty JSON arrays or objects
		trimmed := strings.TrimSpace(v)
		return trimmed == "[]" || trimmed == "{}" || trimmed == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}

	return false
}

// shouldSkipField determines if a field should be skipped based on its value and type
func shouldSkipField(field reflect.Value, fieldType reflect.StructField) bool {
	// Skip unexported fields
	if !field.CanInterface() {
		return true
	}

	// Skip fields with json tag "-"
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag == "-" {
		return true
	}

	// Check if it's a zero value
	if field.IsZero() {
		return true
	}

	// For JSON fields, check if the value represents empty JSON
	if isJSONField(fieldType) {
		return isEmptyJSONValue(field.Interface())
	}

	return false
}

// PartialUpdateMap creates a map of fields to update based on non-zero values
// This is a generic function that can be used for any struct
func PartialUpdateMap(entity interface{}, alwaysUpdateFields ...string) map[string]interface{} {
	updates := make(map[string]interface{})

	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip fields that should not be updated
		if shouldSkipField(field, fieldType) {
			continue
		}

		// Get the JSON tag for the field name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		// Remove omitempty from the tag
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			jsonTag = jsonTag[:commaIndex]
		}

		// Add the field to updates
		updates[jsonTag] = field.Interface()
	}

	// Always update specified fields
	for _, fieldName := range alwaysUpdateFields {
		switch fieldName {
		case "updated_at":
			updates["updated_at"] = time.Now()
		}
	}

	return updates
}

// PartialUpdateMapWithCustomFields creates a partial update map with custom field mapping
func PartialUpdateMapWithCustomFields(entity interface{}, fieldMappings map[string]string, alwaysUpdateFields ...string) map[string]interface{} {
	updates := make(map[string]interface{})

	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip fields that should not be updated
		if shouldSkipField(field, fieldType) {
			continue
		}

		// Get the JSON tag for the field name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty from the tag
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			jsonTag = jsonTag[:commaIndex]
		}

		// Use custom mapping if available, otherwise use the JSON tag
		dbField := jsonTag
		if mappedField, exists := fieldMappings[jsonTag]; exists {
			dbField = mappedField
		}
		updates[dbField] = field.Interface()
	}

	// Always update specified fields
	for _, fieldName := range alwaysUpdateFields {
		switch fieldName {
		case "updated_at":
			updates["updated_at"] = time.Now()
		}
	}

	return updates
}
