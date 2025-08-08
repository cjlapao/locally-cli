// Package interfaces provides the system service interface.
package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/models"
)

// SystemServiceInterface defines the interface for the system service
type SystemServiceInterface interface {
	GetName() string
	AddService(service *models.ServiceDefinition)
	GetService(name string) (*models.ServiceDefinition, bool)
	ValidateClaim(claim *models.Claim) error
	ParseClaim(slug string) (*models.Claim, error)
	CreateDefaultClaim(service, module string, securityLevel models.SecurityLevel) (*models.Claim, error)
	GetAllowedActions(service, module string) ([]models.AccessLevel, error)
	GenerateSystemClaims() []*models.Claim
	GenerateDefaultClaimsForSecurityLevel(securityLevel models.SecurityLevel) []*models.Claim
	GenerateClaimsForSecurityLevels() map[models.SecurityLevel][]*models.Claim
	GenerateDefaultRoles() []*models.Role
	GetRoleBySecurityLevel(securityLevel models.SecurityLevel) (*models.Role, error)
	GetSystemSummary() map[string]interface{}
	String() string
	GetRoleByName(name string) *models.Role
	GetServiceModuleByName(service, module string) *models.ModuleDefinition
	LogSummary(ctx *appctx.AppContext)
}
