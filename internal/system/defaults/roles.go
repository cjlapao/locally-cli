// Package defaults provides the default roles for the role service.
package defaults

import (
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

var DefaultRoles = []models.Role{
	{
		ID:            config.RoleSuperUserID,
		Name:          "Super User",
		Slug:          utils.Slugify("Super User"),
		Description:   "Super user role",
		SecurityLevel: models.SecurityLevelSuperUser,
	},
	{
		ID:            config.RoleAdminUserID,
		Name:          "Admin",
		Slug:          utils.Slugify("Admin"),
		Description:   "Admin role",
		SecurityLevel: models.SecurityLevelAdmin,
	},
	{
		ID:            config.RoleManagerUserID,
		Name:          "Manager",
		Slug:          utils.Slugify("Manager"),
		Description:   "Manager role",
		SecurityLevel: models.SecurityLevelManager,
	},
	{
		ID:            config.RoleUserID,
		Name:          "User",
		Slug:          utils.Slugify("User"),
		Description:   "User role",
		SecurityLevel: models.SecurityLevelUser,
	},
	{
		ID:            config.RoleGuestUserID,
		Name:          "Guest",
		Slug:          utils.Slugify("Guest"),
		Description:   "Guest role",
		SecurityLevel: models.SecurityLevelGuest,
	},
	{
		ID:            config.RoleNoneUserID,
		Name:          "None",
		Slug:          utils.Slugify("None"),
		Description:   "No access role",
		SecurityLevel: models.SecurityLevelNone,
	},
}
