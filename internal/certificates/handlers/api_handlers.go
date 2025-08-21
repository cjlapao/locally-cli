// Package handlers provides the API handlers for the certificates service
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/errors"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
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
			Path:        "/v1/certificates",
			Handler:     h.HandleGetCertificates,
			Description: "Get all certificates",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/certificates",
			Handler:     h.HandleCreateCertificate,
			Description: "Create a certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/root",
			Handler:     h.HandleGetRootCertificate,
			Description: "Get the root certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/ca",
			Handler:     h.HandleGetIntermediateCertificate,
			Description: "Get the intermediate certificate",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/type/{type}",
			Handler:     h.HandleGetCertificateByType,
			Description: "Get a certificate by type",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/certificates/{certificate_id}",
			Handler:     h.HandleGetCertificate,
			Description: "Get a certificate by id",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "certificates", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
	}
}

func (h *CertificatesApiHandlers) HandleGetCertificates(w http.ResponseWriter, r *http.Request) {
	diag := diagnostics.New("get_certificates_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting all certificates")
	pagination := utils.ParseQueryRequest(r)
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		diag.AddError(errors.ErrorMissingTenantID, "Tenant ID is missing", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, errors.ErrorMissingTenantID, "Tenant ID is missing", diag)
		return
	}

	certificates, diag := h.certificateService.GetCertificates(ctx, tenantID, pagination)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingCertificates, "Error getting certificates", diag)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(certificates)

	ctx.Log().Info("Certificates retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleCreateCertificate(w http.ResponseWriter, r *http.Request) {
	// diag := diagnostics.New("create_certificate_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating a certificate")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ctx.Log().Info("Certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleGetCertificate(w http.ResponseWriter, r *http.Request) {
	diag := diagnostics.New("get_certificate_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting a certificate")
	vars := mux.Vars(r)
	certificateID := vars["certificate_id"]
	if certificateID == "" {
		diag.AddError(errors.ErrorMissingCertificateID, "Certificate ID is missing", "certificates_handler", map[string]interface{}{
			"certificate_id": certificateID,
		})
	}
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		diag.AddError(errors.ErrorMissingTenantID, "Tenant ID is missing", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
		})
	}

	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, errors.ErrorMissingCertificateID, "Certificate ID is missing", diag)
		return
	}

	certificate, getDiag := h.certificateService.GetCertificateBy(ctx, tenantID, certificateID)
	if getDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingCertificate, "Error getting certificate", getDiag)
		return
	}
	if certificate == nil {
		diag.AddError(errors.ErrorGettingCertificate, "Certificate not found", "certificates_handler", map[string]interface{}{
			"tenant_id":      tenantID,
			"certificate_id": certificateID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusNotFound, errors.ErrorGettingCertificate, "Certificate not found", diag)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(certificate)

	ctx.Log().Info("Certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleGetRootCertificate(w http.ResponseWriter, r *http.Request) {
	diag := diagnostics.New("get_root_certificate_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the root certificate")

	certificate, getDiag := h.certificateService.GetCertificateBy(ctx, config.GlobalTenantID, config.GlobalRootCertificateID)
	if getDiag.HasErrors() {
		diag.AddError(errors.ErrorGettingRootCertificate, "Error getting root certificate", "certificates_handler", map[string]interface{}{
			"tenant_id": config.GlobalTenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingRootCertificate, "Error getting root certificate", diag)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certificate)

		ctx.Log().Info("Certificate retrieved successfully")
		return
	}

	if certificate == nil {
		diag.AddError(errors.ErrorGettingRootCertificate, "Root certificate not found", "certificates_handler", map[string]interface{}{
			"tenant_id": config.GlobalTenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusNotFound, errors.ErrorGettingRootCertificate, "Root certificate not found", diag)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(certificate)

	ctx.Log().Info("Root certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleGetCertificateByType(w http.ResponseWriter, r *http.Request) {
	diag := diagnostics.New("get_certificate_by_type_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting a certificate by type")
	vars := mux.Vars(r)
	pagination := utils.ParseQueryRequest(r)
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		diag.AddError(errors.ErrorMissingTenantID, "Tenant ID is missing", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, errors.ErrorMissingTenantID, "Tenant ID is missing", diag)
		return
	}
	typeStr := vars["type"]
	if typeStr == "" {
		diag.AddError(errors.ErrorMissingCertificateType, "Certificate type is missing", "certificates_handler", map[string]interface{}{
			"type": typeStr,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, errors.ErrorMissingCertificateType, "Certificate type is missing", diag)
		return
	}

	response, diag := h.certificateService.GetCertificatesByType(ctx, tenantID, types.CertificateType(typeStr), pagination)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingCertificates, "Error getting certificates", diag)
		return
	}

	if response == nil {
		diag.AddError(errors.ErrorGettingCertificatesByType, "Certificate not found", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
			"type":      typeStr,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusNotFound, errors.ErrorGettingCertificatesByType, "Certificate not found", diag)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ctx.Log().Info("Certificate retrieved successfully")
}

func (h *CertificatesApiHandlers) HandleGetIntermediateCertificate(w http.ResponseWriter, r *http.Request) {
	diag := diagnostics.New("get_intermediate_certificate_handler")
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting the intermediate certificate")
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		diag.AddError(errors.ErrorMissingTenantID, "Tenant ID is missing", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, errors.ErrorMissingTenantID, "Tenant ID is missing", diag)
		return
	}

	certificate, diag := h.certificateService.GetTenantIntermediateCertificate(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, errors.ErrorGettingIntermediateCertificate, "Error getting intermediate certificate", diag)
		return
	}
	if certificate == nil {
		diag.AddError(errors.ErrorGettingIntermediateCertificate, "Root certificate not found", "certificates_handler", map[string]interface{}{
			"tenant_id": tenantID,
		})
		api.WriteErrorWithDiagnostics(w, r, http.StatusNotFound, errors.ErrorGettingIntermediateCertificate, "Root certificate not found", diag)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(certificate)

	ctx.Log().Info("Root certificate retrieved successfully")
}
