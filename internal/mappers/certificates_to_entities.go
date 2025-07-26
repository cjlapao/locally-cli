package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	db_types "github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapRootCertificateToEntity(rootCA models.RootCertificate) entities.RootCertificate {
	result := entities.RootCertificate{
		Name:                      rootCA.Name,
		PemCertificate:            rootCA.PemCertificate,
		PemPrivateKey:             rootCA.PemPrivateKey,
		CertificateSigningRequest: rootCA.Csr,
		IntermediateCertificates:  make([]entities.IntermediateCertificate, 0),
	}

	result.Slug = rootCA.Slug
	result.ID = rootCA.ID
	result.CreatedAt = rootCA.CreatedAt
	result.UpdatedAt = rootCA.UpdatedAt

	// Map the config to the entity
	config := MapCertificateConfigToEntity(*rootCA.Config)
	jsonObj := db_types.JSONObject[entities.CertificateConfig]{}
	jsonObj.Set(config)

	result.Config = jsonObj

	return result
}

func MapIntermediateCertificateToEntity(intermediateCertificate models.IntermediateCertificate) entities.IntermediateCertificate {
	result := entities.IntermediateCertificate{
		RootCertificateID:         intermediateCertificate.RootCertificateID,
		CertificateSigningRequest: intermediateCertificate.Csr,
		Name:                      intermediateCertificate.Name,
		PemCertificate:            intermediateCertificate.PemCertificate,
		PemPrivateKey:             intermediateCertificate.PemPrivateKey,
		Certificates:              make([]entities.Certificate, 0),
	}

	result.Slug = intermediateCertificate.Slug
	result.ID = intermediateCertificate.ID
	result.CreatedAt = intermediateCertificate.CreatedAt
	result.UpdatedAt = intermediateCertificate.UpdatedAt

	// Map the config to the entity
	config := MapCertificateConfigToEntity(*intermediateCertificate.Config)
	jsonObj := db_types.JSONObject[entities.CertificateConfig]{}
	jsonObj.Set(config)

	result.Config = jsonObj

	return result
}

func MapCertificateToEntity(certificate models.Certificate) entities.Certificate {
	result := entities.Certificate{
		RootCertificateID:         certificate.RootCertificateID,
		IntermediateCertificateID: certificate.IntermediateCertificateID,
		CertificateSigningRequest: certificate.Csr,
		Name:                      certificate.Name,
		PemCertificate:            certificate.PemCertificate,
		PemPrivateKey:             certificate.PemPrivateKey,
	}

	result.Slug = certificate.Slug
	result.ID = certificate.ID
	result.CreatedAt = certificate.CreatedAt
	result.UpdatedAt = certificate.UpdatedAt

	config := MapCertificateConfigToEntity(*certificate.Config)
	jsonObj := db_types.JSONObject[entities.CertificateConfig]{}
	jsonObj.Set(config)

	result.Config = jsonObj

	return result
}

func MapCertificateConfigToEntity(config models.CertificateConfig) entities.CertificateConfig {
	return entities.CertificateConfig{
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
