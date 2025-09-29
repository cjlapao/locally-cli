// Package service provides a service for managing the environment
package service

import (
	"context"
	"sync"
	"time"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/internal/vaults/configvault"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
)

var (
	globalEnvironmentService *EnvironmentService
	environmentServiceOnce   sync.Once
	environmentServiceMutex  sync.Mutex
)

type EnvironmentService struct {
	environments     []interfaces.EnvironmentInterface
	environmentStore stores.EnvironmentDataStoreInterface
	functions        map[string]interfaces.EnvironmentVariableFunction
	variables        map[string]map[string]interface{}
	isSynced         bool
	mu               sync.RWMutex
	once             sync.Once
}

func Initialize(environmentStore stores.EnvironmentDataStoreInterface) interfaces.EnvironmentServiceInterface {
	environmentServiceMutex.Lock()
	defer environmentServiceMutex.Unlock()

	environmentServiceOnce.Do(func() {
		ctx := appctx.NewContext(context.Background())
		globalEnvironmentService = new(environmentStore)
		// creating the global environment, this will be the default environment and will not be persisted
		globalEnvironmentModel := models.Environment{
			BaseModelWithTenant: models.BaseModelWithTenant{
				ID:        "global",
				Slug:      "global",
				CreatedBy: "system",
				UpdatedBy: "system",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				TenantID:  "global",
			},
			Name:      "Global",
			Type:      types.EnvironmentTypeGlobal,
			ProjectID: "",
			Enabled:   true,
		}
		globalEnvironment := NewEnvironment(&globalEnvironmentModel)
		globalEnvironmentService.environments = append(globalEnvironmentService.environments, globalEnvironment)

		// Registering the global vaults like the config ones
		configVault := configvault.New()
		globalEnvironment.RegisterVault(ctx, configVault, true)
	})
	return globalEnvironmentService
}

func GetInstance() interfaces.EnvironmentServiceInterface {
	if globalEnvironmentService == nil {
		panic("environment service not initialized")
	}
	return globalEnvironmentService
}

// Reset resets the singleton for testing purposes
func Reset() {
	environmentServiceMutex.Lock()
	defer environmentServiceMutex.Unlock()
	globalEnvironmentService = nil
	environmentServiceOnce = sync.Once{}
}

func new(environmentStore stores.EnvironmentDataStoreInterface) *EnvironmentService {
	result := &EnvironmentService{
		environmentStore: environmentStore,
	}
	result.environments = make([]interfaces.EnvironmentInterface, 0)
	result.functions = make(map[string]interfaces.EnvironmentVariableFunction)
	result.variables = make(map[string]map[string]interface{})
	result.isSynced = false
	result.mu = sync.RWMutex{}
	result.once = sync.Once{}
	return result
}

func (s *EnvironmentService) GetName() string {
	return "environment"
}

// RegisterFunction registers a function with the environment service
func (env *EnvironmentService) RegisterFunction(ctx *appctx.AppContext, function interfaces.EnvironmentVariableFunction) *diagnostics.Diagnostics {
	diag := diagnostics.New("register_function")
	defer diag.Complete()

	diag.AddPathEntry("start", "environment", map[string]interface{}{
		"function_name": function.Name(),
	})

	env.mu.Lock()
	defer env.mu.Unlock()

	// Check if function already exists
	if _, exists := env.functions[function.Name()]; exists {
		diag.AddWarning("FUNCTION_ALREADY_EXISTS", "Function already registered", "environment", map[string]interface{}{
			"function_name": function.Name(),
		})
		return diag
	}

	// Register the function
	env.functions[function.Name()] = function

	diag.AddPathEntry("function_registered", "environment", map[string]interface{}{
		"function_name":   function.Name(),
		"total_functions": len(env.functions),
	})

	logging.Infof("Registered function: %s", function.Name())
	return diag
}

func (s *EnvironmentService) GetEnvironments(ctx *appctx.AppContext, tenantID string) ([]models.Environment, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_environments")
	defer diag.Complete()
	result := make([]models.Environment, 0)

	// We will first append the global environment
	for _, environment := range s.environments {
		result = append(result, *environment.GetEnvironment())
	}

	environments, errDiags := s.environmentStore.GetEnvironments(ctx, tenantID)
	if errDiags.HasErrors() {
		diag.Append(errDiags)
		return nil, diag
	}
	result = append(result, mappers.MapEnvironmentsToDto(environments)...)
	return result, diag
}

func (s *EnvironmentService) GetPaginatedEnvironments(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[models.Environment], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_paginated_environments")
	defer diag.Complete()

	var queryBuilder *filters.QueryBuilder
	if pagination != nil {
		queryBuilder = filters.NewQueryBuilder("")
	}
	result := make([]models.Environment, 0)
	for _, environment := range s.environments {
		result = append(result, *environment.GetEnvironment())
	}

	environments, errDiags := s.environmentStore.GetEnvironmentsByQuery(ctx, tenantID, queryBuilder)
	if errDiags.HasErrors() {
		diag.Append(errDiags)
		return nil, diag
	}

	environmentsDto := mappers.MapEnvironmentsToDto(environments.Items)
	result = append(result, environmentsDto...)

	response := api_models.PaginationResponse[models.Environment]{
		Data:       result,
		TotalCount: environments.Total,
		Pagination: api_models.Pagination{
			Page:       environments.Page,
			PageSize:   environments.PageSize,
			TotalPages: environments.TotalPages,
		},
	}

	return &response, diag
}
