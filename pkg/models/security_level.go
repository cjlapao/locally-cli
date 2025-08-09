package models

type SecurityLevel string

const (
	SecurityLevelSuperUser SecurityLevel = "superuser"
	SecurityLevelAdmin     SecurityLevel = "admin"
	SecurityLevelManager   SecurityLevel = "manager"
	SecurityLevelUser      SecurityLevel = "user"
	SecurityLevelAuditor   SecurityLevel = "auditor"
	SecurityLevelGuest     SecurityLevel = "guest"
	SecurityLevelNone      SecurityLevel = "none"
)

// IsParentOf checks if the current level is a parent of the given level
// A parent has higher privileges than its children
func (c SecurityLevel) IsParentOf(child SecurityLevel) bool {
	switch c {
	case SecurityLevelSuperUser:
		return child == SecurityLevelAdmin || child == SecurityLevelManager || child == SecurityLevelUser || child == SecurityLevelAuditor || child == SecurityLevelGuest || child == SecurityLevelNone
	case SecurityLevelAdmin:
		return child == SecurityLevelManager || child == SecurityLevelUser || child == SecurityLevelAuditor || child == SecurityLevelGuest || child == SecurityLevelNone
	case SecurityLevelManager:
		return child == SecurityLevelUser || child == SecurityLevelAuditor || child == SecurityLevelGuest || child == SecurityLevelNone
	case SecurityLevelUser:
		return child == SecurityLevelAuditor || child == SecurityLevelGuest || child == SecurityLevelNone
	case SecurityLevelAuditor:
		return child == SecurityLevelGuest || child == SecurityLevelNone
	case SecurityLevelGuest:
		return child == SecurityLevelNone
	case SecurityLevelNone:
		return false
	}
	return false
}

// IsChildOf checks if the current level is a child of the given level
// A child has lower privileges than its parent
func (c SecurityLevel) IsChildOf(parent SecurityLevel) bool {
	switch c {
	case SecurityLevelSuperUser:
		return false // SuperUser has no parent
	case SecurityLevelAdmin:
		return parent == SecurityLevelSuperUser
	case SecurityLevelManager:
		return parent == SecurityLevelSuperUser || parent == SecurityLevelAdmin
	case SecurityLevelUser:
		return parent == SecurityLevelSuperUser || parent == SecurityLevelAdmin || parent == SecurityLevelManager
	case SecurityLevelAuditor:
		return parent == SecurityLevelSuperUser || parent == SecurityLevelAdmin || parent == SecurityLevelManager || parent == SecurityLevelUser
	case SecurityLevelGuest:
		return parent == SecurityLevelSuperUser || parent == SecurityLevelAdmin || parent == SecurityLevelManager || parent == SecurityLevelUser || parent == SecurityLevelAuditor
	case SecurityLevelNone:
		return parent == SecurityLevelSuperUser || parent == SecurityLevelAdmin || parent == SecurityLevelManager || parent == SecurityLevelUser || parent == SecurityLevelAuditor || parent == SecurityLevelGuest
	}
	return false
}

// GetParent returns the immediate parent of the current level
func (c SecurityLevel) GetParent() SecurityLevel {
	switch c {
	case SecurityLevelSuperUser:
		return SecurityLevelSuperUser // No parent, return self
	case SecurityLevelAdmin:
		return SecurityLevelSuperUser
	case SecurityLevelManager:
		return SecurityLevelAdmin
	case SecurityLevelUser:
		return SecurityLevelManager
	case SecurityLevelAuditor:
		return SecurityLevelUser
	case SecurityLevelGuest:
		return SecurityLevelAuditor
	case SecurityLevelNone:
		return SecurityLevelGuest
	}
	return SecurityLevelSuperUser
}

// GetChildren returns all immediate children of the current level
func (c SecurityLevel) GetChildren() []SecurityLevel {
	switch c {
	case SecurityLevelSuperUser:
		return []SecurityLevel{SecurityLevelAdmin}
	case SecurityLevelAdmin:
		return []SecurityLevel{SecurityLevelManager}
	case SecurityLevelManager:
		return []SecurityLevel{SecurityLevelUser}
	case SecurityLevelUser:
		return []SecurityLevel{SecurityLevelAuditor}
	case SecurityLevelAuditor:
		return []SecurityLevel{SecurityLevelGuest}
	case SecurityLevelGuest:
		return []SecurityLevel{SecurityLevelNone}
	case SecurityLevelNone:
		return []SecurityLevel{}
	}
	return []SecurityLevel{}
}

// GetAllChildren returns all descendants (children, grandchildren, etc.) of the current level
func (c SecurityLevel) GetAllChildren() []SecurityLevel {
	switch c {
	case SecurityLevelSuperUser:
		return []SecurityLevel{SecurityLevelAdmin, SecurityLevelManager, SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}
	case SecurityLevelAdmin:
		return []SecurityLevel{SecurityLevelManager, SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}
	case SecurityLevelManager:
		return []SecurityLevel{SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}
	case SecurityLevelUser:
		return []SecurityLevel{SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}
	case SecurityLevelAuditor:
		return []SecurityLevel{SecurityLevelGuest, SecurityLevelNone}
	case SecurityLevelGuest:
		return []SecurityLevel{SecurityLevelNone}
	case SecurityLevelNone:
		return []SecurityLevel{}
	}
	return []SecurityLevel{}
}

// IsHigherThan checks if the current level has higher privileges than the other level
func (c SecurityLevel) IsHigherThan(other SecurityLevel) bool {
	// Define hierarchy order (lower index = higher privilege)
	hierarchy := []SecurityLevel{
		SecurityLevelSuperUser,
		SecurityLevelAdmin,
		SecurityLevelManager,
		SecurityLevelUser,
		SecurityLevelAuditor,
		SecurityLevelGuest,
		SecurityLevelNone,
	}

	// Find positions in hierarchy
	var currentIndex, otherIndex int
	for i, level := range hierarchy {
		if level == c {
			currentIndex = i
		}
		if level == other {
			otherIndex = i
		}
	}

	// Lower index means higher privilege
	return currentIndex < otherIndex
}

// IsLowerThan checks if the current level has lower privileges than the other level
func (c SecurityLevel) IsLowerThan(other SecurityLevel) bool {
	return other.IsHigherThan(c)
}

// IsEqual checks if the current level has the same privileges as the other level
func (c SecurityLevel) IsEqual(other SecurityLevel) bool {
	return c == other
}

// GetLevel returns the numeric level (0 = highest privilege, 6 = lowest privilege)
func (c SecurityLevel) GetLevel() int {
	switch c {
	case SecurityLevelSuperUser:
		return 0
	case SecurityLevelAdmin:
		return 1
	case SecurityLevelManager:
		return 2
	case SecurityLevelUser:
		return 3
	case SecurityLevelAuditor:
		return 4
	case SecurityLevelGuest:
		return 5
	case SecurityLevelNone:
		return 6
	}
	return 6 // Default to lowest level
}

// String returns the string representation of the security level
func (c SecurityLevel) String() string {
	return string(c)
}
