package stores

import (
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestTenantDataStoreWithSQLite demonstrates testing with real SQLite database
func TestTenantDataStoreWithSQLite(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create the store with the test database
	store := &TenantDataStore{
		BaseDataStore: *database.NewBaseDataStore(db),
	}

	// Run migrations
	diag := store.Migrate()
	assert.False(t, diag.HasErrors())

	// Test data
	ctx := appctx.NewContext(nil)
	testTenant := &entities.Tenant{
		Name:        "Test Tenant",
		Description: "Test Description",
		Domain:      "test.com",
		Status:      "active",
	}

	t.Run("CreateTenant", func(t *testing.T) {
		createdTenant, err := store.CreateTenant(ctx, testTenant)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdTenant.ID)
		assert.NotEmpty(t, createdTenant.Slug)
		assert.Equal(t, "test-tenant", createdTenant.Slug)
		assert.Equal(t, testTenant.Name, createdTenant.Name)
		assert.True(t, createdTenant.CreatedAt.After(time.Now().Add(-time.Second)))
	})

	t.Run("GetTenantByID", func(t *testing.T) {
		// First create a tenant
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Get By ID Tenant",
			Domain: "getbyid.com",
		})
		require.NoError(t, err)

		// Then retrieve it
		retrievedTenant, err := store.GetTenantByID(ctx, createdTenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, createdTenant.ID, retrievedTenant.ID)
		assert.Equal(t, createdTenant.Name, retrievedTenant.Name)
	})

	t.Run("GetTenantBySlug", func(t *testing.T) {
		// First create a tenant
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Get By Slug Tenant",
			Domain: "getbyslug.com",
		})
		require.NoError(t, err)

		// Then retrieve it by slug
		retrievedTenant, err := store.GetTenantBySlug(ctx, createdTenant.Slug)
		assert.NoError(t, err)
		assert.Equal(t, createdTenant.ID, retrievedTenant.ID)
		assert.Equal(t, createdTenant.Slug, retrievedTenant.Slug)
	})

	t.Run("GetTenantByIdOrSlug", func(t *testing.T) {
		// First create a tenant
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Get By ID or Slug Tenant",
			Domain: "getbyidorslug.com",
		})
		require.NoError(t, err)

		// Test by ID
		retrievedByID, err := store.GetTenantByIdOrSlug(ctx, createdTenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, createdTenant.ID, retrievedByID.ID)

		// Test by Slug
		retrievedBySlug, err := store.GetTenantByIdOrSlug(ctx, createdTenant.Slug)
		assert.NoError(t, err)
		assert.Equal(t, createdTenant.ID, retrievedBySlug.ID)
	})

	t.Run("GetTenants", func(t *testing.T) {
		// Create multiple tenants
		tenant1, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Tenant 1",
			Domain: "tenant1.com",
		})
		require.NoError(t, err)

		tenant2, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Tenant 2",
			Domain: "tenant2.com",
		})
		require.NoError(t, err)

		// Get all tenants
		tenants, err := store.GetTenants(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(tenants), 2)

		// Verify our tenants are in the list
		tenantIDs := make(map[string]bool)
		for _, tenant := range tenants {
			tenantIDs[tenant.ID] = true
		}
		assert.True(t, tenantIDs[tenant1.ID])
		assert.True(t, tenantIDs[tenant2.ID])
	})

	t.Run("GetTenantsByFilter", func(t *testing.T) {
		// Create a tenant for testing
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Filter Test Tenant",
			Domain: "filtertest.com",
		})
		require.NoError(t, err)

		// Test pagination
		filter := &filters.Filter{
			Page:     1,
			PageSize: 10,
		}

		result, err := store.GetTenantsByFilter(ctx, filter)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.Total, int64(1))
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)
		assert.GreaterOrEqual(t, len(result.Items), 1)

		// Verify our tenant is in the results
		found := false
		for _, tenant := range result.Items {
			if tenant.ID == createdTenant.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created tenant should be in the filtered results")
	})

	t.Run("UpdateTenant", func(t *testing.T) {
		// Create a tenant
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Update Test Tenant",
			Domain: "updatetest.com",
		})
		require.NoError(t, err)

		// Update the tenant
		createdTenant.Name = "Updated Tenant Name"
		createdTenant.Description = "Updated Description"

		err = store.UpdateTenant(ctx, createdTenant)
		assert.NoError(t, err)

		// Retrieve and verify the update
		updatedTenant, err := store.GetTenantByID(ctx, createdTenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Tenant Name", updatedTenant.Name)
		assert.Equal(t, "Updated Description", updatedTenant.Description)
		assert.Equal(t, "updated-tenant-name", updatedTenant.Slug)
	})

	t.Run("DeleteTenant", func(t *testing.T) {
		// Create a tenant
		createdTenant, err := store.CreateTenant(ctx, &entities.Tenant{
			Name:   "Delete Test Tenant",
			Domain: "deletetest.com",
		})
		require.NoError(t, err)

		// Delete the tenant
		err = store.DeleteTenant(ctx, createdTenant.ID)
		assert.NoError(t, err)

		// Verify it's deleted
		_, err = store.GetTenantByID(ctx, createdTenant.ID)
		assert.Error(t, err) // Should return error for deleted tenant
	})

	t.Run("NotFound Scenarios", func(t *testing.T) {
		// Test getting non-existent tenant by ID
		_, err := store.GetTenantByID(ctx, "non-existent-id")
		assert.Error(t, err)

		// Test getting non-existent tenant by slug
		_, err = store.GetTenantBySlug(ctx, "non-existent-slug")
		assert.Error(t, err)

		// Test getting non-existent tenant by ID or slug
		_, err = store.GetTenantByIdOrSlug(ctx, "non-existent")
		assert.Error(t, err)
	})
}

