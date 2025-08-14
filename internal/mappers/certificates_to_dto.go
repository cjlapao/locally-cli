package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapCertificateConfigToDto(config entities.CertificateConfig) models.CertificateConfig {
	return models.CertificateConfig{
		Country:            config.Country,
		State:              config.State,
		Organization:       config.Organization,
		CommonName:         config.CommonName,
		City:               config.City,
		OrganizationalUnit: config.OrganizationalUnit,
		AdminEmailAddress:  config.AdminEmailAddress,
		FQDNs:              config.FQDNs,
		IpAddresses:        config.IpAddresses,
		ExpiresInYears:     config.ExpiresInYears,
		KeySize:            config.KeySize,
		SignatureAlgorithm: config.SignatureAlgorithm,
		Password:           config.Password,
	}
}
