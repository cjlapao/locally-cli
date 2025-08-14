// Package handlers provides the API handlers for the certificates service
package handlers

import (
	"net/http"

	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/pkg/models"
)

type CertificatesApiHandlers struct {
	certificateService interfaces.CertificateServiceInterface
}

func NewCertificatesApiHandler(certificateService interfaces.CertificateServiceInterface) *CertificatesApiHandlers {
	return &CertificatesApiHandlers{
		certificateService: certificateService,
	}
}

func (h *CertificatesApiHandlers) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/root",
			Handler:     h.HandleGetRootCertificate,
			Description: "Get the root certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/certificates/root",
			Handler:     h.HandleCreateRootCertificate,
			Description: "Create a new root certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelSuperUser,
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/certificates/root",
			Handler:     h.HandleDeleteRootCertificate,
			Description: "Delete a root certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelSuperUser,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/ca",
			Handler:     h.HandleGetIntermediateCertificate,
			Description: "Get the intermediate certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/certificates/ca",
			Handler:     h.HandleCreateIntermediateCertificate,
			Description: "Create a new intermediate certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelSuperUser,
			},
		},
	}
}

func (h *CertificatesApiHandlers) HandleGetRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the root certificate")

	// dbCertificates, diag := h.certificateService.GetCertificate(ctx, config.RootCertificateSlug)
	// if diag.HasErrors() {
	// 	ctx.Log().Error("Error getting root certificate", diag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", diag)
	// 	return
	// }
	// if dbCertificates == nil {
	// 	ctx.Log().Error("Root certificate not found")
	// 	api.WriteNotFound(w, r, "Root certificate not found")
	// 	return
	// }

	// dtoModel := mappers.MapRootCertificateToDto(dbCertificates.GetCertificate())

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(dtoModel)

	ctx.Log().Info("Root certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleCreateRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating a new root certificate")

	// ctx.Log().Info("Checking if root certificate already exists")
	// // checking if we already have a root certificate with us, we can only get one root certificate per database
	// dbCertificate, dbCertDiag := h.certificateService.GetCertificate(ctx, config.RootCertificateSlug)
	// if dbCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error getting root certificate", dbCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", dbCertDiag)
	// 	return
	// }

	// if dbCertificate != nil {
	// 	ctx.Log().Info("Root certificate already exists")
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusOK, errors.ErrorGettingRootCertificate, "Root certificate already exists", nil)
	// 	return
	// }

	// ctx.Log().Info("Generating root certificate")
	// rootCA, dbCertDiag := h.certificateService.GenerateRootCertificate(ctx)
	// if dbCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error generating root certificate", dbCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingRootCertificate, "Error generating root certificate", dbCertDiag)
	// 	return
	// }

	// ctx.Log().Info("Persisting root certificate")
	// dbEntity := mappers.MapRootCertificateToEntity(rootCA.GetCertificate())
	// createdEntity, dbCertDiag := h.certificateService.CreateRootCertificate(ctx, &dbEntity)
	// if dbCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error creating root certificate", dbCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingRootCertificate, "Error creating root certificate", dbCertDiag)
	// 	return
	// }

	// result := mappers.MapRootCertificateToDto(*createdEntity)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(result)

	ctx.Log().Info("Root certificate created successfully")
}

func (h *CertificatesApiHandlers) HandleDeleteRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Deleting the root certificate")

	// diag := h.certificateService.DeleteRootCertificate(ctx, config.RootCertificateSlug)
	// if diag.HasErrors() {
	// 	ctx.Log().Error("Error deleting root certificate", diag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorDeletingRootCertificate, "Error deleting root certificate", diag)
	// 	return
	// }

	// response := api_models.StatusResponse{
	// 	ID:     config.RootCertificateSlug,
	// 	Status: "deleted",
	// }

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(response)

	ctx.Log().Info("Root certificate deleted successfully")
}

func (h *CertificatesApiHandlers) HandleGetIntermediateCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the intermediate certificate")

	// dbCertificates, diag := h.certificateService.GetCertificate(ctx, config.IntermediateCertificateSlug)
	// if diag.HasErrors() {
	// 	ctx.Log().Error("Error getting intermediate certificate", diag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingIntermediateCertificate, "Error getting intermediate certificate", diag)
	// 	return
	// }
	// if dbCertificates == nil {
	// 	ctx.Log().Error("Intermediate certificate not found")
	// 	api.WriteNotFound(w, r, "Intermediate certificate not found")
	// 	return
	// }

	// dtoModel := mappers.MapIntermediateCertificateToDto(*dbCertificates)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(dtoModel)

	ctx.Log().Info("Intermediate certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleCreateIntermediateCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating a new intermediate certificate")

	// dbCertificates, diag := h.certificateService.GetCertificate(ctx, config.IntermediateCertificateSlug)
	// if diag.HasErrors() {
	// 	ctx.Log().Error("Error getting intermediate certificate", diag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingIntermediateCertificate, "Error getting intermediate certificate", diag)
	// 	return
	// }

	// ctx.Log().Info("Checking if intermediate certificate already exists")
	// if dbCertificates != nil {
	// 	ctx.Log().Info("Intermediate certificate already exists")
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusOK, errors.ErrorGettingIntermediateCertificate, "Intermediate certificate already exists", nil)
	// 	return
	// }

	// ctx.Log().Info("Generating intermediate certificate")
	// rootCA, dbCertDiag := h.certificateService.GetCertificate(ctx, config.RootCertificateSlug)
	// if dbCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error getting root certificate", dbCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", dbCertDiag)
	// 	return
	// }
	// dtoRootCA := mappers.MapRootCertificateToDto(*rootCA)

	// intermediateCA, intermediateCertDiag := h.certificateService.GenerateIntermediateCertificate(ctx, &dtoRootCA)
	// if intermediateCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error generating intermediate certificate", intermediateCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingIntermediateCertificate, "Error generating intermediate certificate", intermediateCertDiag)
	// 	return
	// }

	// ctx.Log().Info("Persisting intermediate certificate")
	// dbEntity := mappers.MapIntermediateCertificateToEntity(*intermediateCA)
	// createdEntity, createIntermediateCertDiag := h.certificateService.CreateIntermediateCertificate(ctx, &dbEntity)
	// if createIntermediateCertDiag.HasErrors() {
	// 	ctx.Log().Error("Error creating intermediate certificate", createIntermediateCertDiag.Errors)
	// 	api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingIntermediateCertificate, "Error creating intermediate certificate", createIntermediateCertDiag)
	// 	return
	// }

	// result := mappers.MapIntermediateCertificateToDto(*createdEntity)

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(result)

	ctx.Log().Info("Intermediate certificate created successfully")
}
