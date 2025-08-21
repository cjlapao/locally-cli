// Package service provides the API keys service implementation.
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"sync"
	"time"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	api_keys_interfaces "github.com/cjlapao/locally-cli/internal/api_keys/interfaces"
	"github.com/cjlapao/locally-cli/internal/api_keys/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/internal/system"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

var (
	globalApiKeysService *ApiKeysService
	apiKeysServiceOnce   sync.Once
	apiKeysServiceMutex  sync.Mutex
)

type ApiKeysService struct {
	apiKeyStore   stores.ApiKeyStoreInterface
	claimStore    stores.ClaimDataStoreInterface
	systemService system_interfaces.SystemServiceInterface
}

func Initialize(apiKeyStore stores.ApiKeyStoreInterface) api_keys_interfaces.ApiKeysServiceInterface {
	apiKeysServiceMutex.Lock()
	defer apiKeysServiceMutex.Unlock()

	apiKeysServiceOnce.Do(func() {
		globalApiKeysService = new(apiKeyStore)
	})
	return globalApiKeysService
}

func GetInstance() api_keys_interfaces.ApiKeysServiceInterface {
	if globalApiKeysService == nil {
		panic("api keys service not initialized")
	}
	return globalApiKeysService
}

// Reset resets the singleton for testing purposes
func Reset() {
	apiKeysServiceMutex.Lock()
	defer apiKeysServiceMutex.Unlock()
	globalApiKeysService = nil
	apiKeysServiceOnce = sync.Once{}
}

func new(apiKeyStore stores.ApiKeyStoreInterface) *ApiKeysService {
	return &ApiKeysService{
		apiKeyStore:   apiKeyStore,
		claimStore:    stores.GetClaimDataStoreInstance(),
		systemService: system.GetInstance(),
	}
}

func (s *ApiKeysService) GetName() string {
	return "api_keys"
}

// GenerateSecureAPIKey generates a cryptographically secure API key
func (s *ApiKeysService) GenerateSecureAPIKey() (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_secure_api_key")

	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		diag.AddError("generate_secure_api_key", "failed to generate random bytes", "api_key", map[string]interface{}{
			"error": err.Error(),
		})
		return "", diag
	}

	// Encode as base64 and remove padding
	key := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)

	// Add prefix for identification
	prefix := config.ApiKeyPrefix
	return prefix + key, diag
}

// getClaimsBySlugs retrieves claim entities from the database using claim slugs
func (s *ApiKeysService) getClaimsBySlugs(ctx *appctx.AppContext, tenantID string, claimSlugs []string) ([]entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claims_by_slugs")
	defer diag.Complete()

	var claims []entities.Claim
	for _, slug := range claimSlugs {
		claim, getClaimDiag := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, slug)
		if getClaimDiag.HasErrors() {
			diag.Append(getClaimDiag)
			return nil, diag
		}
		if claim == nil {
			diag.AddError("claim_not_found", "Claim not found", "api_key", map[string]interface{}{
				"slug": slug,
			})
			return nil, diag
		}
		claims = append(claims, *claim)
	}

	return claims, diag
}

