// Package service provides a service for managing the system
package service

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/system/defaults"
	"github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/pkg/models"
)

var (
	instance *SystemService
	once     sync.Once
)

// SystemService represents the system service singleton
type SystemService struct {
	services map[string]*models.ServiceDefinition
}

// GetInstance returns the singleton instance of the system service
func GetInstance() interfaces.SystemServiceInterface {
	if instance == nil {
		panic("SystemService instance is nil")
	}
	return instance
}

// Initialize initializes the system service singleton
func Initialize() interfaces.SystemServiceInterface {
	once.Do(func() {
		service := &SystemService{
			services: make(map[string]*models.ServiceDefinition),
		}

		for _, definition := range defaults.SystemBlueprint {
			service.AddService(definition)
		}

		instance = service
	})

	return instance
}

// ResetForTesting resets the singleton instance for testing purposes
// This should only be used in tests
func ResetForTesting() {
	instance = nil
	once = sync.Once{}
}

// GetName returns the name of the service
func (s *SystemService) GetName() string {
	return "system"
}

// AddService adds a service to the system
func (s *SystemService) AddService(service *models.ServiceDefinition) {
	s.services[service.Name] = service
}

// GetService returns a service definition by name
func (s *SystemService) GetService(name string) (*models.ServiceDefinition, bool) {
	service, exists := s.services[name]
	return service, exists
}

// ValidateClaim validates if a claim is valid according to the system definition
func (s *SystemService) ValidateClaim(claim *models.Claim) error {
	// Check if service exists
	service, exists := s.services[claim.Service]
	if !exists {
		return fmt.Errorf("service '%s' does not exist", claim.Service)
	}

	// Check if module exists
	module, exists := service.Modules[claim.Module]
	if !exists {
		return fmt.Errorf("module '%s' does not exist in service '%s'", claim.Module, claim.Service)
	}

	// Check if action is allowed for this module
	actionAllowed := false
	for _, allowedAction := range module.Actions {
		if allowedAction == claim.Action {
			actionAllowed = true
			break
		}
	}

	if !actionAllowed {
		return fmt.Errorf("action '%s' is not allowed for module '%s' in service '%s'", claim.Action, claim.Module, claim.Service)
	}

	return nil
}

// ParseClaim parses a claim from a slug and validates it against the system
func (s *SystemService) ParseClaim(slug string) (*models.Claim, error) {
	claim, err := models.ParseClaim(slug)
	if err != nil {
		return nil, err
	}

	// Validate the claim against the system
	if err := s.ValidateClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}

// GetDefaultAccessLevel returns the default access level for a security level
func GetDefaultAccessLevel(securityLevel models.SecurityLevel) models.AccessLevel {
	switch securityLevel {
	case models.SecurityLevelSuperUser:
		return models.AccessLevelAll
	case models.SecurityLevelAdmin:
		return models.AccessLevelAll
	case models.SecurityLevelManager:
		return models.AccessLevelWrite
	case models.SecurityLevelUser:
		return models.AccessLevelRead
	case models.SecurityLevelGuest:
		return models.AccessLevelView
	case models.SecurityLevelNone:
		return models.AccessLevelNone
	default:
		return models.AccessLevelNone
	}
}

// CreateDefaultClaim creates a default claim for a service and module based on security level
func (s *SystemService) CreateDefaultClaim(service, module string, securityLevel models.SecurityLevel) (*models.Claim, error) {
	// Validate service and module exist
	if err := s.ValidateClaim(&models.Claim{
		Service: service,
		Module:  module,
		Action:  models.AccessLevelRead, // Use read as a test action
	}); err != nil {
		return nil, err
	}

	// Get default access level for the security level
	defaultAction := GetDefaultAccessLevel(securityLevel)

	// Create the claim
	claim := &models.Claim{
		Service: service,
		Module:  module,
		Action:  defaultAction,
		Slug:    fmt.Sprintf("%s::%s::%s", service, module, defaultAction),
	}

	return claim, nil
}

// GetAllowedActions returns all allowed actions for a service and module
func (s *SystemService) GetAllowedActions(service, module string) ([]models.AccessLevel, error) {
	serviceDef, exists := s.services[service]
	if !exists {
		return nil, fmt.Errorf("service '%s' does not exist", service)
	}

	moduleDef, exists := serviceDef.Modules[module]
	if !exists {
		return nil, fmt.Errorf("module '%s' does not exist in service '%s'", module, service)
	}

	return moduleDef.Actions, nil
}

