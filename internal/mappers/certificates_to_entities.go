package mappers

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/google/uuid"
)

func MapX509CertificateToEntity(certificate interfaces.X509Certificate) entities.Certificate {
	result := entities.Certificate{
		Name:           certificate.GetName(),
		PemCertificate: string(certificate.GetPemCertificate()),
		PemPrivateKey:  string(certificate.GetPemPrivateKey()),
		PemCsr:         string(certificate.GetPemCsr()),
	}

	result.Slug = certificate.GetSlug()
	result.ID = uuid.New().String()
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()

	config := certificate.GetConfiguration()
	result.Config = MapCertificateConfigToEntity(*config)

	return result
}

func MapCertificateToEntity(certificate models.Certificate) entities.Certificate {
	result := entities.Certificate{
		Type:           certificate.Type,
		PemCsr:         certificate.Csr,
		Name:           certificate.Name,
		PemCertificate: certificate.PemCertificate,
		PemPrivateKey:  certificate.PemPrivateKey,
	}

	result.Slug = certificate.Slug
	result.ID = certificate.ID
	result.CreatedAt = certificate.CreatedAt
	result.UpdatedAt = certificate.UpdatedAt

	if certificate.Config.ID != "" {
		config := MapCertificateConfigToEntity(certificate.Config)
		result.Config = config
	}

	return result
}

func MapCertificateConfigToEntity(config models.CertificateConfig) entities.CertificateConfig {
	return entities.CertificateConfig{
		CertificateType:           config.CertificateType,
		CertificateID:             config.CertificateID,
		Country:                   config.Country,
		State:                     config.State,
		Organization:              config.Organization,
		CommonName:                config.CommonName,
		City:                      config.City,
		OrganizationalUnit:        config.OrganizationalUnit,
		AdminEmailAddress:         config.AdminEmailAddress,
		FQDNs:                     config.FQDNs,
		IpAddresses:               config.IpAddresses,
		ExpiresInYears:            config.ExpiresInYears,
		KeySize:                   config.KeySize,
		SignatureAlgorithm:        config.SignatureAlgorithm,
		Password:                  config.Password,
		RootCertificateID:         config.RootCertificateID,
		IntermediateCertificateID: config.IntermediateCertificateID,
	}
}
