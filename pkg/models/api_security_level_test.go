package models

import (
	"testing"
)

func TestApiKeySecurityLevel_RequiresAuthentication(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected bool
	}{
		{"Any requires authentication", ApiKeySecurityLevelAny, true},
		{"SuperUser requires authentication", ApiKeySecurityLevelSuperUser, true},
		{"Bearer requires authentication", ApiKeySecurityLevelBearer, true},
		{"ApiKey requires authentication", ApiKeySecurityLevelApiKey, true},
		{"None does not require authentication", ApiKeySecurityLevelNone, false},
		{"Invalid level does not require authentication", ApiKeySecurityLevel("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.RequiresAuthentication()
			if result != tt.expected {
				t.Errorf("RequiresAuthentication() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_CanAuthenticateWith(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		authType ApiKeySecurityLevel
		expected bool
	}{
		// Any level tests
		{"Any can authenticate with SuperUser", ApiKeySecurityLevelAny, ApiKeySecurityLevelSuperUser, true},
		{"Any can authenticate with Bearer", ApiKeySecurityLevelAny, ApiKeySecurityLevelBearer, true},
		{"Any can authenticate with ApiKey", ApiKeySecurityLevelAny, ApiKeySecurityLevelApiKey, true},
		{"Any cannot authenticate with None", ApiKeySecurityLevelAny, ApiKeySecurityLevelNone, false},

		// SuperUser level tests
		{"SuperUser can authenticate with SuperUser", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelSuperUser, true},
		{"SuperUser cannot authenticate with Bearer", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, false},
		{"SuperUser cannot authenticate with ApiKey", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelApiKey, false},
		{"SuperUser cannot authenticate with None", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelNone, false},

		// Bearer level tests
		{"Bearer can authenticate with SuperUser", ApiKeySecurityLevelBearer, ApiKeySecurityLevelSuperUser, true},
		{"Bearer can authenticate with Bearer", ApiKeySecurityLevelBearer, ApiKeySecurityLevelBearer, true},
		{"Bearer cannot authenticate with ApiKey", ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey, false},
		{"Bearer cannot authenticate with None", ApiKeySecurityLevelBearer, ApiKeySecurityLevelNone, false},

		// ApiKey level tests
		{"ApiKey can authenticate with SuperUser", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelSuperUser, true},
		{"ApiKey cannot authenticate with Bearer", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelBearer, false},
		{"ApiKey can authenticate with ApiKey", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelApiKey, true},
		{"ApiKey cannot authenticate with None", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelNone, false},

		// None level tests
		{"None cannot authenticate with SuperUser", ApiKeySecurityLevelNone, ApiKeySecurityLevelSuperUser, false},
		{"None cannot authenticate with Bearer", ApiKeySecurityLevelNone, ApiKeySecurityLevelBearer, false},
		{"None cannot authenticate with ApiKey", ApiKeySecurityLevelNone, ApiKeySecurityLevelApiKey, false},
		{"None can authenticate with None", ApiKeySecurityLevelNone, ApiKeySecurityLevelNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.CanAuthenticateWith(tt.authType)
			if result != tt.expected {
				t.Errorf("CanAuthenticateWith() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_GetCompatibleAuthTypes(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected []ApiKeySecurityLevel
	}{
		{"Any compatible auth types", ApiKeySecurityLevelAny, []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey}},
		{"SuperUser compatible auth types", ApiKeySecurityLevelSuperUser, []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser}},
		{"Bearer compatible auth types", ApiKeySecurityLevelBearer, []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer}},
		{"ApiKey compatible auth types", ApiKeySecurityLevelApiKey, []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelApiKey}},
		{"None compatible auth types", ApiKeySecurityLevelNone, []ApiKeySecurityLevel{ApiKeySecurityLevelNone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.GetCompatibleAuthTypes()
			if len(result) != len(tt.expected) {
				t.Errorf("GetCompatibleAuthTypes() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("GetCompatibleAuthTypes()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestApiKeySecurityLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected bool
	}{
		{"Any is valid", ApiKeySecurityLevelAny, true},
		{"SuperUser is valid", ApiKeySecurityLevelSuperUser, true},
		{"Bearer is valid", ApiKeySecurityLevelBearer, true},
		{"ApiKey is valid", ApiKeySecurityLevelApiKey, true},
		{"None is valid", ApiKeySecurityLevelNone, true},
		{"Invalid level is not valid", ApiKeySecurityLevel("invalid"), false},
		{"Empty string is not valid", ApiKeySecurityLevel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_GetAuthenticationType(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected string
	}{
		{"Any authentication type", ApiKeySecurityLevelAny, "any"},
		{"SuperUser authentication type", ApiKeySecurityLevelSuperUser, "superuser"},
		{"Bearer authentication type", ApiKeySecurityLevelBearer, "bearer"},
		{"ApiKey authentication type", ApiKeySecurityLevelApiKey, "apikey"},
		{"None authentication type", ApiKeySecurityLevelNone, "none"},
		{"Invalid authentication type", ApiKeySecurityLevel("invalid"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.GetAuthenticationType()
			if result != tt.expected {
				t.Errorf("GetAuthenticationType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected string
	}{
		{"Any string", ApiKeySecurityLevelAny, "any"},
		{"SuperUser string", ApiKeySecurityLevelSuperUser, "superuser"},
		{"Bearer string", ApiKeySecurityLevelBearer, "bearer"},
		{"ApiKey string", ApiKeySecurityLevelApiKey, "apikey"},
		{"None string", ApiKeySecurityLevelNone, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_GetLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    ApiKeySecurityLevel
		expected int
	}{
		{"Any level", ApiKeySecurityLevelAny, 0},
		{"SuperUser level", ApiKeySecurityLevelSuperUser, 1},
		{"Bearer level", ApiKeySecurityLevelBearer, 2},
		{"ApiKey level", ApiKeySecurityLevelApiKey, 3},
		{"None level", ApiKeySecurityLevelNone, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.GetLevel()
			if result != tt.expected {
				t.Errorf("GetLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_IsMoreFlexibleThan(t *testing.T) {
	tests := []struct {
		name     string
		current  ApiKeySecurityLevel
		other    ApiKeySecurityLevel
		expected bool
	}{
		// Any is most flexible
		{"Any is more flexible than SuperUser", ApiKeySecurityLevelAny, ApiKeySecurityLevelSuperUser, true},
		{"Any is more flexible than Bearer", ApiKeySecurityLevelAny, ApiKeySecurityLevelBearer, true},
		{"Any is more flexible than ApiKey", ApiKeySecurityLevelAny, ApiKeySecurityLevelApiKey, true},
		{"Any is more flexible than None", ApiKeySecurityLevelAny, ApiKeySecurityLevelNone, true},

		// SuperUser comparisons
		{"SuperUser is not more flexible than Any", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelAny, false},
		{"SuperUser is more flexible than Bearer", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, true},
		{"SuperUser is more flexible than ApiKey", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelApiKey, true},
		{"SuperUser is more flexible than None", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelNone, true},

		// Bearer comparisons
		{"Bearer is not more flexible than Any", ApiKeySecurityLevelBearer, ApiKeySecurityLevelAny, false},
		{"Bearer is not more flexible than SuperUser", ApiKeySecurityLevelBearer, ApiKeySecurityLevelSuperUser, false},
		{"Bearer is more flexible than ApiKey", ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey, true},
		{"Bearer is more flexible than None", ApiKeySecurityLevelBearer, ApiKeySecurityLevelNone, true},

		// ApiKey comparisons
		{"ApiKey is not more flexible than Any", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelAny, false},
		{"ApiKey is not more flexible than SuperUser", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelSuperUser, false},
		{"ApiKey is not more flexible than Bearer", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelBearer, false},
		{"ApiKey is more flexible than None", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelNone, true},

		// None is least flexible
		{"None is not more flexible than Any", ApiKeySecurityLevelNone, ApiKeySecurityLevelAny, false},
		{"None is not more flexible than SuperUser", ApiKeySecurityLevelNone, ApiKeySecurityLevelSuperUser, false},
		{"None is not more flexible than Bearer", ApiKeySecurityLevelNone, ApiKeySecurityLevelBearer, false},
		{"None is not more flexible than ApiKey", ApiKeySecurityLevelNone, ApiKeySecurityLevelApiKey, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsMoreFlexibleThan(tt.other)
			if result != tt.expected {
				t.Errorf("IsMoreFlexibleThan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_IsMoreRestrictiveThan(t *testing.T) {
	tests := []struct {
		name     string
		current  ApiKeySecurityLevel
		other    ApiKeySecurityLevel
		expected bool
	}{
		// Any is most flexible, so it's not more restrictive than others
		{"Any is not more restrictive than SuperUser", ApiKeySecurityLevelAny, ApiKeySecurityLevelSuperUser, false},
		{"Any is not more restrictive than Bearer", ApiKeySecurityLevelAny, ApiKeySecurityLevelBearer, false},
		{"Any is not more restrictive than ApiKey", ApiKeySecurityLevelAny, ApiKeySecurityLevelApiKey, false},
		{"Any is not more restrictive than None", ApiKeySecurityLevelAny, ApiKeySecurityLevelNone, false},

		// SuperUser is more restrictive than Any, but less restrictive than others
		{"SuperUser is more restrictive than Any", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelAny, true},
		{"SuperUser is not more restrictive than Bearer", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, false},
		{"SuperUser is not more restrictive than ApiKey", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelApiKey, false},
		{"SuperUser is not more restrictive than None", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelNone, false},

		// Bearer is more restrictive than Any and SuperUser
		{"Bearer is more restrictive than Any", ApiKeySecurityLevelBearer, ApiKeySecurityLevelAny, true},
		{"Bearer is more restrictive than SuperUser", ApiKeySecurityLevelBearer, ApiKeySecurityLevelSuperUser, true},
		{"Bearer is not more restrictive than ApiKey", ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey, false},
		{"Bearer is not more restrictive than None", ApiKeySecurityLevelBearer, ApiKeySecurityLevelNone, false},

		// ApiKey is more restrictive than Any, SuperUser, and Bearer
		{"ApiKey is more restrictive than Any", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelAny, true},
		{"ApiKey is more restrictive than SuperUser", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelSuperUser, true},
		{"ApiKey is more restrictive than Bearer", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelBearer, true},
		{"ApiKey is not more restrictive than None", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelNone, false},

		// None is most restrictive
		{"None is more restrictive than Any", ApiKeySecurityLevelNone, ApiKeySecurityLevelAny, true},
		{"None is more restrictive than SuperUser", ApiKeySecurityLevelNone, ApiKeySecurityLevelSuperUser, true},
		{"None is more restrictive than Bearer", ApiKeySecurityLevelNone, ApiKeySecurityLevelBearer, true},
		{"None is more restrictive than ApiKey", ApiKeySecurityLevelNone, ApiKeySecurityLevelApiKey, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsMoreRestrictiveThan(tt.other)
			if result != tt.expected {
				t.Errorf("IsMoreRestrictiveThan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestApiKeySecurityLevel_IsEqual(t *testing.T) {
	tests := []struct {
		name     string
		current  ApiKeySecurityLevel
		other    ApiKeySecurityLevel
		expected bool
	}{
		{"Any equals Any", ApiKeySecurityLevelAny, ApiKeySecurityLevelAny, true},
		{"Any not equals SuperUser", ApiKeySecurityLevelAny, ApiKeySecurityLevelSuperUser, false},
		{"SuperUser equals SuperUser", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelSuperUser, true},
		{"SuperUser not equals Bearer", ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, false},
		{"Bearer equals Bearer", ApiKeySecurityLevelBearer, ApiKeySecurityLevelBearer, true},
		{"Bearer not equals ApiKey", ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey, false},
		{"ApiKey equals ApiKey", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelApiKey, true},
		{"ApiKey not equals None", ApiKeySecurityLevelApiKey, ApiKeySecurityLevelNone, false},
		{"None equals None", ApiKeySecurityLevelNone, ApiKeySecurityLevelNone, true},
		{"None not equals Any", ApiKeySecurityLevelNone, ApiKeySecurityLevelAny, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsEqual(tt.other)
			if result != tt.expected {
				t.Errorf("IsEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test flexibility consistency
func TestApiKeySecurityLevel_FlexibilityConsistency(t *testing.T) {
	levels := []ApiKeySecurityLevel{
		ApiKeySecurityLevelAny,
		ApiKeySecurityLevelSuperUser,
		ApiKeySecurityLevelBearer,
		ApiKeySecurityLevelApiKey,
		ApiKeySecurityLevelNone,
	}

	// Test that IsMoreFlexibleThan and IsMoreRestrictiveThan are consistent
	for i, level1 := range levels {
		for j, level2 := range levels {
			if i < j {
				// level1 should be more flexible than level2
				if !level1.IsMoreFlexibleThan(level2) {
					t.Errorf("%v should be more flexible than %v", level1, level2)
				}
				if !level2.IsMoreRestrictiveThan(level1) {
					t.Errorf("%v should be more restrictive than %v", level2, level1)
				}
			} else if i > j {
				// level1 should be more restrictive than level2
				if !level1.IsMoreRestrictiveThan(level2) {
					t.Errorf("%v should be more restrictive than %v", level1, level2)
				}
				if !level2.IsMoreFlexibleThan(level1) {
					t.Errorf("%v should be more flexible than %v", level2, level1)
				}
			} else {
				// Same level
				if !level1.IsEqual(level2) {
					t.Errorf("%v should be equal to %v", level1, level2)
				}
			}
		}
	}
}

// Test authentication compatibility consistency
func TestApiKeySecurityLevel_AuthenticationCompatibilityConsistency(t *testing.T) {
	levels := []ApiKeySecurityLevel{
		ApiKeySecurityLevelAny,
		ApiKeySecurityLevelSuperUser,
		ApiKeySecurityLevelBearer,
		ApiKeySecurityLevelApiKey,
		ApiKeySecurityLevelNone,
	}

	for _, level := range levels {
		compatibleTypes := level.GetCompatibleAuthTypes()

		// Test that all compatible types can authenticate with this level
		for _, authType := range compatibleTypes {
			if !level.CanAuthenticateWith(authType) {
				t.Errorf("%v should be able to authenticate with %v", level, authType)
			}
		}

		// Test that incompatible types cannot authenticate
		for _, testAuthType := range levels {
			if !level.CanAuthenticateWith(testAuthType) {
				// This should be incompatible
				found := false
				for _, compatibleType := range compatibleTypes {
					if compatibleType == testAuthType {
						found = true
						break
					}
				}
				if found {
					t.Errorf("%v should not be compatible with %v but CanAuthenticateWith returned true", level, testAuthType)
				}
			}
		}

		// Valid levels should be valid
		if !level.IsValid() {
			t.Errorf("%v should be valid", level)
		}
	}
}