// associateClaimsWithApiKey associates claims with an API key using GORM's many-to-many relationship
func (s *ApiKeysService) associateClaimsWithApiKey(ctx *appctx.AppContext, apiKey *entities.ApiKey, claims []entities.Claim) *diagnostics.Diagnostics {
	diag := diagnostics.New("associate_claims_with_api_key")
	defer diag.Complete()

	// First, clear any existing claim associations
	if err := s.apiKeyStore.GetDB().Model(apiKey).Association("Claims").Clear(); err != nil {
		diag.AddError("failed_to_clear_claims", "Failed to clear existing claim associations", "api_key", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	// Then add the new claim associations
	if err := s.apiKeyStore.GetDB().Model(apiKey).Association("Claims").Append(claims); err != nil {
		diag.AddError("failed_to_associate_claims", "Failed to associate claims with API key", "api_key", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *ApiKeysService) GetApiKeys(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.ApiKey], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_keys")
	defer diag.Complete()

	var queryBuilder *filters.QueryBuilder
	if pagination != nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	queryResponse, getApisDiag := s.apiKeyStore.GetApiKeysByQuery(ctx, tenantID, queryBuilder)
	if getApisDiag.HasErrors() {
		diag.Append(getApisDiag)
		return nil, diag
	}

	apiKeysDto := mappers.MapApiKeysToDto(queryResponse.Items)

	response := api_models.PaginationResponse[pkg_models.ApiKey]{
		TotalCount: queryResponse.Total,
		Data:       apiKeysDto,
		Pagination: api_models.Pagination{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalPages: queryResponse.TotalPages,
		},
	}

	return &response, diag
}

func (s *ApiKeysService) GetApiKeyByID(ctx *appctx.AppContext, tenantID string, id string) (*pkg_models.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_id")
	defer diag.Complete()

	apiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return nil, diag
	}
	if apiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key", map[string]interface{}{
			"id": id,
		})
		return nil, diag
	}

	apiKeyDto := mappers.MapApiKeyToDto(apiKey)

	return apiKeyDto, diag
}

func (s *ApiKeysService) CreateApiKey(ctx *appctx.AppContext, tenantID string, request *models.CreateApiKeyRequest) (*pkg_models.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_api_key")
	defer diag.Complete()

	// checking if we already have an api key with the same name
	existingApiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByName(ctx, tenantID, request.Name)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return nil, diag
	}

	if existingApiKey != nil {
		diag.AddError("api_key_already_exists", "API key already exists", "api_key", map[string]interface{}{
			"name": request.Name,
		})
		return nil, diag
	}

	// Validate that either claims or security level is provided, but not both
	if len(request.Claims) > 0 && request.SecurityLevel != "" {
		diag.AddError("invalid_request", "Cannot specify both claims and security level", "api_key", nil)
		return nil, diag
	}

	if len(request.Claims) == 0 && request.SecurityLevel == "" {
		diag.AddError("invalid_request", "Must specify either claims or security level", "api_key", nil)
		return nil, diag
	}

	// parse the expires at date
	expiresAt, err := time.Parse("02-01-2006", request.ExpiresAt)
	if err != nil {
		diag.AddError("invalid_request", "Invalid expires at date", "api_key", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Generate a secure API key
	apiKeyString, keyDiag := s.GenerateSecureAPIKey()
	if keyDiag.HasErrors() {
		diag.Append(keyDiag)
		return nil, diag
	}

	// compute deterministic digest and check duplicates (extremely unlikely)
	sha := sha256.Sum256([]byte(apiKeyString))
	digest := hex.EncodeToString(sha[:])
	existingApiKey, getApiKeyDiag = s.apiKeyStore.GetApiKeyByDigest(ctx, tenantID, digest)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return nil, diag
	}

	if existingApiKey != nil {
		diag.AddError("api_key_already_exists", "API key already exists", "api_key", map[string]interface{}{
			"digest": digest,
		})
		return nil, diag
	}

	// Determine the claims to assign
	var claimsToAssign []string
	if request.SecurityLevel != "" {
		// Get default claims for the security level
		defaultClaims := s.systemService.GenerateDefaultClaimsForSecurityLevel(request.SecurityLevel)
		for _, claim := range defaultClaims {
			claimsToAssign = append(claimsToAssign, claim.Slug)
		}
	} else {
		// Use the provided claims
		claimsToAssign = request.Claims
	}
	apiKeyClaims := make([]entities.Claim, len(claimsToAssign))
	for i, claim := range claimsToAssign {
		dbClaim, getClaimDiag := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, claim)
		if getClaimDiag.HasErrors() {
			diag.Append(getClaimDiag)
			return nil, diag
		}
		apiKeyClaims[i] = *dbClaim
	}

	// Create the API key entity
	apiKey := &entities.ApiKey{
		Name:      request.Name,
		KeyHash:   apiKeyString,                              // Will be hashed in CreateAPIKey
		KeyDigest: digest,                                    // Will be encoded and stored as deterministic digest
		KeyPrefix: apiKeyString[:8+len(config.ApiKeyPrefix)], // First 8 chars after prefix
		ExpiresAt: &expiresAt,
		IsActive:  true,
		Claims:    apiKeyClaims,
	}

	// Save to database
	dbAPIKey, createApiKeyDiag := s.apiKeyStore.CreateApiKey(ctx, tenantID, apiKey)
	if createApiKeyDiag.HasErrors() {
		diag.Append(createApiKeyDiag)
		return nil, diag
	}

	// Set the tenant ID after creation since it's part of the embedded struct
	// This is a workaround for the linter issue with embedded fields
	dbAPIKey.TenantID = tenantID

	// Associate claims with the API key
	if len(claimsToAssign) > 0 {
		// Get the claim entities from the database
		claims, claimsDiag := s.getClaimsBySlugs(ctx, tenantID, claimsToAssign)
		if claimsDiag.HasErrors() {
			diag.Append(claimsDiag)
			return nil, diag
		}

		// Associate the claims with the API key
		assocDiag := s.associateClaimsWithApiKey(ctx, dbAPIKey, claims)
		if assocDiag.HasErrors() {
			diag.Append(assocDiag)
			return nil, diag
		}
	}

	apiKeyDto := mappers.MapApiKeyToDto(dbAPIKey)

	plaintextKey := apiKeyString
	apiKeyDto.KeyHash = "" // never return hash
	apiKeyDto.PlaintextKey = plaintextKey
	return apiKeyDto, diag
}

