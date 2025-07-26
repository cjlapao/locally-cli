package auth

import (
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/models"
)

var DefaultRoles = []models.Role{
	{
		Name:        "Super User",
		Slug:        config.SuperUserRole,
		Description: "Super user role",
		IsAdmin:     true,
		IsSuperUser: true,
	},
	{
		Name:        "User",
		Slug:        "user",
		Description: "User role",
		IsAdmin:     false,
		IsSuperUser: false,
	},
	{
		Name:        "Admin",
		Slug:        "admin",
		Description: "Admin role",
		IsAdmin:     true,
		IsSuperUser: false,
	},
	{
		Name:        "Guest",
		Slug:        "guest",
		Description: "Guest role",
		IsAdmin:     false,
		IsSuperUser: false,
	},
}

var DefaultUserClaims = []models.Claim{
	{
		Service: "locally",
		Module:  "*",
		Action:  models.ClaimActionAll,
	},
}

var DefaultRootClaims = []models.Claim{
	{
		Service: "locally",
		Module:  "*",
		Action:  models.ClaimActionAll,
	},
}
