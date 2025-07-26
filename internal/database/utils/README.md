# Database Utils

This package contains utility functions for database operations.

## Pagination Helper

The pagination helper provides generic functions to execute paginated queries without repeating the same pagination logic across different data stores.

### Functions

#### `PaginatedQuery[T any]`

Executes a paginated query and returns a `FilterResponse[T]`.

**Parameters:**
- `db *gorm.DB` - The database connection
- `filterObj *filters.Filter` - The filter object containing pagination and filtering criteria
- `model T` - The model type to query (used for type inference)

**Returns:**
- `*filters.FilterResponse[T]` - The paginated response with items, total count, page, and page size
- `error` - Any error that occurred during the query

**Example:**
```go
func (s *TenantDataStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
    return utils.PaginatedQuery(s.GetDB(), filterObj, entities.Tenant{})
}
```

#### `PaginatedQueryWithPreload[T any]`

Executes a paginated query with preloads and returns a `FilterResponse[T]`.

**Parameters:**
- `db *gorm.DB` - The database connection
- `filterObj *filters.Filter` - The filter object containing pagination and filtering criteria
- `model T` - The model type to query (used for type inference)
- `preloads ...string` - Variable number of preload relationships to include

**Returns:**
- `*filters.FilterResponse[T]` - The paginated response with items, total count, page, and page size
- `error` - Any error that occurred during the query

**Example:**
```go
func (s *UserDataStore) GetUsersByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
    return utils.PaginatedQueryWithPreload(s.GetDB(), filterObj, entities.User{}, "Roles", "Claims")
}
```

### Benefits

1. **DRY Principle**: Eliminates code duplication across data stores
2. **Type Safety**: Uses Go generics to ensure type safety
3. **Consistency**: Ensures all paginated queries follow the same pattern
4. **Maintainability**: Changes to pagination logic only need to be made in one place
5. **Flexibility**: Supports both simple queries and queries with preloads

### Migration Guide

To migrate existing pagination code to use the helper:

**Before:**
```go
func (s *TenantDataStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
    var tenants []entities.Tenant
    filterString, args := filterObj.Generate()
    pageIndex := filterObj.Page - 1
    pageSize := filterObj.PageSize
    total := int64(0)
    offset := pageIndex * pageSize
    
    // Get total count of tenants
    if err := s.GetDB().Model(&entities.Tenant{}).Where(filterString, args...).Count(&total).Error; err != nil {
        return nil, err
    }
    
    // Get tenants with pagination, if no page size is provided, return all tenants
    if pageSize == -1 {
        if err := s.GetDB().Where(filterString, args...).Find(&tenants).Error; err != nil {
            return nil, err
        }
    } else {
        if err := s.GetDB().Where(filterString, args...).Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
            return nil, err
        }
    }

    response := filters.FilterResponse[entities.Tenant]{
        Items:    tenants,
        Total:    total,
        Page:     filterObj.Page,
        PageSize: filterObj.PageSize,
    }

    return &response, nil
}
```

**After:**
```go
func (s *TenantDataStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
    return utils.PaginatedQuery(s.GetDB(), filterObj, entities.Tenant{})
}
```

The migration reduces the code from ~25 lines to just 1 line while maintaining the exact same functionality.

### Advanced Examples

#### Example 1: Simple Pagination
```go
func (s *TenantDataStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
    return utils.PaginatedQuery(s.GetDB(), filterObj, entities.Tenant{})
}
```

#### Example 2: Pagination with Preloads
```go
func (s *UserDataStore) GetUsersByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
    return utils.PaginatedQueryWithPreload(s.GetDB(), filterObj, entities.User{}, "Roles", "Claims")
}
```

#### Example 3: Adding Additional Filters
```go
func (s *AuthDataStore) ListAPIKeysByUserIDWithFilter(ctx *appctx.AppContext, userID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.APIKey], error) {
    // Add the user_id filter to the existing filter
    filterObj.WithField("user_id", filters.FilterOperatorEqual, userID, filters.FilterJoinerAnd)
    
    // Use the generic pagination helper
    return utils.PaginatedQuery(s.GetDB(), filterObj, entities.APIKey{})
}
```

#### Example 4: Complex Filtering with Multiple Conditions
```go
func (s *MessageDataStore) GetMessagesByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Message], error) {
    // Add default filters if not provided
    if filterObj.Page <= 0 {
        filterObj.WithPage(1)
    }
    if filterObj.PageSize <= 0 {
        filterObj.WithPageSize(20)
    }
    
    // Add default ordering by created_at DESC
    // Note: This would require extending the Filter struct to support ordering
    
    return utils.PaginatedQueryWithPreload(s.GetDB(), filterObj, entities.Message{}, "Sender", "Recipients")
}
```

### Usage Patterns

1. **Simple List with Pagination**: Use `PaginatedQuery` for basic pagination
2. **List with Related Data**: Use `PaginatedQueryWithPreload` when you need to include related entities
3. **Filtered Lists**: Add filters to the filter object before calling the pagination helper
4. **User-Specific Data**: Add user-specific filters (like user_id) to scope the results

### Best Practices

1. **Always provide default pagination values** in your service layer if not provided by the client
2. **Use preloads sparingly** - only include relationships that are actually needed
3. **Consider adding ordering support** to the Filter struct for consistent sorting
4. **Handle edge cases** like empty results gracefully
5. **Add appropriate indexes** to your database for the fields you commonly filter on 