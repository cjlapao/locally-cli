package service

import (
	"strings"
	"sync"

	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
)

type Environments struct {
	mu                    sync.RWMutex
	IsInitialized         bool
	IsSynced              bool
	AvailableEnvironments []interfaces.EnvironmentInterface
	functions             map[string]interfaces.EnvironmentVariableFunction
}

func NewEnvironments() *Environments {
	return &Environments{
		mu:                    sync.RWMutex{},
		IsInitialized:         false,
		IsSynced:              false,
		AvailableEnvironments: make([]interfaces.EnvironmentInterface, 0),
	}
}

func (e *Environments) AddEnvironment(environment interfaces.EnvironmentInterface) {
	for _, existingEnvironment := range e.AvailableEnvironments {
		if existingEnvironment.GetName() == environment.GetName() {
			return
		}
	}

	e.AvailableEnvironments = append(e.AvailableEnvironments, environment)
}

func (e *Environments) GetEnvironment(slug string) (interfaces.EnvironmentInterface, bool) {
	for _, environment := range e.AvailableEnvironments {
		if strings.EqualFold(environment.GetName(), slug) {
			return environment, true
		}
	}
	return nil, false
}

func (e *Environments) GetAvailableEnvironments() []interfaces.EnvironmentInterface {
	return e.AvailableEnvironments
}

func (e *Environments) GetAvailableVaults() []interfaces.EnvironmentVault {
	vaults := make([]interfaces.EnvironmentVault, 0)
	for _, environment := range e.AvailableEnvironments {
		vaults = append(vaults, environment.GetAvailableVaults()...)
	}
	return vaults
}

func (e *Environments) GetAvailableVaultItems() []interfaces.EnvironmentVaultItem {
	items := make([]interfaces.EnvironmentVaultItem, 0)
	for _, environment := range e.AvailableEnvironments {
		vaults := environment.GetVaults()
		for _, vault := range vaults {
			items = append(items, vault.GetItems()...)
		}
	}
	return items
}
