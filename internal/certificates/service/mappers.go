package service

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/models"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
)

func mapCertificateConfigToModel(certificateConfig *entities.CertificateConfig) *models.CertificateConfig {
	return &models.CertificateConfig{
		CommonName:         certificateConfig.CommonName,
		Country:            certificateConfig.Country,
		State:              certificateConfig.State,
		Organization:       certificateConfig.Organization,
		OrganizationalUnit: certificateConfig.OrganizationalUnit,
		AdminEmailAddress:  certificateConfig.AdminEmailAddress,
		ExpiresInYears:     certificateConfig.ExpiresInYears,
		KeySize:            certificateConfig.KeySize,
		SignatureAlgorithm: certificateConfig.SignatureAlgorithm,
	}
}

func CertificateFromEntity(ctx *appctx.AppContext, certificate *entities.Certificate) interfaces.X509Certificate {
	switch certificate.Type {
	case pkg_types.CertificateTypeRoot:
		rootCert := NewEmptyX509RootCertificate()
		rootCert.FromDatabase(ctx, certificate)
		return rootCert
	case pkg_types.CertificateTypeIntermediate:
		rootCert := NewEmptyX509IntermediateCertificate()
		rootCert.FromDatabase(ctx, certificate)
		return rootCert
	case pkg_types.CertificateTypeCertificate:
		rootCert := NewEmptyX509ServerCertificate()
		rootCert.FromDatabase(ctx, certificate)
		return rootCert
	default:
		return nil
	}
}

func CertificateFromDto(ctx *appctx.AppContext, certificate *models.Certificate) interfaces.X509Certificate {
	dbEntity := mappers.MapCertificateToEntity(*certificate)
	return CertificateFromEntity(ctx, &dbEntity)
}
