package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapCertificateConfigToDto(config entities.CertificateConfig) models.CertificateConfig {
	result := models.CertificateConfig{
		BaseModelWithTenant: *MapBaseModelWithTenantToDto(&config.BaseModelWithTenant),
		CertificateType:     config.CertificateType,
		CertificateID:       config.CertificateID,
		Country:             config.Country,
		State:               config.State,
		Organization:        config.Organization,
		CommonName:          config.CommonName,
		City:                config.City,
		OrganizationalUnit:  config.OrganizationalUnit,
		AdminEmailAddress:   config.AdminEmailAddress,
		FQDNs:               config.FQDNs,
		IpAddresses:         config.IpAddresses,
		ExpiresInYears:      config.ExpiresInYears,
		KeySize:             config.KeySize,
		SignatureAlgorithm:  config.SignatureAlgorithm,
		Password:            config.Password,
	}

	return result
}

func MapCertificateToDto(certificate entities.Certificate) models.Certificate {
	return models.Certificate{
		BaseModelWithTenant: *MapBaseModelWithTenantToDto(&certificate.BaseModelWithTenant),
		Name:                certificate.Name,
		Type:                certificate.Type,
		Config:              MapCertificateConfigToDto(certificate.Config),
		PemCertificate:      certificate.PemCertificate,
		PemPrivateKey:       certificate.PemPrivateKey,
		Csr:                 certificate.PemCsr,
	}
}