// GetSecurityLevelForAction determines the appropriate security level for a given action
func GetSecurityLevelForAction(action models.AccessLevel) models.SecurityLevel {
	switch action {
	case models.AccessLevelAll:
		return models.SecurityLevelSuperUser
	case models.AccessLevelDelete, models.AccessLevelApprove, models.AccessLevelReject, models.AccessLevelSuspend, models.AccessLevelResume, models.AccessLevelReset, models.AccessLevelUnlock, models.AccessLevelLock:
		return models.SecurityLevelAdmin
	case models.AccessLevelWrite, models.AccessLevelUpdate, models.AccessLevelCreate:
		return models.SecurityLevelManager
	case models.AccessLevelRead, models.AccessLevelView:
		return models.SecurityLevelUser
	case models.AccessLevelNone:
		return models.SecurityLevelNone
	default:
		return models.SecurityLevelUser // Default to user level for unknown actions
	}
}

// GenerateSystemClaims generates all necessary claims for the entire system
// This includes claims for all services, modules, and actions defined in the system
func (s *SystemService) GenerateSystemClaims() []*models.Claim {
	var claims []*models.Claim

	// Add the wildcard claim for superuser access
	wildcardClaim := &models.Claim{
		Service:       "*",
		Module:        "*",
		Action:        models.AccessLevelAll,
		Slug:          "*::*::*",
		SecurityLevel: models.SecurityLevelSuperUser,
	}
	claims = append(claims, wildcardClaim)

	for serviceName, service := range s.services {
		for moduleName, module := range service.Modules {
			for _, action := range module.Actions {
				// Automatically determine the security level based on the action
				securityLevel := GetSecurityLevelForAction(action)

				claim := &models.Claim{
					Service:       serviceName,
					Module:        moduleName,
					Action:        action,
					Slug:          fmt.Sprintf("%s::%s::%s", serviceName, moduleName, action),
					SecurityLevel: securityLevel,
				}
				claims = append(claims, claim)
			}
		}
	}

	return claims
}

// GenerateDefaultClaimsForSecurityLevel generates default claims for a specific security level
// across all services and modules in the system
func (s *SystemService) GenerateDefaultClaimsForSecurityLevel(securityLevel models.SecurityLevel) []*models.Claim {
	// Special case for SuperUser - return single wildcard claim
	if securityLevel == models.SecurityLevelSuperUser {
		wildcardClaim := &models.Claim{
			Service:       "*",
			Module:        "*",
			Action:        models.AccessLevelAll,
			Slug:          "*::*::*",
			SecurityLevel: securityLevel,
		}
		return []*models.Claim{wildcardClaim}
	}

	var claims []*models.Claim
	defaultAction := GetDefaultAccessLevel(securityLevel)

	for serviceName, service := range s.services {
		for moduleName, module := range service.Modules {
			// Check if the default action is allowed for this module
			actionAllowed := false
			for _, allowedAction := range module.Actions {
				if allowedAction == defaultAction {
					actionAllowed = true
					break
				}
			}

			// If the default action is not allowed, find the highest allowed action
			selectedAction := defaultAction
			if !actionAllowed {
				highestAction := models.AccessLevelNone
				for _, allowedAction := range module.Actions {
					if allowedAction.IsHigherThan(highestAction) {
						highestAction = allowedAction
					}
				}
				selectedAction = highestAction
			}

			claim := &models.Claim{
				Service:       serviceName,
				Module:        moduleName,
				Action:        selectedAction,
				Slug:          fmt.Sprintf("%s::%s::%s", serviceName, moduleName, selectedAction),
				SecurityLevel: securityLevel,
			}
			claims = append(claims, claim)
		}
	}

	return claims
}

// GenerateClaimsForSecurityLevels generates default claims for all security levels
// Returns a map of security level to claims
func (s *SystemService) GenerateClaimsForSecurityLevels() map[models.SecurityLevel][]*models.Claim {
	securityLevels := []models.SecurityLevel{
		models.SecurityLevelSuperUser,
		models.SecurityLevelAdmin,
		models.SecurityLevelManager,
		models.SecurityLevelUser,
		models.SecurityLevelAuditor,
		models.SecurityLevelGuest,
		models.SecurityLevelNone,
	}

	claimsMap := make(map[models.SecurityLevel][]*models.Claim)
	for _, level := range securityLevels {
		claimsMap[level] = s.GenerateDefaultClaimsForSecurityLevel(level)
	}

	return claimsMap
}

