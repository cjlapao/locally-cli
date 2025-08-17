// Package certificates provides a service for generating and managing certificates
package certificates

import (
	"github.com/cjlapao/locally-cli/internal/certificates/handlers"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/certificates/service"
	"github.com/cjlapao/locally-cli/internal/database/stores"
)

// Initialize initializes the claim service
func Initialize(certificatesStore stores.CertificatesDataStoreInterface,
	tenantStore stores.TenantDataStoreInterface,
) interfaces.CertificateServiceInterface {
	return service.Initialize(certificatesStore, tenantStore)
}

// GetInstance returns the claim service instance
func GetInstance() interfaces.CertificateServiceInterface {
	return service.GetInstance()
}

// Reset resets the claim service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new claims API handler
func NewApiHandler(certificateService interfaces.CertificateServiceInterface) *handlers.CertificatesApiHandlers {
	return handlers.NewCertificatesApiHandler(certificateService)
}
