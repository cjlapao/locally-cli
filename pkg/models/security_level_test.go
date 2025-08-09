package models

import (
	"testing"
)

func TestSecurityLevel_IsParentOf(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		child    SecurityLevel
		expected bool
	}{
		// SuperUser tests
		{"SuperUser is parent of Admin", SecurityLevelSuperUser, SecurityLevelAdmin, true},
		{"SuperUser is parent of Manager", SecurityLevelSuperUser, SecurityLevelManager, true},
		{"SuperUser is parent of User", SecurityLevelSuperUser, SecurityLevelUser, true},
		{"SuperUser is parent of Auditor", SecurityLevelSuperUser, SecurityLevelAuditor, true},
		{"SuperUser is parent of Guest", SecurityLevelSuperUser, SecurityLevelGuest, true},
		{"SuperUser is parent of None", SecurityLevelSuperUser, SecurityLevelNone, true},
		{"SuperUser is not parent of SuperUser", SecurityLevelSuperUser, SecurityLevelSuperUser, false},

		// Admin tests
		{"Admin is parent of Manager", SecurityLevelAdmin, SecurityLevelManager, true},
		{"Admin is parent of User", SecurityLevelAdmin, SecurityLevelUser, true},
		{"Admin is parent of Auditor", SecurityLevelAdmin, SecurityLevelAuditor, true},
		{"Admin is parent of Guest", SecurityLevelAdmin, SecurityLevelGuest, true},
		{"Admin is parent of None", SecurityLevelAdmin, SecurityLevelNone, true},
		{"Admin is not parent of SuperUser", SecurityLevelAdmin, SecurityLevelSuperUser, false},
		{"Admin is not parent of Admin", SecurityLevelAdmin, SecurityLevelAdmin, false},

		// Manager tests
		{"Manager is parent of User", SecurityLevelManager, SecurityLevelUser, true},
		{"Manager is parent of Auditor", SecurityLevelManager, SecurityLevelAuditor, true},
		{"Manager is parent of Guest", SecurityLevelManager, SecurityLevelGuest, true},
		{"Manager is parent of None", SecurityLevelManager, SecurityLevelNone, true},
		{"Manager is not parent of SuperUser", SecurityLevelManager, SecurityLevelSuperUser, false},
		{"Manager is not parent of Admin", SecurityLevelManager, SecurityLevelAdmin, false},
		{"Manager is not parent of Manager", SecurityLevelManager, SecurityLevelManager, false},

		// User tests
		{"User is parent of Auditor", SecurityLevelUser, SecurityLevelAuditor, true},
		{"User is parent of Guest", SecurityLevelUser, SecurityLevelGuest, true},
		{"User is parent of None", SecurityLevelUser, SecurityLevelNone, true},
		{"User is not parent of SuperUser", SecurityLevelUser, SecurityLevelSuperUser, false},
		{"User is not parent of Admin", SecurityLevelUser, SecurityLevelAdmin, false},
		{"User is not parent of Manager", SecurityLevelUser, SecurityLevelManager, false},
		{"User is not parent of User", SecurityLevelUser, SecurityLevelUser, false},

		// Auditor tests
		{"Auditor is parent of Guest", SecurityLevelAuditor, SecurityLevelGuest, true},
		{"Auditor is parent of None", SecurityLevelAuditor, SecurityLevelNone, true},
		{"Auditor is not parent of SuperUser", SecurityLevelAuditor, SecurityLevelSuperUser, false},
		{"Auditor is not parent of Admin", SecurityLevelAuditor, SecurityLevelAdmin, false},
		{"Auditor is not parent of Manager", SecurityLevelAuditor, SecurityLevelManager, false},
		{"Auditor is not parent of User", SecurityLevelAuditor, SecurityLevelUser, false},
		{"Auditor is not parent of Auditor", SecurityLevelAuditor, SecurityLevelAuditor, false},

		// Guest tests
		{"Guest is parent of None", SecurityLevelGuest, SecurityLevelNone, true},
		{"Guest is not parent of SuperUser", SecurityLevelGuest, SecurityLevelSuperUser, false},
		{"Guest is not parent of Admin", SecurityLevelGuest, SecurityLevelAdmin, false},
		{"Guest is not parent of Manager", SecurityLevelGuest, SecurityLevelManager, false},
		{"Guest is not parent of User", SecurityLevelGuest, SecurityLevelUser, false},
		{"Guest is not parent of Auditor", SecurityLevelGuest, SecurityLevelAuditor, false},
		{"Guest is not parent of Guest", SecurityLevelGuest, SecurityLevelGuest, false},

		// None tests
		{"None is not parent of anyone", SecurityLevelNone, SecurityLevelSuperUser, false},
		{"None is not parent of Admin", SecurityLevelNone, SecurityLevelAdmin, false},
		{"None is not parent of Manager", SecurityLevelNone, SecurityLevelManager, false},
		{"None is not parent of User", SecurityLevelNone, SecurityLevelUser, false},
		{"None is not parent of Auditor", SecurityLevelNone, SecurityLevelAuditor, false},
		{"None is not parent of Guest", SecurityLevelNone, SecurityLevelGuest, false},
		{"None is not parent of None", SecurityLevelNone, SecurityLevelNone, false},
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

func TestSecurityLevel_IsChildOf(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		parent   SecurityLevel
		expected bool
	}{
		// SuperUser tests
		{"SuperUser is not child of anyone", SecurityLevelSuperUser, SecurityLevelSuperUser, false},
		{"SuperUser is not child of Admin", SecurityLevelSuperUser, SecurityLevelAdmin, false},
		{"SuperUser is not child of Manager", SecurityLevelSuperUser, SecurityLevelManager, false},
		{"SuperUser is not child of User", SecurityLevelSuperUser, SecurityLevelUser, false},
		{"SuperUser is not child of Auditor", SecurityLevelSuperUser, SecurityLevelAuditor, false},
		{"SuperUser is not child of Guest", SecurityLevelSuperUser, SecurityLevelGuest, false},
		{"SuperUser is not child of None", SecurityLevelSuperUser, SecurityLevelNone, false},

		// Admin tests
		{"Admin is child of SuperUser", SecurityLevelAdmin, SecurityLevelSuperUser, true},
		{"Admin is not child of Admin", SecurityLevelAdmin, SecurityLevelAdmin, false},
		{"Admin is not child of Manager", SecurityLevelAdmin, SecurityLevelManager, false},
		{"Admin is not child of User", SecurityLevelAdmin, SecurityLevelUser, false},
		{"Admin is not child of Auditor", SecurityLevelAdmin, SecurityLevelAuditor, false},
		{"Admin is not child of Guest", SecurityLevelAdmin, SecurityLevelGuest, false},
		{"Admin is not child of None", SecurityLevelAdmin, SecurityLevelNone, false},

		// Manager tests
		{"Manager is child of SuperUser", SecurityLevelManager, SecurityLevelSuperUser, true},
		{"Manager is child of Admin", SecurityLevelManager, SecurityLevelAdmin, true},
		{"Manager is not child of Manager", SecurityLevelManager, SecurityLevelManager, false},
		{"Manager is not child of User", SecurityLevelManager, SecurityLevelUser, false},
		{"Manager is not child of Auditor", SecurityLevelManager, SecurityLevelAuditor, false},
		{"Manager is not child of Guest", SecurityLevelManager, SecurityLevelGuest, false},
		{"Manager is not child of None", SecurityLevelManager, SecurityLevelNone, false},

		// User tests
		{"User is child of SuperUser", SecurityLevelUser, SecurityLevelSuperUser, true},
		{"User is child of Admin", SecurityLevelUser, SecurityLevelAdmin, true},
		{"User is child of Manager", SecurityLevelUser, SecurityLevelManager, true},
		{"User is not child of User", SecurityLevelUser, SecurityLevelUser, false},
		{"User is not child of Auditor", SecurityLevelUser, SecurityLevelAuditor, false},
		{"User is not child of Guest", SecurityLevelUser, SecurityLevelGuest, false},
		{"User is not child of None", SecurityLevelUser, SecurityLevelNone, false},

		// Auditor tests
		{"Auditor is child of SuperUser", SecurityLevelAuditor, SecurityLevelSuperUser, true},
		{"Auditor is child of Admin", SecurityLevelAuditor, SecurityLevelAdmin, true},
		{"Auditor is child of Manager", SecurityLevelAuditor, SecurityLevelManager, true},
		{"Auditor is child of User", SecurityLevelAuditor, SecurityLevelUser, true},
		{"Auditor is not child of Auditor", SecurityLevelAuditor, SecurityLevelAuditor, false},
		{"Auditor is not child of Guest", SecurityLevelAuditor, SecurityLevelGuest, false},
		{"Auditor is not child of None", SecurityLevelAuditor, SecurityLevelNone, false},

		// Guest tests
		{"Guest is child of SuperUser", SecurityLevelGuest, SecurityLevelSuperUser, true},
		{"Guest is child of Admin", SecurityLevelGuest, SecurityLevelAdmin, true},
		{"Guest is child of Manager", SecurityLevelGuest, SecurityLevelManager, true},
		{"Guest is child of User", SecurityLevelGuest, SecurityLevelUser, true},
		{"Guest is child of Auditor", SecurityLevelGuest, SecurityLevelAuditor, true},
		{"Guest is not child of Guest", SecurityLevelGuest, SecurityLevelGuest, false},
		{"Guest is not child of None", SecurityLevelGuest, SecurityLevelNone, false},

		// None tests
		{"None is child of SuperUser", SecurityLevelNone, SecurityLevelSuperUser, true},
		{"None is child of Admin", SecurityLevelNone, SecurityLevelAdmin, true},
		{"None is child of Manager", SecurityLevelNone, SecurityLevelManager, true},
		{"None is child of User", SecurityLevelNone, SecurityLevelUser, true},
		{"None is child of Auditor", SecurityLevelNone, SecurityLevelAuditor, true},
		{"None is child of Guest", SecurityLevelNone, SecurityLevelGuest, true},
		{"None is not child of None", SecurityLevelNone, SecurityLevelNone, false},
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

func TestSecurityLevel_GetParent(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		expected SecurityLevel
	}{
		{"SuperUser parent is SuperUser", SecurityLevelSuperUser, SecurityLevelSuperUser},
		{"Admin parent is SuperUser", SecurityLevelAdmin, SecurityLevelSuperUser},
		{"Manager parent is Admin", SecurityLevelManager, SecurityLevelAdmin},
		{"User parent is Manager", SecurityLevelUser, SecurityLevelManager},
		{"Auditor parent is User", SecurityLevelAuditor, SecurityLevelUser},
		{"Guest parent is Auditor", SecurityLevelGuest, SecurityLevelAuditor},
		{"None parent is Guest", SecurityLevelNone, SecurityLevelGuest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.GetParent()
			if result != tt.expected {
				t.Errorf("GetParent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSecurityLevel_GetChildren(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		expected []SecurityLevel
	}{
		{"SuperUser children", SecurityLevelSuperUser, []SecurityLevel{SecurityLevelAdmin}},
		{"Admin children", SecurityLevelAdmin, []SecurityLevel{SecurityLevelManager}},
		{"Manager children", SecurityLevelManager, []SecurityLevel{SecurityLevelUser}},
		{"User children", SecurityLevelUser, []SecurityLevel{SecurityLevelAuditor}},
		{"Auditor children", SecurityLevelAuditor, []SecurityLevel{SecurityLevelGuest}},
		{"Guest children", SecurityLevelGuest, []SecurityLevel{SecurityLevelNone}},
		{"None children", SecurityLevelNone, []SecurityLevel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.GetChildren()
			if len(result) != len(tt.expected) {
				t.Errorf("GetChildren() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("GetChildren()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestSecurityLevel_GetAllChildren(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		expected []SecurityLevel
	}{
		{"SuperUser all children", SecurityLevelSuperUser, []SecurityLevel{SecurityLevelAdmin, SecurityLevelManager, SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}},
		{"Admin all children", SecurityLevelAdmin, []SecurityLevel{SecurityLevelManager, SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}},
		{"Manager all children", SecurityLevelManager, []SecurityLevel{SecurityLevelUser, SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}},
		{"User all children", SecurityLevelUser, []SecurityLevel{SecurityLevelAuditor, SecurityLevelGuest, SecurityLevelNone}},
		{"Auditor all children", SecurityLevelAuditor, []SecurityLevel{SecurityLevelGuest, SecurityLevelNone}},
		{"Guest all children", SecurityLevelGuest, []SecurityLevel{SecurityLevelNone}},
		{"None all children", SecurityLevelNone, []SecurityLevel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.GetAllChildren()
			if len(result) != len(tt.expected) {
				t.Errorf("GetAllChildren() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("GetAllChildren()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestSecurityLevel_IsHigherThan(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		other    SecurityLevel
		expected bool
	}{
		// SuperUser comparisons
		{"SuperUser is higher than Admin", SecurityLevelSuperUser, SecurityLevelAdmin, true},
		{"SuperUser is higher than Manager", SecurityLevelSuperUser, SecurityLevelManager, true},
		{"SuperUser is higher than User", SecurityLevelSuperUser, SecurityLevelUser, true},
		{"SuperUser is higher than Auditor", SecurityLevelSuperUser, SecurityLevelAuditor, true},
		{"SuperUser is higher than Guest", SecurityLevelSuperUser, SecurityLevelGuest, true},
		{"SuperUser is higher than None", SecurityLevelSuperUser, SecurityLevelNone, true},
		{"SuperUser is not higher than SuperUser", SecurityLevelSuperUser, SecurityLevelSuperUser, false},

		// Admin comparisons
		{"Admin is not higher than SuperUser", SecurityLevelAdmin, SecurityLevelSuperUser, false},
		{"Admin is higher than Manager", SecurityLevelAdmin, SecurityLevelManager, true},
		{"Admin is higher than User", SecurityLevelAdmin, SecurityLevelUser, true},
		{"Admin is higher than Auditor", SecurityLevelAdmin, SecurityLevelAuditor, true},
		{"Admin is higher than Guest", SecurityLevelAdmin, SecurityLevelGuest, true},
		{"Admin is higher than None", SecurityLevelAdmin, SecurityLevelNone, true},
		{"Admin is not higher than Admin", SecurityLevelAdmin, SecurityLevelAdmin, false},

		// Manager comparisons
		{"Manager is not higher than SuperUser", SecurityLevelManager, SecurityLevelSuperUser, false},
		{"Manager is not higher than Admin", SecurityLevelManager, SecurityLevelAdmin, false},
		{"Manager is higher than User", SecurityLevelManager, SecurityLevelUser, true},
		{"Manager is higher than Auditor", SecurityLevelManager, SecurityLevelAuditor, true},
		{"Manager is higher than Guest", SecurityLevelManager, SecurityLevelGuest, true},
		{"Manager is higher than None", SecurityLevelManager, SecurityLevelNone, true},
		{"Manager is not higher than Manager", SecurityLevelManager, SecurityLevelManager, false},

		// User comparisons
		{"User is not higher than SuperUser", SecurityLevelUser, SecurityLevelSuperUser, false},
		{"User is not higher than Admin", SecurityLevelUser, SecurityLevelAdmin, false},
		{"User is not higher than Manager", SecurityLevelUser, SecurityLevelManager, false},
		{"User is higher than Auditor", SecurityLevelUser, SecurityLevelAuditor, true},
		{"User is higher than Guest", SecurityLevelUser, SecurityLevelGuest, true},
		{"User is higher than None", SecurityLevelUser, SecurityLevelNone, true},
		{"User is not higher than User", SecurityLevelUser, SecurityLevelUser, false},

		// Auditor comparisons
		{"Auditor is not higher than SuperUser", SecurityLevelAuditor, SecurityLevelSuperUser, false},
		{"Auditor is not higher than Admin", SecurityLevelAuditor, SecurityLevelAdmin, false},
		{"Auditor is not higher than Manager", SecurityLevelAuditor, SecurityLevelManager, false},
		{"Auditor is not higher than User", SecurityLevelAuditor, SecurityLevelUser, false},
		{"Auditor is higher than Guest", SecurityLevelAuditor, SecurityLevelGuest, true},
		{"Auditor is higher than None", SecurityLevelAuditor, SecurityLevelNone, true},
		{"Auditor is not higher than Auditor", SecurityLevelAuditor, SecurityLevelAuditor, false},

		// Guest comparisons
		{"Guest is not higher than SuperUser", SecurityLevelGuest, SecurityLevelSuperUser, false},
		{"Guest is not higher than Admin", SecurityLevelGuest, SecurityLevelAdmin, false},
		{"Guest is not higher than Manager", SecurityLevelGuest, SecurityLevelManager, false},
		{"Guest is not higher than User", SecurityLevelGuest, SecurityLevelUser, false},
		{"Guest is not higher than Auditor", SecurityLevelGuest, SecurityLevelAuditor, false},
		{"Guest is higher than None", SecurityLevelGuest, SecurityLevelNone, true},
		{"Guest is not higher than Guest", SecurityLevelGuest, SecurityLevelGuest, false},

		// None comparisons
		{"None is not higher than SuperUser", SecurityLevelNone, SecurityLevelSuperUser, false},
		{"None is not higher than Admin", SecurityLevelNone, SecurityLevelAdmin, false},
		{"None is not higher than Manager", SecurityLevelNone, SecurityLevelManager, false},
		{"None is not higher than User", SecurityLevelNone, SecurityLevelUser, false},
		{"None is not higher than Auditor", SecurityLevelNone, SecurityLevelAuditor, false},
		{"None is not higher than Guest", SecurityLevelNone, SecurityLevelGuest, false},
		{"None is not higher than None", SecurityLevelNone, SecurityLevelNone, false},
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

func TestSecurityLevel_IsLowerThan(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		other    SecurityLevel
		expected bool
	}{
		// SuperUser comparisons
		{"SuperUser is not lower than Admin", SecurityLevelSuperUser, SecurityLevelAdmin, false},
		{"SuperUser is not lower than Manager", SecurityLevelSuperUser, SecurityLevelManager, false},
		{"SuperUser is not lower than User", SecurityLevelSuperUser, SecurityLevelUser, false},
		{"SuperUser is not lower than Auditor", SecurityLevelSuperUser, SecurityLevelAuditor, false},
		{"SuperUser is not lower than Guest", SecurityLevelSuperUser, SecurityLevelGuest, false},
		{"SuperUser is not lower than None", SecurityLevelSuperUser, SecurityLevelNone, false},
		{"SuperUser is not lower than SuperUser", SecurityLevelSuperUser, SecurityLevelSuperUser, false},

		// Admin comparisons
		{"Admin is lower than SuperUser", SecurityLevelAdmin, SecurityLevelSuperUser, true},
		{"Admin is not lower than Manager", SecurityLevelAdmin, SecurityLevelManager, false},
		{"Admin is not lower than User", SecurityLevelAdmin, SecurityLevelUser, false},
		{"Admin is not lower than Auditor", SecurityLevelAdmin, SecurityLevelAuditor, false},
		{"Admin is not lower than Guest", SecurityLevelAdmin, SecurityLevelGuest, false},
		{"Admin is not lower than None", SecurityLevelAdmin, SecurityLevelNone, false},
		{"Admin is not lower than Admin", SecurityLevelAdmin, SecurityLevelAdmin, false},

		// Manager comparisons
		{"Manager is lower than SuperUser", SecurityLevelManager, SecurityLevelSuperUser, true},
		{"Manager is lower than Admin", SecurityLevelManager, SecurityLevelAdmin, true},
		{"Manager is not lower than User", SecurityLevelManager, SecurityLevelUser, false},
		{"Manager is not lower than Auditor", SecurityLevelManager, SecurityLevelAuditor, false},
		{"Manager is not lower than Guest", SecurityLevelManager, SecurityLevelGuest, false},
		{"Manager is not lower than None", SecurityLevelManager, SecurityLevelNone, false},
		{"Manager is not lower than Manager", SecurityLevelManager, SecurityLevelManager, false},

		// User comparisons
		{"User is lower than SuperUser", SecurityLevelUser, SecurityLevelSuperUser, true},
		{"User is lower than Admin", SecurityLevelUser, SecurityLevelAdmin, true},
		{"User is lower than Manager", SecurityLevelUser, SecurityLevelManager, true},
		{"User is not lower than Auditor", SecurityLevelUser, SecurityLevelAuditor, false},
		{"User is not lower than Guest", SecurityLevelUser, SecurityLevelGuest, false},
		{"User is not lower than None", SecurityLevelUser, SecurityLevelNone, false},
		{"User is not lower than User", SecurityLevelUser, SecurityLevelUser, false},

		// Auditor comparisons
		{"Auditor is lower than SuperUser", SecurityLevelAuditor, SecurityLevelSuperUser, true},
		{"Auditor is lower than Admin", SecurityLevelAuditor, SecurityLevelAdmin, true},
		{"Auditor is lower than Manager", SecurityLevelAuditor, SecurityLevelManager, true},
		{"Auditor is lower than User", SecurityLevelAuditor, SecurityLevelUser, true},
		{"Auditor is not lower than Guest", SecurityLevelAuditor, SecurityLevelGuest, false},
		{"Auditor is not lower than None", SecurityLevelAuditor, SecurityLevelNone, false},
		{"Auditor is not lower than Auditor", SecurityLevelAuditor, SecurityLevelAuditor, false},

		// Guest comparisons
		{"Guest is lower than SuperUser", SecurityLevelGuest, SecurityLevelSuperUser, true},
		{"Guest is lower than Admin", SecurityLevelGuest, SecurityLevelAdmin, true},
		{"Guest is lower than Manager", SecurityLevelGuest, SecurityLevelManager, true},
		{"Guest is lower than User", SecurityLevelGuest, SecurityLevelUser, true},
		{"Guest is lower than Auditor", SecurityLevelGuest, SecurityLevelAuditor, true},
		{"Guest is not lower than None", SecurityLevelGuest, SecurityLevelNone, false},
		{"Guest is not lower than Guest", SecurityLevelGuest, SecurityLevelGuest, false},

		// None comparisons
		{"None is lower than SuperUser", SecurityLevelNone, SecurityLevelSuperUser, true},
		{"None is lower than Admin", SecurityLevelNone, SecurityLevelAdmin, true},
		{"None is lower than Manager", SecurityLevelNone, SecurityLevelManager, true},
		{"None is lower than User", SecurityLevelNone, SecurityLevelUser, true},
		{"None is lower than Auditor", SecurityLevelNone, SecurityLevelAuditor, true},
		{"None is lower than Guest", SecurityLevelNone, SecurityLevelGuest, true},
		{"None is not lower than None", SecurityLevelNone, SecurityLevelNone, false},
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

func TestSecurityLevel_IsEqual(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		other    SecurityLevel
		expected bool
	}{
		{"SuperUser equals SuperUser", SecurityLevelSuperUser, SecurityLevelSuperUser, true},
		{"SuperUser not equals Admin", SecurityLevelSuperUser, SecurityLevelAdmin, false},
		{"Admin equals Admin", SecurityLevelAdmin, SecurityLevelAdmin, true},
		{"Admin not equals Manager", SecurityLevelAdmin, SecurityLevelManager, false},
		{"Manager equals Manager", SecurityLevelManager, SecurityLevelManager, true},
		{"Manager not equals User", SecurityLevelManager, SecurityLevelUser, false},
		{"User equals User", SecurityLevelUser, SecurityLevelUser, true},
		{"User not equals Auditor", SecurityLevelUser, SecurityLevelAuditor, false},
		{"Auditor equals Auditor", SecurityLevelAuditor, SecurityLevelAuditor, true},
		{"Auditor not equals Guest", SecurityLevelAuditor, SecurityLevelGuest, false},
		{"Guest equals Guest", SecurityLevelGuest, SecurityLevelGuest, true},
		{"Guest not equals None", SecurityLevelGuest, SecurityLevelNone, false},
		{"None equals None", SecurityLevelNone, SecurityLevelNone, true},
		{"None not equals SuperUser", SecurityLevelNone, SecurityLevelSuperUser, false},
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

func TestSecurityLevel_GetLevel(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		expected int
	}{
		{"SuperUser level", SecurityLevelSuperUser, 0},
		{"Admin level", SecurityLevelAdmin, 1},
		{"Manager level", SecurityLevelManager, 2},
		{"User level", SecurityLevelUser, 3},
		{"Auditor level", SecurityLevelAuditor, 4},
		{"Guest level", SecurityLevelGuest, 5},
		{"None level", SecurityLevelNone, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.current.GetLevel()
			if result != tt.expected {
				t.Errorf("GetLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSecurityLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		current  SecurityLevel
		expected string
	}{
		{"SuperUser string", SecurityLevelSuperUser, "superuser"},
		{"Admin string", SecurityLevelAdmin, "admin"},
		{"Manager string", SecurityLevelManager, "manager"},
		{"User string", SecurityLevelUser, "user"},
		{"Auditor string", SecurityLevelAuditor, "auditor"},
		{"Guest string", SecurityLevelGuest, "guest"},
		{"None string", SecurityLevelNone, "none"},
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
func TestSecurityLevel_HierarchyConsistency(t *testing.T) {
	levels := []SecurityLevel{
		SecurityLevelSuperUser,
		SecurityLevelAdmin,
		SecurityLevelManager,
		SecurityLevelUser,
		SecurityLevelAuditor,
		SecurityLevelGuest,
		SecurityLevelNone,
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
func TestSecurityLevel_ParentChildConsistency(t *testing.T) {
	levels := []SecurityLevel{
		SecurityLevelSuperUser,
		SecurityLevelAdmin,
		SecurityLevelManager,
		SecurityLevelUser,
		SecurityLevelAuditor,
		SecurityLevelGuest,
		SecurityLevelNone,
	}

	for _, level := range levels {
		parent := level.GetParent()

		// If level has a parent (not SuperUser), it should be a child of that parent
		if level != SecurityLevelSuperUser {
			if !level.IsChildOf(parent) {
				t.Errorf("%v should be a child of its parent %v", level, parent)
			}
		}

		// Test that children are actually children
		children := level.GetChildren()
		for _, child := range children {
			if !child.IsChildOf(level) {
				t.Errorf("%v should be a child of %v", child, level)
			}
			if !level.IsParentOf(child) {
				t.Errorf("%v should be a parent of %v", level, child)
			}
		}
	}
}
