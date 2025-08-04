package models

type ApiKeySecurityLevel string

const (
	ApiKeySecurityLevelAny       ApiKeySecurityLevel = "any"
	ApiKeySecurityLevelSuperUser ApiKeySecurityLevel = "superuser"
	ApiKeySecurityLevelBearer    ApiKeySecurityLevel = "bearer"
	ApiKeySecurityLevelApiKey    ApiKeySecurityLevel = "apikey"
	ApiKeySecurityLevelNone      ApiKeySecurityLevel = "none"
)

// RequiresAuthentication checks if the current level requires authentication
// Returns true if the level is not "none", false otherwise
func (a ApiKeySecurityLevel) RequiresAuthentication() bool {
	// If the level is not valid, it doesn't require authentication
	if !a.IsValid() {
		return false
	}
	return a != ApiKeySecurityLevelNone
}

// CanAuthenticateWith checks if the current level can authenticate with the given authentication type
func (a ApiKeySecurityLevel) CanAuthenticateWith(authType ApiKeySecurityLevel) bool {
	switch a {
	case ApiKeySecurityLevelAny:
		return authType == ApiKeySecurityLevelSuperUser || authType == ApiKeySecurityLevelBearer || authType == ApiKeySecurityLevelApiKey
	case ApiKeySecurityLevelSuperUser:
		return authType == ApiKeySecurityLevelSuperUser
	case ApiKeySecurityLevelBearer:
		return authType == ApiKeySecurityLevelSuperUser || authType == ApiKeySecurityLevelBearer
	case ApiKeySecurityLevelApiKey:
		return authType == ApiKeySecurityLevelSuperUser || authType == ApiKeySecurityLevelApiKey
	case ApiKeySecurityLevelNone:
		return authType == ApiKeySecurityLevelNone
	}
	return false
}

// GetCompatibleAuthTypes returns all authentication types that are compatible with this level
func (a ApiKeySecurityLevel) GetCompatibleAuthTypes() []ApiKeySecurityLevel {
	switch a {
	case ApiKeySecurityLevelAny:
		return []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey}
	case ApiKeySecurityLevelSuperUser:
		return []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser}
	case ApiKeySecurityLevelBearer:
		return []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer}
	case ApiKeySecurityLevelApiKey:
		return []ApiKeySecurityLevel{ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelApiKey}
	case ApiKeySecurityLevelNone:
		return []ApiKeySecurityLevel{ApiKeySecurityLevelNone}
	}
	return []ApiKeySecurityLevel{}
}

// IsValid checks if the API security level is a valid value
func (a ApiKeySecurityLevel) IsValid() bool {
	switch a {
	case ApiKeySecurityLevelAny, ApiKeySecurityLevelSuperUser, ApiKeySecurityLevelBearer, ApiKeySecurityLevelApiKey, ApiKeySecurityLevelNone:
		return true
	default:
		return false
	}
}

// GetAuthenticationType returns the type of authentication required for this level
func (a ApiKeySecurityLevel) GetAuthenticationType() string {
	switch a {
	case ApiKeySecurityLevelAny:
		return "any"
	case ApiKeySecurityLevelSuperUser:
		return "superuser"
	case ApiKeySecurityLevelBearer:
		return "bearer"
	case ApiKeySecurityLevelApiKey:
		return "apikey"
	case ApiKeySecurityLevelNone:
		return "none"
	default:
		return "unknown"
	}
}

// String returns the string representation of the API security level
func (a ApiKeySecurityLevel) String() string {
	return string(a)
}

// GetLevel returns the numeric level (0 = most flexible, 4 = most restrictive)
func (a ApiKeySecurityLevel) GetLevel() int {
	switch a {
	case ApiKeySecurityLevelAny:
		return 0
	case ApiKeySecurityLevelSuperUser:
		return 1
	case ApiKeySecurityLevelBearer:
		return 2
	case ApiKeySecurityLevelApiKey:
		return 3
	case ApiKeySecurityLevelNone:
		return 4
	}
	return 4 // Default to most restrictive level
}

// IsMoreFlexibleThan checks if the current level is more flexible than the other level
func (a ApiKeySecurityLevel) IsMoreFlexibleThan(other ApiKeySecurityLevel) bool {
	// Define flexibility order (lower index = more flexible)
	flexibility := []ApiKeySecurityLevel{
		ApiKeySecurityLevelAny,       // Most flexible
		ApiKeySecurityLevelSuperUser, // Only superuser
		ApiKeySecurityLevelBearer,    // Superuser or bearer
		ApiKeySecurityLevelApiKey,    // Superuser or apikey
		ApiKeySecurityLevelNone,      // No auth required
	}

	// Find positions in flexibility order
	var currentIndex, otherIndex int
	for i, level := range flexibility {
		if level == a {
			currentIndex = i
		}
		if level == other {
			otherIndex = i
		}
	}

	// Lower index means more flexible
	return currentIndex < otherIndex
}

// IsMoreRestrictiveThan checks if the current level is more restrictive than the other level
func (a ApiKeySecurityLevel) IsMoreRestrictiveThan(other ApiKeySecurityLevel) bool {
	return other.IsMoreFlexibleThan(a)
}

// IsEqual checks if the current level has the same flexibility as the other level
func (a ApiKeySecurityLevel) IsEqual(other ApiKeySecurityLevel) bool {
	return a == other
}
