package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// MigrationWorker represents a database seed or migration worker
type MigrationWorker interface {
	// GetName returns the unique name/identifier for this seed
	GetName() string

	// GetDescription returns a human-readable description of what this seed does
	GetDescription() string

	// GetVersion returns the version number for ordering (optional)
	GetVersion() int

	// Up applies the seed/migration
	Up(ctx *appctx.AppContext) *diagnostics.Diagnostics

	// Down rolls back the seed/migration
	Down(ctx *appctx.AppContext) *diagnostics.Diagnostics
}
