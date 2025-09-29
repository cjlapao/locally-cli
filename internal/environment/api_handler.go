package environment

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/cjlapao/locally-cli/internal/api"
// 	api_types "github.com/cjlapao/locally-cli/internal/api/types"
// 	"github.com/cjlapao/locally-cli/internal/appctx"
// 	"github.com/cjlapao/locally-cli/pkg/models"
// 	"github.com/gorilla/mux"
// )

// type ApiHandler struct {
// 	environment *Environment
// }

// func NewApiHandler(environment *Environment) *ApiHandler {
// 	return &ApiHandler{environment: environment}
// }

// func (h *ApiHandler) Routes() []api_types.Route {
// 	return []api_types.Route{
// 		{
// 			Method:      http.MethodGet,
// 			Path:        "/v1/environment/vaults",
// 			Handler:     h.HandleGetVaults,
// 			Description: "Get all vaults from the environment",
// 			SecurityRequirement: &api_types.SecurityRequirement{
// 				SecurityLevel: models.ApiKeySecurityLevelAny,
// 			},
// 		},
// 		{
// 			Method:      http.MethodGet,
// 			Path:        "/v1/environment/{vault_name}/get/{key}",
// 			Handler:     h.HandleGetVaultKey,
// 			Description: "Get a specific vault from the  environment",
// 			SecurityRequirement: &api_types.SecurityRequirement{
// 				SecurityLevel: models.ApiKeySecurityLevelAny,
// 			},
// 		},
// 	}
// }

// func (h *ApiHandler) HandleGetVaults(w http.ResponseWriter, r *http.Request) {
// 	ctx := appctx.FromContext(r.Context())
// 	ctx.LogInfo("Getting all vaults from the environment")
// 	vaults := h.environment.ListVaults(ctx)

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(vaults)
// 	ctx.LogWithField("count", len(vaults)).Info("Vaults retrieved successfully")
// }

// func (h *ApiHandler) HandleGetVaultKey(w http.ResponseWriter, r *http.Request) {
// 	ctx := appctx.FromContext(r.Context())
// 	vaultName := mux.Vars(r)["vault_name"]
// 	key := mux.Vars(r)["key"]
// 	vault, exists := h.environment.GetVault(ctx, vaultName)
// 	if !exists {
// 		ctx.LogError("Vault not found")
// 		api.WriteError(w, r, http.StatusInternalServerError, "Vault not found", "vault_name", vaultName)
// 		return
// 	}
// 	if vault == nil {
// 		ctx.LogError("Vault not found")
// 		api.WriteError(w, r, http.StatusNotFound, "Vault not found", "vault_name", vaultName)
// 		return
// 	}

// 	value, exists := vault.Get(key)
// 	if !exists {
// 		ctx.LogError("Key not found")
// 		api.WriteError(w, r, http.StatusNotFound, "Key not found", "vault_name", vaultName, "key", key)
// 		return
// 	}

// 	response := map[string]interface{}{
// 		"value": value,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(response)
// }
