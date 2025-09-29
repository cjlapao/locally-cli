package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
)

type EnvironmentInterface interface {
	GetName() string
	GetVault(name string) (*EnvironmentVault, *diagnostics.Diagnostics)
	GetAvailableVaults() []EnvironmentVault
	GetVaults() []EnvironmentVault
	RegisterVault(ctx *appctx.AppContext, vault EnvironmentVault, sync bool) *diagnostics.Diagnostics
	GetEnvironment() *models.Environment
}
