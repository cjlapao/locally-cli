package models

import (
	"testing"
)

func TestAccessLevel_IsParentOf(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		child    AccessLevel
		expected bool
	}{
		// AccessLevelAll tests
		{"All is parent of Write", AccessLevelAll, AccessLevelWrite, true},
		{"All is parent of Read", AccessLevelAll, AccessLevelRead, true},
		{"All is parent of View", AccessLevelAll, AccessLevelView, true},
		{"All is parent of None", AccessLevelAll, AccessLevelNone, true},
		{"All is parent of Audit", AccessLevelAll, AccessLevelAudit, true},
		{"All is not parent of All", AccessLevelAll, AccessLevelAll, false},

		// AccessLevelWrite tests
		{"Write is parent of Read", AccessLevelWrite, AccessLevelRead, true},
		{"Write is parent of Update", AccessLevelWrite, AccessLevelUpdate, true},
		{"Write is parent of Create", AccessLevelWrite, AccessLevelCreate, true},
		{"Write is parent of View", AccessLevelWrite, AccessLevelView, true},
		{"Write is not parent of Write", AccessLevelWrite, AccessLevelWrite, false},
		{"Write is not parent of All", AccessLevelWrite, AccessLevelAll, false},
		{"Write is not parent of Delete", AccessLevelWrite, AccessLevelDelete, false},

		// AccessLevelUpdate tests
		{"Update is parent of Read", AccessLevelUpdate, AccessLevelRead, true},
		{"Update is parent of View", AccessLevelUpdate, AccessLevelView, true},
		{"Update is not parent of Update", AccessLevelUpdate, AccessLevelUpdate, false},
		{"Update is not parent of Write", AccessLevelUpdate, AccessLevelWrite, false},

		// AccessLevelCreate tests
		{"Create is parent of Read", AccessLevelCreate, AccessLevelRead, true},
		{"Create is parent of View", AccessLevelCreate, AccessLevelView, true},
		{"Create is not parent of Create", AccessLevelCreate, AccessLevelCreate, false},
		{"Create is not parent of Write", AccessLevelCreate, AccessLevelWrite, false},

		// AccessLevelRead tests
		{"Read is parent of View", AccessLevelRead, AccessLevelView, true},
		{"Read is not parent of Read", AccessLevelRead, AccessLevelRead, false},
		{"Read is not parent of Write", AccessLevelRead, AccessLevelWrite, false},

		// AccessLevelView tests
		{"View is parent of None", AccessLevelView, AccessLevelNone, true},
		{"View is not parent of View", AccessLevelView, AccessLevelView, false},
		{"View is not parent of Read", AccessLevelView, AccessLevelRead, false},

		// AccessLevelDelete tests
		{"Delete is parent of Read", AccessLevelDelete, AccessLevelRead, true},
		{"Delete is parent of View", AccessLevelDelete, AccessLevelView, true},
		{"Delete is not parent of Delete", AccessLevelDelete, AccessLevelDelete, false},

		// AccessLevelApprove tests
		{"Approve is parent of Read", AccessLevelApprove, AccessLevelRead, true},
		{"Approve is parent of View", AccessLevelApprove, AccessLevelView, true},
		{"Approve is not parent of Approve", AccessLevelApprove, AccessLevelApprove, false},

		// AccessLevelReject tests
		{"Reject is parent of Read", AccessLevelReject, AccessLevelRead, true},
		{"Reject is parent of View", AccessLevelReject, AccessLevelView, true},
		{"Reject is not parent of Reject", AccessLevelReject, AccessLevelReject, false},

		// AccessLevelCancel tests
		{"Cancel is parent of Read", AccessLevelCancel, AccessLevelRead, true},
		{"Cancel is parent of View", AccessLevelCancel, AccessLevelView, true},
		{"Cancel is not parent of Cancel", AccessLevelCancel, AccessLevelCancel, false},

		// AccessLevelSuspend tests
		{"Suspend is parent of Read", AccessLevelSuspend, AccessLevelRead, true},
		{"Suspend is parent of View", AccessLevelSuspend, AccessLevelView, true},
		{"Suspend is not parent of Suspend", AccessLevelSuspend, AccessLevelSuspend, false},

		// AccessLevelResume tests
		{"Resume is parent of Read", AccessLevelResume, AccessLevelRead, true},
		{"Resume is parent of View", AccessLevelResume, AccessLevelView, true},
		{"Resume is not parent of Resume", AccessLevelResume, AccessLevelResume, false},

		// AccessLevelReset tests
		{"Reset is parent of Read", AccessLevelReset, AccessLevelRead, true},
		{"Reset is parent of View", AccessLevelReset, AccessLevelView, true},
		{"Reset is not parent of Reset", AccessLevelReset, AccessLevelReset, false},

		// AccessLevelLock tests
		{"Lock is parent of Read", AccessLevelLock, AccessLevelRead, true},
		{"Lock is parent of View", AccessLevelLock, AccessLevelView, true},
		{"Lock is not parent of Lock", AccessLevelLock, AccessLevelLock, false},

		// AccessLevelUnlock tests
		{"Unlock is parent of Read", AccessLevelUnlock, AccessLevelRead, true},
		{"Unlock is parent of View", AccessLevelUnlock, AccessLevelView, true},
		{"Unlock is not parent of Unlock", AccessLevelUnlock, AccessLevelUnlock, false},

		// AccessLevelRevoke tests
		{"Revoke is parent of Read", AccessLevelRevoke, AccessLevelRead, true},
		{"Revoke is parent of View", AccessLevelRevoke, AccessLevelView, true},
		{"Revoke is not parent of Revoke", AccessLevelRevoke, AccessLevelRevoke, false},

		// AccessLevelAudit tests
		{"Audit is parent of Read", AccessLevelAudit, AccessLevelRead, true},
		{"Audit is parent of View", AccessLevelAudit, AccessLevelView, true},
		{"Audit is not parent of Audit", AccessLevelAudit, AccessLevelAudit, false},

		// AccessLevelNone tests
		{"None is not parent of anyone", AccessLevelNone, AccessLevelAll, false},
		{"None is not parent of Read", AccessLevelNone, AccessLevelRead, false},
		{"None is not parent of None", AccessLevelNone, AccessLevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsParentOf(tt.child)
			if result != tt.expected {
				t.Errorf("IsParentOf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_IsChildOf(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		parent   AccessLevel
		expected bool
	}{
		// AccessLevelAll tests
		{"All is not child of anyone", AccessLevelAll, AccessLevelAll, false},
		{"All is not child of Write", AccessLevelAll, AccessLevelWrite, false},
		{"All is not child of Read", AccessLevelAll, AccessLevelRead, false},

		// AccessLevelWrite tests
		{"Write is child of All", AccessLevelWrite, AccessLevelAll, true},
		{"Write is not child of Write", AccessLevelWrite, AccessLevelWrite, false},
		{"Write is not child of Read", AccessLevelWrite, AccessLevelRead, false},

		// AccessLevelUpdate tests
		{"Update is child of All", AccessLevelUpdate, AccessLevelAll, true},
		{"Update is child of Write", AccessLevelUpdate, AccessLevelWrite, true},
		{"Update is not child of Update", AccessLevelUpdate, AccessLevelUpdate, false},

		// AccessLevelCreate tests
		{"Create is child of All", AccessLevelCreate, AccessLevelAll, true},
		{"Create is child of Write", AccessLevelCreate, AccessLevelWrite, true},
		{"Create is not child of Create", AccessLevelCreate, AccessLevelCreate, false},

		// AccessLevelRead tests
		{"Read is child of All", AccessLevelRead, AccessLevelAll, true},
		{"Read is child of Write", AccessLevelRead, AccessLevelWrite, true},
		{"Read is child of Update", AccessLevelRead, AccessLevelUpdate, true},
		{"Read is child of Create", AccessLevelRead, AccessLevelCreate, true},
		{"Read is child of Delete", AccessLevelRead, AccessLevelDelete, true},
		{"Read is child of Approve", AccessLevelRead, AccessLevelApprove, true},
		{"Read is child of Reject", AccessLevelRead, AccessLevelReject, true},
		{"Read is child of Cancel", AccessLevelRead, AccessLevelCancel, true},
		{"Read is child of Suspend", AccessLevelRead, AccessLevelSuspend, true},
		{"Read is child of Resume", AccessLevelRead, AccessLevelResume, true},
		{"Read is child of Reset", AccessLevelRead, AccessLevelReset, true},
		{"Read is child of Lock", AccessLevelRead, AccessLevelLock, true},
		{"Read is child of Unlock", AccessLevelRead, AccessLevelUnlock, true},
		{"Read is child of Revoke", AccessLevelRead, AccessLevelRevoke, true},
		{"Read is child of Audit", AccessLevelRead, AccessLevelAudit, true},
		{"Read is not child of Read", AccessLevelRead, AccessLevelRead, false},

		// AccessLevelView tests
		{"View is child of All", AccessLevelView, AccessLevelAll, true},
		{"View is child of Write", AccessLevelView, AccessLevelWrite, true},
		{"View is child of Update", AccessLevelView, AccessLevelUpdate, true},
		{"View is child of Create", AccessLevelView, AccessLevelCreate, true},
		{"View is child of Delete", AccessLevelView, AccessLevelDelete, true},
		{"View is child of Approve", AccessLevelView, AccessLevelApprove, true},
		{"View is child of Reject", AccessLevelView, AccessLevelReject, true},
		{"View is child of Cancel", AccessLevelView, AccessLevelCancel, true},
		{"View is child of Suspend", AccessLevelView, AccessLevelSuspend, true},
		{"View is child of Resume", AccessLevelView, AccessLevelResume, true},
		{"View is child of Reset", AccessLevelView, AccessLevelReset, true},
		{"View is child of Lock", AccessLevelView, AccessLevelLock, true},
		{"View is child of Unlock", AccessLevelView, AccessLevelUnlock, true},
		{"View is child of Read", AccessLevelView, AccessLevelRead, true},
		{"View is child of Revoke", AccessLevelView, AccessLevelRevoke, true},
		{"View is child of Audit", AccessLevelView, AccessLevelAudit, true},
		{"View is not child of View", AccessLevelView, AccessLevelView, false},

		// AccessLevelDelete tests
		{"Delete is child of All", AccessLevelDelete, AccessLevelAll, true},
		{"Delete is not child of Delete", AccessLevelDelete, AccessLevelDelete, false},
		{"Delete is not child of Write", AccessLevelDelete, AccessLevelWrite, false},

		// AccessLevelApprove tests
		{"Approve is child of All", AccessLevelApprove, AccessLevelAll, true},
		{"Approve is not child of Approve", AccessLevelApprove, AccessLevelApprove, false},

		// AccessLevelReject tests
		{"Reject is child of All", AccessLevelReject, AccessLevelAll, true},
		{"Reject is not child of Reject", AccessLevelReject, AccessLevelReject, false},

		// AccessLevelCancel tests
		{"Cancel is child of All", AccessLevelCancel, AccessLevelAll, true},
		{"Cancel is not child of Cancel", AccessLevelCancel, AccessLevelCancel, false},

		// AccessLevelSuspend tests
		{"Suspend is child of All", AccessLevelSuspend, AccessLevelAll, true},
		{"Suspend is not child of Suspend", AccessLevelSuspend, AccessLevelSuspend, false},

		// AccessLevelResume tests
		{"Resume is child of All", AccessLevelResume, AccessLevelAll, true},
		{"Resume is not child of Resume", AccessLevelResume, AccessLevelResume, false},

		// AccessLevelReset tests
		{"Reset is child of All", AccessLevelReset, AccessLevelAll, true},
		{"Reset is not child of Reset", AccessLevelReset, AccessLevelReset, false},

		// AccessLevelLock tests
		{"Lock is child of All", AccessLevelLock, AccessLevelAll, true},
		{"Lock is not child of Lock", AccessLevelLock, AccessLevelLock, false},

		// AccessLevelUnlock tests
		{"Unlock is child of All", AccessLevelUnlock, AccessLevelAll, true},
		{"Unlock is not child of Unlock", AccessLevelUnlock, AccessLevelUnlock, false},

		// AccessLevelRevoke tests
		{"Revoke is child of All", AccessLevelRevoke, AccessLevelAll, true},
		{"Revoke is not child of Revoke", AccessLevelRevoke, AccessLevelRevoke, false},

		// AccessLevelAudit tests
		{"Audit is child of All", AccessLevelAudit, AccessLevelAll, true},
		{"Audit is not child of Audit", AccessLevelAudit, AccessLevelAudit, false},

		// AccessLevelNone tests
		{"None is child of All", AccessLevelNone, AccessLevelAll, true},
		{"None is child of Write", AccessLevelNone, AccessLevelWrite, true},
		{"None is child of Update", AccessLevelNone, AccessLevelUpdate, true},
		{"None is child of Create", AccessLevelNone, AccessLevelCreate, true},
		{"None is child of Delete", AccessLevelNone, AccessLevelDelete, true},
		{"None is child of Approve", AccessLevelNone, AccessLevelApprove, true},
		{"None is child of Reject", AccessLevelNone, AccessLevelReject, true},
		{"None is child of Cancel", AccessLevelNone, AccessLevelCancel, true},
		{"None is child of Suspend", AccessLevelNone, AccessLevelSuspend, true},
		{"None is child of Resume", AccessLevelNone, AccessLevelResume, true},
		{"None is child of Reset", AccessLevelNone, AccessLevelReset, true},
		{"None is child of Lock", AccessLevelNone, AccessLevelLock, true},
		{"None is child of Unlock", AccessLevelNone, AccessLevelUnlock, true},
		{"None is child of Read", AccessLevelNone, AccessLevelRead, true},
		{"None is child of View", AccessLevelNone, AccessLevelView, true},
		{"None is child of Revoke", AccessLevelNone, AccessLevelRevoke, true},
		{"None is child of Audit", AccessLevelNone, AccessLevelAudit, true},
		{"None is not child of None", AccessLevelNone, AccessLevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsChildOf(tt.parent)
			if result != tt.expected {
				t.Errorf("IsChildOf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_IsHigherThan(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		other    AccessLevel
		expected bool
	}{
		// AccessLevelAll comparisons
		{"All is higher than Write", AccessLevelAll, AccessLevelWrite, true},
		{"All is higher than Read", AccessLevelAll, AccessLevelRead, true},
		{"All is higher than View", AccessLevelAll, AccessLevelView, true},
		{"All is higher than None", AccessLevelAll, AccessLevelNone, true},
		{"All is higher than Audit", AccessLevelAll, AccessLevelAudit, true},
		{"All is not higher than All", AccessLevelAll, AccessLevelAll, false},

		// AccessLevelWrite comparisons
		{"Write is not higher than All", AccessLevelWrite, AccessLevelAll, false},
		{"Write is higher than Read", AccessLevelWrite, AccessLevelRead, true},
		{"Write is higher than View", AccessLevelWrite, AccessLevelView, true},
		{"Write is higher than None", AccessLevelWrite, AccessLevelNone, true},
		{"Write is higher than Audit", AccessLevelWrite, AccessLevelAudit, true},
		{"Write is not higher than Write", AccessLevelWrite, AccessLevelWrite, false},

		// AccessLevelRead comparisons
		{"Read is not higher than All", AccessLevelRead, AccessLevelAll, false},
		{"Read is not higher than Write", AccessLevelRead, AccessLevelWrite, false},
		{"Read is higher than View", AccessLevelRead, AccessLevelView, true},
		{"Read is higher than None", AccessLevelRead, AccessLevelNone, true},
		{"Read is not higher than Audit", AccessLevelRead, AccessLevelAudit, false},
		{"Read is not higher than Read", AccessLevelRead, AccessLevelRead, false},

		// AccessLevelView comparisons
		{"View is not higher than All", AccessLevelView, AccessLevelAll, false},
		{"View is not higher than Write", AccessLevelView, AccessLevelWrite, false},
		{"View is not higher than Read", AccessLevelView, AccessLevelRead, false},
		{"View is not higher than Audit", AccessLevelView, AccessLevelAudit, false},
		{"View is higher than None", AccessLevelView, AccessLevelNone, true},
		{"View is not higher than View", AccessLevelView, AccessLevelView, false},

		// AccessLevelAudit comparisons
		{"Audit is not higher than All", AccessLevelAudit, AccessLevelAll, false},
		{"Audit is not higher than Write", AccessLevelAudit, AccessLevelWrite, false},
		{"Audit is higher than Read", AccessLevelAudit, AccessLevelRead, true},
		{"Audit is higher than View", AccessLevelAudit, AccessLevelView, true},
		{"Audit is higher than None", AccessLevelAudit, AccessLevelNone, true},
		{"Audit is not higher than Audit", AccessLevelAudit, AccessLevelAudit, false},

		// AccessLevelNone comparisons
		{"None is not higher than All", AccessLevelNone, AccessLevelAll, false},
		{"None is not higher than Write", AccessLevelNone, AccessLevelWrite, false},
		{"None is not higher than Read", AccessLevelNone, AccessLevelRead, false},
		{"None is not higher than View", AccessLevelNone, AccessLevelView, false},
		{"None is not higher than Audit", AccessLevelNone, AccessLevelAudit, false},
		{"None is not higher than None", AccessLevelNone, AccessLevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsHigherThan(tt.other)
			if result != tt.expected {
				t.Errorf("IsHigherThan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_IsLowerThan(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		other    AccessLevel
		expected bool
	}{
		// AccessLevelAll comparisons
		{"All is not lower than Write", AccessLevelAll, AccessLevelWrite, false},
		{"All is not lower than Read", AccessLevelAll, AccessLevelRead, false},
		{"All is not lower than View", AccessLevelAll, AccessLevelView, false},
		{"All is not lower than None", AccessLevelAll, AccessLevelNone, false},
		{"All is not lower than Audit", AccessLevelAll, AccessLevelAudit, false},
		{"All is not lower than All", AccessLevelAll, AccessLevelAll, false},

		// AccessLevelWrite comparisons
		{"Write is lower than All", AccessLevelWrite, AccessLevelAll, true},
		{"Write is not lower than Read", AccessLevelWrite, AccessLevelRead, false},
		{"Write is not lower than View", AccessLevelWrite, AccessLevelView, false},
		{"Write is not lower than None", AccessLevelWrite, AccessLevelNone, false},
		{"Write is not lower than Audit", AccessLevelWrite, AccessLevelAudit, false},
		{"Write is not lower than Write", AccessLevelWrite, AccessLevelWrite, false},

		// AccessLevelRead comparisons
		{"Read is lower than All", AccessLevelRead, AccessLevelAll, true},
		{"Read is lower than Write", AccessLevelRead, AccessLevelWrite, true},
		{"Read is not lower than View", AccessLevelRead, AccessLevelView, false},
		{"Read is not lower than None", AccessLevelRead, AccessLevelNone, false},
		{"Read is lower than Audit", AccessLevelRead, AccessLevelAudit, true},
		{"Read is not lower than Read", AccessLevelRead, AccessLevelRead, false},

		// AccessLevelView comparisons
		{"View is lower than All", AccessLevelView, AccessLevelAll, true},
		{"View is lower than Write", AccessLevelView, AccessLevelWrite, true},
		{"View is lower than Read", AccessLevelView, AccessLevelRead, true},
		{"View is lower than Audit", AccessLevelView, AccessLevelAudit, true},
		{"View is not lower than None", AccessLevelView, AccessLevelNone, false},
		{"View is not lower than View", AccessLevelView, AccessLevelView, false},

		// AccessLevelAudit comparisons
		{"Audit is lower than All", AccessLevelAudit, AccessLevelAll, true},
		{"Audit is lower than Write", AccessLevelAudit, AccessLevelWrite, true},
		{"Audit is not lower than Read", AccessLevelAudit, AccessLevelRead, false},
		{"Audit is not lower than View", AccessLevelAudit, AccessLevelView, false},
		{"Audit is not lower than None", AccessLevelAudit, AccessLevelNone, false},
		{"Audit is not lower than Audit", AccessLevelAudit, AccessLevelAudit, false},

		// AccessLevelNone comparisons
		{"None is lower than All", AccessLevelNone, AccessLevelAll, true},
		{"None is lower than Write", AccessLevelNone, AccessLevelWrite, true},
		{"None is lower than Read", AccessLevelNone, AccessLevelRead, true},
		{"None is lower than View", AccessLevelNone, AccessLevelView, true},
		{"None is lower than Audit", AccessLevelNone, AccessLevelAudit, true},
		{"None is not lower than None", AccessLevelNone, AccessLevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.IsLowerThan(tt.other)
			if result != tt.expected {
				t.Errorf("IsLowerThan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_IsEqual(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		other    AccessLevel
		expected bool
	}{
		{"All equals All", AccessLevelAll, AccessLevelAll, true},
		{"All not equals Write", AccessLevelAll, AccessLevelWrite, false},
		{"Write equals Write", AccessLevelWrite, AccessLevelWrite, true},
		{"Write not equals Read", AccessLevelWrite, AccessLevelRead, false},
		{"Read equals Read", AccessLevelRead, AccessLevelRead, true},
		{"Read not equals View", AccessLevelRead, AccessLevelView, false},
		{"View equals View", AccessLevelView, AccessLevelView, true},
		{"View not equals None", AccessLevelView, AccessLevelNone, false},
		{"None equals None", AccessLevelNone, AccessLevelNone, true},
		{"None not equals All", AccessLevelNone, AccessLevelAll, false},
		{"Audit equals Audit", AccessLevelAudit, AccessLevelAudit, true},
		{"Audit not equals Read", AccessLevelAudit, AccessLevelRead, false},
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

func TestAccessLevel_CanAccess(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		required AccessLevel
		expected bool
	}{
		// AccessLevelAll can access everything
		{"All can access All", AccessLevelAll, AccessLevelAll, true},
		{"All can access Write", AccessLevelAll, AccessLevelWrite, true},
		{"All can access Read", AccessLevelAll, AccessLevelRead, true},
		{"All can access View", AccessLevelAll, AccessLevelView, true},
		{"All can access None", AccessLevelAll, AccessLevelNone, true},
		{"All can access Audit", AccessLevelAll, AccessLevelAudit, true},

		// Same level can access itself
		{"Write can access Write", AccessLevelWrite, AccessLevelWrite, true},
		{"Read can access Read", AccessLevelRead, AccessLevelRead, true},
		{"View can access View", AccessLevelView, AccessLevelView, true},
		{"None can access None", AccessLevelNone, AccessLevelNone, true},
		{"Audit can access Audit", AccessLevelAudit, AccessLevelAudit, true},

		// Parent can access child
		{"Write can access Read", AccessLevelWrite, AccessLevelRead, true},
		{"Write can access View", AccessLevelWrite, AccessLevelView, true},
		{"Read can access View", AccessLevelRead, AccessLevelView, true},
		{"View can access None", AccessLevelView, AccessLevelNone, true},
		{"Audit can access Read", AccessLevelAudit, AccessLevelRead, true},
		{"Audit can access View", AccessLevelAudit, AccessLevelView, true},

		// Higher level can access lower level
		{"Write can access Read", AccessLevelWrite, AccessLevelRead, true},
		{"Read can access View", AccessLevelRead, AccessLevelView, true},
		{"Audit can access Read", AccessLevelAudit, AccessLevelRead, true},

		// Child cannot access parent
		{"Read cannot access Write", AccessLevelRead, AccessLevelWrite, false},
		{"View cannot access Read", AccessLevelView, AccessLevelRead, false},
		{"None cannot access View", AccessLevelNone, AccessLevelView, false},
		{"Read cannot access Audit", AccessLevelRead, AccessLevelAudit, false},

		// Lower level cannot access higher level
		{"View cannot access Read", AccessLevelView, AccessLevelRead, false},
		{"None cannot access View", AccessLevelNone, AccessLevelView, false},
		{"Read cannot access Audit", AccessLevelRead, AccessLevelAudit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.CanAccess(tt.required)
			if result != tt.expected {
				t.Errorf("CanAccess() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAccessLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		current  AccessLevel
		expected string
	}{
		{"All string", AccessLevelAll, "*"},
		{"Write string", AccessLevelWrite, "write"},
		{"Read string", AccessLevelRead, "read"},
		{"View string", AccessLevelView, "view"},
		{"None string", AccessLevelNone, "none"},
		{"Audit string", AccessLevelAudit, "audit"},
		{"Update string", AccessLevelUpdate, "update"},
		{"Create string", AccessLevelCreate, "create"},
		{"Delete string", AccessLevelDelete, "delete"},
		{"Approve string", AccessLevelApprove, "approve"},
		{"Reject string", AccessLevelReject, "reject"},
		{"Cancel string", AccessLevelCancel, "cancel"},
		{"Suspend string", AccessLevelSuspend, "suspend"},
		{"Resume string", AccessLevelResume, "resume"},
		{"Reset string", AccessLevelReset, "reset"},
		{"Lock string", AccessLevelLock, "lock"},
		{"Unlock string", AccessLevelUnlock, "unlock"},
		{"Revoke string", AccessLevelRevoke, "revoke"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test hierarchy consistency
func TestAccessLevel_HierarchyConsistency(t *testing.T) {
	levels := []AccessLevel{
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

	// Test that IsHigherThan and IsLowerThan are consistent
	for i, level1 := range levels {
		for j, level2 := range levels {
			if i < j {
				// level1 should be higher than level2
				if !level1.IsHigherThan(level2) {
					t.Errorf("%v should be higher than %v", level1, level2)
				}
				if !level2.IsLowerThan(level1) {
					t.Errorf("%v should be lower than %v", level2, level1)
				}
			} else if i > j {
				// level1 should be lower than level2
				if !level1.IsLowerThan(level2) {
					t.Errorf("%v should be lower than %v", level1, level2)
				}
				if !level2.IsHigherThan(level1) {
					t.Errorf("%v should be higher than %v", level2, level1)
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

// Test parent-child relationship consistency
func TestAccessLevel_ParentChildConsistency(t *testing.T) {
	levels := []AccessLevel{
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

	for _, level := range levels {
		// Test that children are actually children
		for _, otherLevel := range levels {
			if level.IsParentOf(otherLevel) {
				if !otherLevel.IsChildOf(level) {
					t.Errorf("%v should be a child of %v", otherLevel, level)
				}
			}
		}
	}
}

// Test edge cases and invalid inputs
func TestAccessLevel_EdgeCases(t *testing.T) {
	// Test with empty string
	emptyLevel := AccessLevel("")
	if emptyLevel.String() != "" {
		t.Errorf("Empty AccessLevel should return empty string")
	}

	// Test with unknown level
	unknownLevel := AccessLevel("unknown")
	if unknownLevel.IsParentOf(AccessLevelRead) {
		t.Errorf("Unknown level should not be parent of anything")
	}
	if unknownLevel.IsChildOf(AccessLevelAll) {
		t.Errorf("Unknown level should not be child of anything")
	}
	// Unknown levels are not in the hierarchy, so they should not be higher or lower than anything
	// The current implementation returns false for unknown levels, which is correct
	// Note: Unknown levels might still be able to access themselves if they match exactly
	// This is a design decision - we could change it if needed
}

// Test all constants are defined
func TestAccessLevel_Constants(t *testing.T) {
	expectedLevels := map[string]AccessLevel{
		"AccessLevelRead":    AccessLevelRead,
		"AccessLevelWrite":   AccessLevelWrite,
		"AccessLevelDelete":  AccessLevelDelete,
		"AccessLevelAll":     AccessLevelAll,
		"AccessLevelNone":    AccessLevelNone,
		"AccessLevelUpdate":  AccessLevelUpdate,
		"AccessLevelCreate":  AccessLevelCreate,
		"AccessLevelView":    AccessLevelView,
		"AccessLevelApprove": AccessLevelApprove,
		"AccessLevelReject":  AccessLevelReject,
		"AccessLevelCancel":  AccessLevelCancel,
		"AccessLevelSuspend": AccessLevelSuspend,
		"AccessLevelResume":  AccessLevelResume,
		"AccessLevelReset":   AccessLevelReset,
		"AccessLevelUnlock":  AccessLevelUnlock,
		"AccessLevelLock":    AccessLevelLock,
		"AccessLevelRevoke":  AccessLevelRevoke,
		"AccessLevelAudit":   AccessLevelAudit,
	}

	for name, level := range expectedLevels {
		if level == "" {
			t.Errorf("Constant %s is not defined", name)
		}
	}
}
