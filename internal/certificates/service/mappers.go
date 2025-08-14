package service

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
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
