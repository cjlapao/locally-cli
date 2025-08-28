// Package models contains the models for the certificates
package models

import "github.com/cjlapao/locally-cli/pkg/types"

type CreateCertificateRequest struct {
	CertificateType    types.CertificateType    `json:"certificate_type" yaml:"certificate_type"`
	SubDomain          string                   `json:"subDomain" yaml:"subDomain" validate:"required"`
	Country            string                   `json:"country" yaml:"country"`
	State              string                   `json:"state" yaml:"state"`
	Organization       string                   `json:"organization" yaml:"organization"`
	CommonName         string                   `json:"commonName" yaml:"commonName"`
	City               string                   `json:"city" yaml:"city"`
	FQDNs              []string                 `json:"fqdns" yaml:"fqdns"`
	IpAddresses        []string                 `json:"ip_addresses" yaml:"ip_addresses"`
	OrganizationalUnit string                   `json:"organizational_unit" yaml:"organizational_unit"`
	AdminEmailAddress  string                   `json:"admin_email_address" yaml:"admin_email_address"`
	ExpiresInYears     int                      `json:"expires_in_years" yaml:"expires_in_years" validate:"required"`
	KeySize            types.CertificateKeySize `json:"key_size" yaml:"key_size" validate:"required"`
	SignatureAlgorithm types.SignatureAlgorithm `json:"signature_algorithm" yaml:"signature_algorithm" validate:"required"`
	Password           string                   `json:"password" yaml:"password"`
}
