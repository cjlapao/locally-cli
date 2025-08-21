// Package interfaces provides the interfaces for the certificates.
package interfaces

import (
	"crypto/rsa"
	"crypto/x509"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
)

type X509Certificate interface {
	GetType() pkg_types.CertificateType
	GetName() string
	GetSlug() string
	GetCertificateFileName() string
	GetPrivateKeyFileName() string
	GetCertificateRequestFileName() string
	Generate(ctx *appctx.AppContext) (X509Certificate, *diagnostics.Diagnostics)
	LoadFromFile(ctx *appctx.AppContext) *diagnostics.Diagnostics
	SaveToFile(ctx *appctx.AppContext) *diagnostics.Diagnostics
	Parse(ctx *appctx.AppContext, certificate string, privateKey string) *diagnostics.Diagnostics
	Install(ctx *appctx.AppContext) *diagnostics.Diagnostics
	GetPrivateKey() *rsa.PrivateKey
	GetCertificate() *x509.Certificate
	GetConfiguration() *pkg_models.CertificateConfig
	GetPemCertificate() []byte
	GetPemPrivateKey() []byte
	GetPemCsr() []byte
	FromDatabase(ctx *appctx.AppContext, certificate *entities.Certificate) *diagnostics.Diagnostics
}

type CertificateServiceInterface interface {
	GenerateX509RootCertificate(ctx *appctx.AppContext) (X509Certificate, *diagnostics.Diagnostics)
	GenerateX509IntermediateCertificate(ctx *appctx.AppContext, tenantID string) (X509Certificate, *diagnostics.Diagnostics)
	GenerateX509Certificate(ctx *appctx.AppContext, tenantID string, certificateConfig pkg_models.CertificateConfig) (X509Certificate, *diagnostics.Diagnostics)
	GetX509Certificate(ctx *appctx.AppContext, tenantID string, slugOrId string) (X509Certificate, *diagnostics.Diagnostics)
	GetRootCertificate(ctx *appctx.AppContext) (*pkg_models.Certificate, *diagnostics.Diagnostics)
	GetIntermediateCertificate(ctx *appctx.AppContext, tenantId string, slug string) (*pkg_models.Certificate, *diagnostics.Diagnostics)
	GetTenantIntermediateCertificate(ctx *appctx.AppContext, tenantID string) (*pkg_models.Certificate, *diagnostics.Diagnostics)
	GetCertificates(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Certificate], *diagnostics.Diagnostics)
	GetCertificatesByType(ctx *appctx.AppContext, tenantID string, certType pkg_types.CertificateType, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Certificate], *diagnostics.Diagnostics)
	GetCertificateBy(ctx *appctx.AppContext, tenantID string, slugOrId string) (*pkg_models.Certificate, *diagnostics.Diagnostics)
	CreateCertificate(ctx *appctx.AppContext, tenantID string, certType pkg_types.CertificateType, certificateConfig pkg_models.CertificateConfig) (*pkg_models.Certificate, *diagnostics.Diagnostics)
	GetName() string
}