// GenerateDefaultRoles generates default roles for the system
func (s *SystemService) GenerateDefaultRoles() []*models.Role {
	result := []*models.Role{}
	for _, role := range defaults.DefaultRoles {
		result = append(result, &role)
	}
	return result
}

// GetSystemSummary returns a summary of the system including total services, modules, and actions
func (s *SystemService) GetSystemSummary() map[string]interface{} {
	totalServices := len(s.services)
	totalModules := 0
	totalActions := 0
	serviceDetails := make(map[string]interface{})

	for serviceName, service := range s.services {
		moduleCount := len(service.Modules)
		totalModules += moduleCount

		moduleDetails := make(map[string]interface{})
		for moduleName, module := range service.Modules {
			actionCount := len(module.Actions)
			totalActions += actionCount
			moduleDetails[moduleName] = map[string]interface{}{
				"description":  module.Description,
				"actions":      module.Actions,
				"action_count": actionCount,
			}
		}

		serviceDetails[serviceName] = map[string]interface{}{
			"description":  service.Description,
			"modules":      moduleDetails,
			"module_count": moduleCount,
		}
	}

	return map[string]interface{}{
		"total_services": totalServices,
		"total_modules":  totalModules,
		"total_actions":  totalActions,
		"services":       serviceDetails,
	}
}

// String returns a string representation of the system
func (s *SystemService) String() string {
	var sb strings.Builder
	sb.WriteString("System Definition:\n")

	for serviceName, service := range s.services {
		sb.WriteString(fmt.Sprintf("  Service: %s (%s)\n", serviceName, service.Description))
		for moduleName, module := range service.Modules {
			sb.WriteString(fmt.Sprintf("    Module: %s (%s)\n", moduleName, module.Description))
			sb.WriteString(fmt.Sprintf("      Actions: %v\n", module.Actions))
		}
	}

	return sb.String()
}

func (s *SystemService) LogSummary(ctx *appctx.AppContext) {
	summary := s.GetSystemSummary()
	for key, value := range summary {
		if key == "services" {
			for serviceName, service := range value.(map[string]interface{}) {
				serviceDescription := service.(map[string]interface{})["description"].(string)
				moduleCount := service.(map[string]interface{})["module_count"].(int)
				modules := service.(map[string]interface{})["modules"].(map[string]interface{})
				ctx.Log().WithField("service", serviceName).
					WithField("service_description", serviceDescription).
					WithField("module_count", moduleCount).
					Info("System Service registered")

				for moduleName, module := range modules {
					moduleActions := module.(map[string]interface{})["actions"].([]models.AccessLevel)
					actions := []string{}
					for _, action := range moduleActions {
						actions = append(actions, action.String())
					}

					ctx.Log().WithField("service", serviceName).
						WithField("module", moduleName).
						WithField("module_description", module.(map[string]interface{})["description"].(string)).
						WithField("actions", fmt.Sprintf("%v", actions)).
						Info("System Service module registered")
				}
			}
		}
	}
}

func (s *SystemService) GetRoleBySecurityLevel(securityLevel models.SecurityLevel) (*models.Role, error) {
	for _, role := range defaults.DefaultRoles {
		if role.SecurityLevel == securityLevel {
			return &role, nil
		}
	}
	return nil, fmt.Errorf("role with security level %s not found", securityLevel)
}

func (s *SystemService) GetRoleByName(name string) *models.Role {
	for _, role := range defaults.DefaultRoles {
		if strings.EqualFold(role.Name, name) {
			return &role
		}
	}
	return nil
}

func (s *SystemService) GetServiceModuleByName(service, module string) *models.ModuleDefinition {
	for _, systemService := range s.services {
		if strings.EqualFold(systemService.Name, service) {
			for _, systemModule := range systemService.Modules {
				if strings.EqualFold(systemModule.Name, module) {
					systemModule.System = systemService.Name
					return systemModule
				}
			}
		}
	}
	return nil
}
