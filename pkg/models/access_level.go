package models

type AccessLevel string

const (
	AccessLevelRead    AccessLevel = "read"
	AccessLevelWrite   AccessLevel = "write"
	AccessLevelDelete  AccessLevel = "delete"
	AccessLevelAll     AccessLevel = "*"
	AccessLevelNone    AccessLevel = "none"
	AccessLevelUpdate  AccessLevel = "update"
	AccessLevelCreate  AccessLevel = "create"
	AccessLevelView    AccessLevel = "view"
	AccessLevelApprove AccessLevel = "approve"
	AccessLevelReject  AccessLevel = "reject"
	AccessLevelCancel  AccessLevel = "cancel"
	AccessLevelSuspend AccessLevel = "suspend"
	AccessLevelResume  AccessLevel = "resume"
	AccessLevelReset   AccessLevel = "reset"
	AccessLevelUnlock  AccessLevel = "unlock"
	AccessLevelLock    AccessLevel = "lock"
	AccessLevelRevoke  AccessLevel = "revoke"
	AccessLevelAudit   AccessLevel = "audit"
)

// IsParentOf checks if the current level is a parent of the given level
// A parent has higher privileges than its children
func (a AccessLevel) IsParentOf(child AccessLevel) bool {
	switch a {
	case AccessLevelAll:
		return child != AccessLevelAll // All includes everything except itself
	case AccessLevelWrite:
		return child == AccessLevelRead || child == AccessLevelUpdate || child == AccessLevelCreate || child == AccessLevelView
	case AccessLevelUpdate:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelCreate:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelRead:
		return child == AccessLevelView
	case AccessLevelView:
		return child == AccessLevelNone
	case AccessLevelDelete:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelApprove:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelReject:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelCancel:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelSuspend:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelResume:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelReset:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelLock:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelUnlock:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelRevoke:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelAudit:
		return child == AccessLevelRead || child == AccessLevelView
	case AccessLevelNone:
		return false
	}
	return false
}

// IsChildOf checks if the current level is a child of the given level
// A child has lower privileges than its parent
func (a AccessLevel) IsChildOf(parent AccessLevel) bool {
	switch a {
	case AccessLevelAll:
		return false // All has no parent
	case AccessLevelWrite:
		return parent == AccessLevelAll
	case AccessLevelUpdate:
		return parent == AccessLevelAll || parent == AccessLevelWrite
	case AccessLevelCreate:
		return parent == AccessLevelAll || parent == AccessLevelWrite
	case AccessLevelDelete:
		return parent == AccessLevelAll
	case AccessLevelRead:
		return parent == AccessLevelAll || parent == AccessLevelWrite || parent == AccessLevelUpdate || parent == AccessLevelCreate || parent == AccessLevelDelete || parent == AccessLevelApprove || parent == AccessLevelReject || parent == AccessLevelCancel || parent == AccessLevelSuspend || parent == AccessLevelResume || parent == AccessLevelReset || parent == AccessLevelLock || parent == AccessLevelUnlock || parent == AccessLevelRevoke || parent == AccessLevelAudit
	case AccessLevelView:
		return parent == AccessLevelAll || parent == AccessLevelWrite || parent == AccessLevelUpdate || parent == AccessLevelCreate || parent == AccessLevelDelete || parent == AccessLevelApprove || parent == AccessLevelReject || parent == AccessLevelCancel || parent == AccessLevelSuspend || parent == AccessLevelResume || parent == AccessLevelReset || parent == AccessLevelLock || parent == AccessLevelUnlock || parent == AccessLevelRead || parent == AccessLevelRevoke || parent == AccessLevelAudit
	case AccessLevelApprove:
		return parent == AccessLevelAll
	case AccessLevelReject:
		return parent == AccessLevelAll
	case AccessLevelCancel:
		return parent == AccessLevelAll
	case AccessLevelSuspend:
		return parent == AccessLevelAll
	case AccessLevelResume:
		return parent == AccessLevelAll
	case AccessLevelReset:
		return parent == AccessLevelAll
	case AccessLevelLock:
		return parent == AccessLevelAll
	case AccessLevelUnlock:
		return parent == AccessLevelAll
	case AccessLevelRevoke:
		return parent == AccessLevelAll
	case AccessLevelAudit:
		return parent == AccessLevelAll
	case AccessLevelNone:
		return parent == AccessLevelAll || parent == AccessLevelWrite || parent == AccessLevelUpdate || parent == AccessLevelCreate || parent == AccessLevelDelete || parent == AccessLevelApprove || parent == AccessLevelReject || parent == AccessLevelCancel || parent == AccessLevelSuspend || parent == AccessLevelResume || parent == AccessLevelReset || parent == AccessLevelLock || parent == AccessLevelUnlock || parent == AccessLevelRead || parent == AccessLevelView || parent == AccessLevelRevoke || parent == AccessLevelAudit
	}
	return false
}

// IsHigherThan checks if the current level has higher privileges than the other level
func (a AccessLevel) IsHigherThan(other AccessLevel) bool {
	// Define hierarchy order (lower index = higher privilege)
	hierarchy := []AccessLevel{
		AccessLevelAll,
		AccessLevelWrite,
		AccessLevelUpdate,
		AccessLevelCreate,
		AccessLevelDelete,
		AccessLevelApprove,
		AccessLevelReject,
		AccessLevelCancel,
		AccessLevelSuspend,
		AccessLevelResume,
		AccessLevelReset,
		AccessLevelLock,
		AccessLevelUnlock,
		AccessLevelRevoke,
		AccessLevelAudit,
		AccessLevelRead,
		AccessLevelView,
		AccessLevelNone,
	}

	// Find positions in hierarchy
	var currentIndex, otherIndex int
	for i, level := range hierarchy {
		if level == a {
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
func (a AccessLevel) IsLowerThan(other AccessLevel) bool {
	return other.IsHigherThan(a)
}

// IsEqual checks if the current level has the same privileges as the other level
func (a AccessLevel) IsEqual(other AccessLevel) bool {
	return a == other
}

// CanAccess checks if the current level can access the required level
// This is the main method used for claim validation
func (a AccessLevel) CanAccess(required AccessLevel) bool {
	if a == AccessLevelAll {
		return true // All can access everything
	}
	if a == required {
		return true // Same level can access itself
	}
	if a.IsParentOf(required) {
		return true // Parent can access child
	}
	if a.IsHigherThan(required) {
		return true // Higher level can access lower level
	}
	return false
}

// String returns the string representation of the access level
func (a AccessLevel) String() string {
	return string(a)
}
