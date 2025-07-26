package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	db_types "github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapRootCertificateToDto(rootCertificate entities.RootCertificate) models.RootCertificate {
	result := models.RootCertificate{
		Name:           rootCertificate.Name,
		Slug:           rootCertificate.Slug,
		PemCertificate: rootCertificate.PemCertificate,
		PemPrivateKey:  rootCertificate.PemPrivateKey,
		Csr:            rootCertificate.CertificateSigningRequest,
	}
	// Map the base model
	result.ID = rootCertificate.ID
	result.CreatedAt = rootCertificate.CreatedAt
	result.UpdatedAt = rootCertificate.UpdatedAt

	// Map the config
	config := MapCertificateConfigToDto(rootCertificate.Config)
	result.Config = &config

	// Map the intermediate certificates
	if len(rootCertificate.IntermediateCertificates) > 0 {
		intermediateCertificates := []*models.IntermediateCertificate{}
		for _, intermediateCertificate := range rootCertificate.IntermediateCertificates {
			ca := MapIntermediateCertificateToDto(intermediateCertificate)
			intermediateCertificates = append(intermediateCertificates, &ca)
		}
		result.IntermediateCertificates = intermediateCertificates
	} else {
		result.IntermediateCertificates = []*models.IntermediateCertificate{}
	}

	return result
}

func MapIntermediateCertificateToDto(intermediateCertificate entities.IntermediateCertificate) models.IntermediateCertificate {
	result := models.IntermediateCertificate{
		RootCertificateID: intermediateCertificate.RootCertificateID,
		Name:              intermediateCertificate.Name,
		Slug:              intermediateCertificate.Slug,
		PemCertificate:    intermediateCertificate.PemCertificate,
		PemPrivateKey:     intermediateCertificate.PemPrivateKey,
		Csr:               intermediateCertificate.CertificateSigningRequest,
	}

	// Map the base model
	result.ID = intermediateCertificate.ID
	result.CreatedAt = intermediateCertificate.CreatedAt
	result.UpdatedAt = intermediateCertificate.UpdatedAt

	// Map the config
	config := MapCertificateConfigToDto(intermediateCertificate.Config)
	result.Config = &config

	return result
}

func MapCertificateConfigToDto(config db_types.JSONObject[entities.CertificateConfig]) models.CertificateConfig {
	dbConfig := config.Get()
	return models.CertificateConfig{
		Country:            dbConfig.Country,
		State:              dbConfig.State,
		Organization:       dbConfig.Organization,
		CommonName:         dbConfig.CommonName,
		City:               dbConfig.City,
		OrganizationalUnit: dbConfig.OrganizationalUnit,
		AdminEmailAddress:  dbConfig.AdminEmailAddress,
		FQDNs:              dbConfig.FQDNs,
		IpAddresses:        dbConfig.IpAddresses,
		ExpiresInYears:     dbConfig.ExpiresInYears,
		KeySize:            dbConfig.KeySize,
		SignatureAlgorithm: dbConfig.SignatureAlgorithm,
		Password:           dbConfig.Password,
	}
}
