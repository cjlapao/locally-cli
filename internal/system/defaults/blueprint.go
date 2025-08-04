package defaults

import "github.com/cjlapao/locally-cli/pkg/models"

// SystemBlueprint is the blueprint for the system services
var SystemBlueprint = []*models.ServiceDefinition{
	{
		Name:        "tenant",
		Description: "Tenant service",
		Modules: map[string]*models.ModuleDefinition{
			"api": {
				Name:        "api",
				Description: "API module",
				Actions: []models.AccessLevel{
					models.AccessLevelRead,
					models.AccessLevelWrite,
					models.AccessLevelDelete,
				},
			},
			"ui": {
				Name:        "ui",
				Description: "UI module",
				Actions:     []models.AccessLevel{models.AccessLevelView},
			},
		},
	},
	{
		Name:        "user",
		Description: "User service",
		Modules: map[string]*models.ModuleDefinition{
			"api": {
				Name:        "api",
				Description: "API module",
				Actions: []models.AccessLevel{
					models.AccessLevelRead,
					models.AccessLevelWrite,
					models.AccessLevelDelete,
				},
			},
			"ui": {
				Name:        "ui",
				Description: "UI module",
				Actions:     []models.AccessLevel{models.AccessLevelView},
			},
		},
	},
	{
		Name:        "role",
		Description: "Role service",
		Modules: map[string]*models.ModuleDefinition{
			"api": {
				Name:        "api",
				Description: "API module",
				Actions: []models.AccessLevel{
					models.AccessLevelRead,
					models.AccessLevelWrite,
					models.AccessLevelDelete,
				},
			},
			"ui": {
				Name:        "ui",
				Description: "UI module",
				Actions:     []models.AccessLevel{models.AccessLevelView},
			},
		},
	},
	{
		Name:        "claim",
		Description: "Claim service",
		Modules: map[string]*models.ModuleDefinition{
			"api": {
				Name:        "api",
				Description: "API module",
				Actions: []models.AccessLevel{
					models.AccessLevelRead,
					models.AccessLevelWrite,
					models.AccessLevelDelete,
				},
			},
			"ui": {
				Name:        "ui",
				Description: "UI module",
				Actions:     []models.AccessLevel{models.AccessLevelView},
			},
		},
	},
	{
		Name:        "certificate",
		Description: "Certificate service",
		Modules: map[string]*models.ModuleDefinition{
			"api": {
				Name:        "api",
				Description: "API module",
				Actions: []models.AccessLevel{
					models.AccessLevelRead,
					models.AccessLevelWrite,
					models.AccessLevelDelete,
					models.AccessLevelApprove,
					models.AccessLevelReject,
				},
			},
		},
	},
}
