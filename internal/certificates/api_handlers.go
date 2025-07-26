package certificates

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/errors"
	"github.com/cjlapao/locally-cli/internal/mappers"
)

type ApiHandlers struct {
	certificateService *CertificateService
	store              *stores.CertificatesDataStore
}

func NewApiHandlers(certificateService *CertificateService, store *stores.CertificatesDataStore) *ApiHandlers {
	return &ApiHandlers{
		certificateService: certificateService,
		store:              store,
	}
}

func (h *ApiHandlers) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/certificates/root",
			Handler:      h.HandleGetRootCertificate,
			Description:  "Get the root certificate",
			AuthRequired: true,
		},
		{
			Method:            http.MethodPost,
			Path:              "/v1/certificates/root",
			Handler:           h.HandleCreateRootCertificate,
			Description:       "Create a new root certificate",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:            http.MethodDelete,
			Path:              "/v1/certificates/root",
			Handler:           h.HandleDeleteRootCertificate,
			Description:       "Delete a root certificate",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/certificates/ca",
			Handler:      h.HandleGetIntermediateCertificate,
			Description:  "Get the intermediate certificate",
			AuthRequired: true,
		},
		{
			Method:            http.MethodPost,
			Path:              "/v1/certificates/ca",
			Handler:           h.HandleCreateIntermediateCertificate,
			Description:       "Create a new intermediate certificate",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
	}
}

func (h *ApiHandlers) HandleGetRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the root certificate")

	dbCertificates, diag := h.store.GetRootCertificateBySlug(ctx, config.RootCertificateSlug)
	if diag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error getting root certificate", diag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", diag)
		return
	}
	if dbCertificates == nil {
		ctx.Log().WithField("component", CertificateComponent).Error("Root certificate not found")
		api.WriteNotFound(w, r, "Root certificate not found")
		return
	}

	dtoModel := mappers.MapRootCertificateToDto(*dbCertificates)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtoModel)

	ctx.Log().WithField("component", CertificateComponent).Info("Root certificate retrieved successfully")
}

func (h *ApiHandlers) HandleCreateRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating a new root certificate")

	ctx.Log().Info("Checking if root certificate already exists")
	// checking if we already have a root certificate with us, we can only get one root certificate per database
	dbCertificate, dbCertDiag := h.store.GetRootCertificateBySlug(ctx, config.RootCertificateSlug)
	if dbCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error getting root certificate", dbCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", dbCertDiag)
		return
	}

	if dbCertificate != nil {
		ctx.Log().WithField("component", CertificateComponent).Info("Root certificate already exists")
		api.WriteErrorWithDiagnostics(w, r, http.StatusOK, errors.ErrorGettingRootCertificate, "Root certificate already exists", nil)
		return
	}

	ctx.Log().Info("Generating root certificate")
	rootCA, dbCertDiag := h.certificateService.GenerateRootCertificate(ctx)
	if dbCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error generating root certificate", dbCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingRootCertificate, "Error generating root certificate", dbCertDiag)
		return
	}

	ctx.Log().Info("Persisting root certificate")
	dbEntity := mappers.MapRootCertificateToEntity(*rootCA)
	createdEntity, dbCertDiag := h.store.CreateRootCertificate(ctx, &dbEntity)
	if dbCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error creating root certificate", dbCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingRootCertificate, "Error creating root certificate", dbCertDiag)
		return
	}

	result := mappers.MapRootCertificateToDto(*createdEntity)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	ctx.Log().WithField("component", CertificateComponent).Info("Root certificate created successfully")
}

func (h *ApiHandlers) HandleDeleteRootCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Deleting the root certificate")

	diag := h.store.DeleteRootCertificate(ctx, config.RootCertificateSlug)
	if diag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error deleting root certificate", diag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorDeletingRootCertificate, "Error deleting root certificate", diag)
		return
	}

	response := api.StatusResponse{
		ID:     config.RootCertificateSlug,
		Status: "deleted",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ctx.Log().Info("Root certificate deleted successfully")
}

func (h *ApiHandlers) HandleGetIntermediateCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the intermediate certificate")

	dbCertificates, diag := h.store.GetIntermediateCertificateBySlug(ctx, config.IntermediateCertificateSlug)
	if diag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error getting intermediate certificate", diag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingIntermediateCertificate, "Error getting intermediate certificate", diag)
		return
	}
	if dbCertificates == nil {
		ctx.Log().WithField("component", CertificateComponent).Error("Intermediate certificate not found")
		api.WriteNotFound(w, r, "Intermediate certificate not found")
		return
	}

	dtoModel := mappers.MapIntermediateCertificateToDto(*dbCertificates)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtoModel)

	ctx.Log().WithField("component", CertificateComponent).Info("Intermediate certificate retrieved successfully")
}

func (h *ApiHandlers) HandleCreateIntermediateCertificate(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating a new intermediate certificate")

	dbCertificates, diag := h.store.GetIntermediateCertificateBySlug(ctx, config.IntermediateCertificateSlug)
	if diag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error getting intermediate certificate", diag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingIntermediateCertificate, "Error getting intermediate certificate", diag)
		return
	}

	ctx.Log().Info("Checking if intermediate certificate already exists")
	if dbCertificates != nil {
		ctx.Log().WithField("component", CertificateComponent).Info("Intermediate certificate already exists")
		api.WriteErrorWithDiagnostics(w, r, http.StatusOK, errors.ErrorGettingIntermediateCertificate, "Intermediate certificate already exists", nil)
		return
	}

	ctx.Log().Info("Generating intermediate certificate")
	rootCA, dbCertDiag := h.store.GetRootCertificateBySlug(ctx, config.RootCertificateSlug)
	if dbCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error getting root certificate", dbCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", dbCertDiag)
		return
	}
	dtoRootCA := mappers.MapRootCertificateToDto(*rootCA)

	intermediateCA, intermediateCertDiag := h.certificateService.GenerateIntermediateCertificate(ctx, &dtoRootCA)
	if intermediateCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error generating intermediate certificate", intermediateCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingIntermediateCertificate, "Error generating intermediate certificate", intermediateCertDiag)
		return
	}

	ctx.Log().Info("Persisting intermediate certificate")
	dbEntity := mappers.MapIntermediateCertificateToEntity(*intermediateCA)
	createdEntity, createIntermediateCertDiag := h.store.CreateIntermediateCertificate(ctx, &dbEntity)
	if createIntermediateCertDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error creating intermediate certificate", createIntermediateCertDiag.Errors)
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorCreatingIntermediateCertificate, "Error creating intermediate certificate", createIntermediateCertDiag)
		return
	}

	result := mappers.MapIntermediateCertificateToDto(*createdEntity)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	ctx.Log().WithField("component", CertificateComponent).Info("Intermediate certificate created successfully")
}