func (s *ApiKeysService) DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_api_key")
	defer diag.Complete()

	// check if the api key exists
	apiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if apiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key", map[string]interface{}{
			"id": id,
		})
		return diag
	}

	deleteApiKeyDiag := s.apiKeyStore.DeleteApiKey(ctx, tenantID, id)
	if deleteApiKeyDiag.HasErrors() {
		diag.Append(deleteApiKeyDiag)
		return diag
	}

	return diag
}

func (s *ApiKeysService) RevokeApiKey(ctx *appctx.AppContext, tenantID string, revokedBy string, id string, request *models.RevokeApiKeyRequest) *diagnostics.Diagnostics {
	diag := diagnostics.New("revoke_api_key")
	defer diag.Complete()

	// check if the api key exists
	apiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if apiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key", map[string]interface{}{
			"id": id,
		})
		return diag
	}

	revokeApiKeyDiag := s.apiKeyStore.RevokeApiKey(ctx, tenantID, id, revokedBy, request.RevocationReason)
	if revokeApiKeyDiag.HasErrors() {
		diag.Append(revokeApiKeyDiag)
		return diag
	}

	return diag
}

func (s *ApiKeysService) AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_claim_to_api_key")
	defer diag.Complete()

	// check if the api key exists
	apiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}

	if apiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key", map[string]interface{}{
			"id": id,
		})
		return diag
	}

	// check if the claim exists
	claim, getClaimDiag := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return diag
	}

	if claim == nil {
		diag.AddError("claim_not_found", "Claim not found", "api_key", map[string]interface{}{
			"id": claimID,
		})
		return diag
	}

	// add the claim to the api key
	addClaimDiag := s.apiKeyStore.AddClaimToApiKey(ctx, tenantID, id, claimID)
	if addClaimDiag.HasErrors() {
		diag.Append(addClaimDiag)
		return diag
	}

	return diag
}

func (s *ApiKeysService) RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_api_key")
	defer diag.Complete()

	// check if the api key exists
	apiKey, getApiKeyDiag := s.apiKeyStore.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}

	if apiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key", map[string]interface{}{
			"id": id,
		})
		return diag
	}

	// check if the claim exists
	claim, getClaimDiag := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return diag
	}

	if claim == nil {
		diag.AddError("claim_not_found", "Claim not found", "api_key", map[string]interface{}{
			"id": claimID,
		})
		return diag
	}

	// remove the claim from the api key
	removeClaimDiag := s.apiKeyStore.RemoveClaimFromApiKey(ctx, tenantID, id, claimID)
	if removeClaimDiag.HasErrors() {
		diag.Append(removeClaimDiag)
		return diag
	}

	return diag
}
