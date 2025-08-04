// Package defaults provides the default roles for the role service.
package defaults

import (
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/models"
)

var DefaultRoles = []models.Role{
	{
		ID:            config.RoleSuperUserID,
		Name:          "Super User",
		Slug:          config.SuperUserRole,
		Description:   "Super user role",
		SecurityLevel: models.SecurityLevelSuperUser,
	},
	{
		ID:            config.RoleAdminUserID,
		Name:          "Admin",
		Slug:          "admin",
		Description:   "Admin role",
		SecurityLevel: models.SecurityLevelAdmin,
	},
	{
		ID:            config.RoleManagerUserID,
		Name:          "Manager",
		Slug:          "manager",
		Description:   "Manager role",
		SecurityLevel: models.SecurityLevelManager,
	},
	{
		ID:            config.RoleUserID,
		Name:          "User",
		Slug:          "user",
		Description:   "User role",
		SecurityLevel: models.SecurityLevelUser,
	},
	{
		ID:            config.RoleGuestUserID,
		Name:          "Guest",
		Slug:          "guest",
		Description:   "Guest role",
		SecurityLevel: models.SecurityLevelGuest,
	},
	{
		ID:            config.RoleNoneUserID,
		Name:          "None",
		Slug:          "none",
		Description:   "No access role",
		SecurityLevel: models.SecurityLevelNone,
	},
}
