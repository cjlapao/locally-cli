// Package interfaces provides the system service interface.
package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
)

// SystemServiceInterface defines the interface for the system service
type SystemServiceInterface interface {
	GetName() string
	AddService(service *models.ServiceDefinition)
	GetService(name string) (*models.ServiceDefinition, bool)
	ValidateClaim(claim *models.Claim) *diagnostics.Diagnostics
	ParseClaim(slug string) (*models.Claim, *diagnostics.Diagnostics)
	CreateDefaultClaim(service, module string, securityLevel models.SecurityLevel) (*models.Claim, *diagnostics.Diagnostics)
	GetAllowedActions(service, module string) ([]models.AccessLevel, *diagnostics.Diagnostics)
	GenerateSystemClaims() []*models.Claim
	GenerateDefaultClaimsForSecurityLevel(securityLevel models.SecurityLevel) []*models.Claim
	GenerateClaimsForSecurityLevels() map[models.SecurityLevel][]*models.Claim
	GenerateDefaultRoles() []*models.Role
	GetRoleBySecurityLevel(securityLevel models.SecurityLevel) (*models.Role, *diagnostics.Diagnostics)
	GetSystemSummary() map[string]interface{}
	String() string
	GetRoleByName(name string) *models.Role
	GetServiceModuleByName(service, module string) *models.ModuleDefinition
	LogSummary(ctx *appctx.AppContext)
}
