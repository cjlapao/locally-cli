// Package interfaces provides the interfaces for the certificates.
package interfaces

import (
	"crypto/rsa"
	"crypto/x509"

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
	GenerateRootCertificate(ctx *appctx.AppContext) (X509Certificate, *diagnostics.Diagnostics)
	GenerateIntermediateCertificate(ctx *appctx.AppContext, rootCA X509Certificate) (X509Certificate, *diagnostics.Diagnostics)
	GetCertificate(ctx *appctx.AppContext, slugOrId string) (X509Certificate, *diagnostics.Diagnostics)
	GetName() string
}
