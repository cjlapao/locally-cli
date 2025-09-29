package service

import (
	"strings"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
)

type Environment struct {
	mu            sync.RWMutex
	base          *models.Environment
	vaults        []interfaces.EnvironmentVault
	functions     []interfaces.EnvironmentVariableFunction
	variables     map[string]map[string]interface{}
	isSynced      bool
	IsInitialized bool
	CanPersist    bool
}

func NewEnvironment(baseModel *models.Environment) *Environment {
	if baseModel == nil {
		baseModel = &models.Environment{
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
	}

	return &Environment{
		vaults:        make([]interfaces.EnvironmentVault, 0),
		functions:     make([]interfaces.EnvironmentVariableFunction, 0),
		variables:     make(map[string]map[string]interface{}),
		isSynced:      false,
		IsInitialized: false,
		CanPersist:    false,
		mu:            sync.RWMutex{},
		base:          baseModel,
	}
}

func (env *Environment) GetName() string {
	return env.base.Name
}

func (env *Environment) GetVault(name string) (*interfaces.EnvironmentVault, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_vault")
	defer diag.Complete()

	env.mu.RLock()
	defer env.mu.RUnlock()

	for _, vault := range env.vaults {
		if strings.EqualFold(vault.Name(), name) {
			return &vault, diag
		}
	}

	diag.AddError("VAULT_NOT_FOUND", "Vault not found", "environment", map[string]interface{}{
		"vault_name": name,
	})
	return nil, diag
}

func (env *Environment) GetAvailableVaults() []interfaces.EnvironmentVault {
	result := make([]interfaces.EnvironmentVault, 0)
	for _, vault := range env.vaults {
		if !vault.IsEnabled() {
			continue
		}
		result = append(result, vault)
	}
	return result
}

func (env *Environment) GetVaults() []interfaces.EnvironmentVault {
	return env.vaults
}

// RegisterVault registers a vault with the environment service
func (env *Environment) RegisterVault(ctx *appctx.AppContext, vault interfaces.EnvironmentVault, sync bool) *diagnostics.Diagnostics {
	diag := diagnostics.New("register_vault")
	defer diag.Complete()

	diag.AddPathEntry("start", "environment", map[string]interface{}{
		"vault_name": vault.Name(),
	})

	env.mu.Lock()
	defer env.mu.Unlock()

	// Check if vault already exists
	for _, existingVault := range env.vaults {
		if strings.EqualFold(existingVault.Name(), vault.Name()) {
			diag.AddWarning("VAULT_ALREADY_EXISTS", "Vault already registered", "environment", map[string]interface{}{
				"vault_name": vault.Name(),
			})
			return diag
		}
	}

	// we will sync the vault if the sync flag is true
	if sync {
		_, diag = vault.Sync()

		if diag.HasErrors() {
			diag.AddError("VAULT_SYNC_FAILED", "Failed to sync vault", "environment", map[string]interface{}{
				"vault_name": vault.Name(),
			})
			return diag
		}
	}

	// Register the vault
	env.vaults = append(env.vaults, vault)
	env.variables[vault.Name()] = make(map[string]interface{})

	diag.AddPathEntry("vault_registered", "environment", map[string]interface{}{
		"vault_name":   vault.Name(),
		"total_vaults": len(env.vaults),
	})

	logging.Infof("Registered vault: %s", vault.Name())
	return diag
}

func (env *Environment) RegisterFunction(ctx *appctx.AppContext, function interfaces.EnvironmentVariableFunction) *diagnostics.Diagnostics {
	diag := diagnostics.New("register_function")
	defer diag.Complete()

	diag.AddPathEntry("start", "environment", map[string]interface{}{
		"function_name": function.Name(),
	})

	env.functions = append(env.functions, function)

	diag.AddPathEntry("function_registered", "environment", map[string]interface{}{
		"function_name":   function.Name(),
		"total_functions": len(env.functions),
	})

	logging.Infof("Registered function: %s", function.Name())
	return diag
}

func (env *Environment) GetFunction(name string) (interfaces.EnvironmentVariableFunction, bool) {
	for _, function := range env.functions {
		if strings.EqualFold(function.Name(), name) {
			return function, true
		}
	}
	return nil, false
}

func (env *Environment) GetEnvironment() *models.Environment {
	return env.base
}