// TestTenantDataStoreBasicOperations tests basic operations without concurrency
func TestTenantDataStoreBasicOperations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := &TenantDataStore{
		BaseDataStore: *database.NewBaseDataStore(db),
	}

	// Run migrations
	diag := store.Migrate()
	assert.False(t, diag.HasErrors())

	ctx := appctx.NewContext(nil)

	// Test basic CRUD operations
	t.Run("BasicCRUD", func(t *testing.T) {
		// Create
		tenant := &entities.Tenant{
			Name:   "Basic Test Tenant",
			Domain: "basictest.com",
		}

		createdTenant, err := store.CreateTenant(ctx, tenant)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdTenant.ID)
		assert.Equal(t, "basic-test-tenant", createdTenant.Slug)

		// Read
		retrievedTenant, err := store.GetTenantByID(ctx, createdTenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, createdTenant.ID, retrievedTenant.ID)

		// Update
		createdTenant.Name = "Updated Basic Tenant"
		err = store.UpdateTenant(ctx, createdTenant)
		assert.NoError(t, err)

		// Verify update
		updatedTenant, err := store.GetTenantByID(ctx, createdTenant.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Basic Tenant", updatedTenant.Name)

		// Delete
		err = store.DeleteTenant(ctx, createdTenant.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = store.GetTenantByID(ctx, createdTenant.ID)
		assert.Error(t, err) // Should return error for deleted tenant
	})
}

// TestTenantDataStoreEdgeCases tests edge cases and error conditions
func TestTenantDataStoreEdgeCases(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := &TenantDataStore{
		BaseDataStore: *database.NewBaseDataStore(db),
	}

	diag := store.Migrate()
	assert.False(t, diag.HasErrors())

	ctx := appctx.NewContext(nil)

	t.Run("EmptyName", func(t *testing.T) {
		tenant := &entities.Tenant{
			Name:   "",
			Domain: "emptyname.com",
		}

		createdTenant, err := store.CreateTenant(ctx, tenant)
		assert.NoError(t, err)
		assert.Equal(t, "", createdTenant.Slug) // Slug should be empty for empty name
	})

	t.Run("SpecialCharactersInName", func(t *testing.T) {
		tenant := &entities.Tenant{
			Name:   "Special Characters: @#$%^&*()",
			Domain: "specialchars.com",
		}

		createdTenant, err := store.CreateTenant(ctx, tenant)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdTenant.Slug)
		assert.NotEqual(t, tenant.Name, createdTenant.Slug) // Slug should be slugified
	})

	t.Run("DuplicateDomain", func(t *testing.T) {
		tenant1 := &entities.Tenant{
			Name:   "First Tenant",
			Domain: "duplicate.com",
		}

		tenant2 := &entities.Tenant{
			Name:   "Second Tenant",
			Domain: "duplicate.com", // Same domain
		}

		_, err := store.CreateTenant(ctx, tenant1)
		assert.NoError(t, err)

		_, err = store.CreateTenant(ctx, tenant2)
		assert.Error(t, err)
	})
}
