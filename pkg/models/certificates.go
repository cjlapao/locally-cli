package models

import (
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/types"
)

type RootCertificate struct {
	BaseModelWithTenant
	Name                     string                     `json:"name" yaml:"name"`
	Slug                     string                     `json:"slug" yaml:"slug"`
	Config                   *CertificateConfig         `json:"config" yaml:"config"`
	PemCertificate           string                     `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey            string                     `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Csr                      string                     `json:"Csr" yaml:"Csr"`
	IntermediateCertificates []*IntermediateCertificate `json:"intermediateCertificates,omitempty" yaml:"intermediateCertificates,omitempty"`
}

func (r *RootCertificate) TableName() string {
	return "root_certificates"
}

type IntermediateCertificate struct {
	BaseModelWithTenant
	RootCertificateID string             `json:"root_certificate_id" yaml:"root_certificate_id"`
	Name              string             `json:"name" yaml:"name"`
	Slug              string             `json:"slug" yaml:"slug"`
	Config            *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate    string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey     string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Csr               string             `json:"Csr" yaml:"Csr"`
	Certificates      []*Certificate     `json:"certificates,omitempty" yaml:"certificates,omitempty"`
}

type Certificate struct {
	BaseModelWithTenant
	RootCertificateID         string             `json:"root_certificate_id,omitempty" yaml:"root_certificate_id,omitempty"`
	IntermediateCertificateID string             `json:"intermediate_certificate_id,omitempty" yaml:"intermediate_certificate_id,omitempty"`
	Name                      string             `json:"name" yaml:"name"`
	Slug                      string             `json:"slug" yaml:"slug"`
	Config                    *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate            string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey             string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Csr                       string             `json:"Csr" yaml:"Csr"`
}

type CertificateConfig struct {
	Country                   string                   `json:"country" yaml:"country"`
	State                     string                   `json:"state" yaml:"state"`
	Organization              string                   `json:"organization" yaml:"organization"`
	CommonName                string                   `json:"commonName" yaml:"commonName"`
	City                      string                   `json:"city" yaml:"city"`
	FQDNs                     []string                 `json:"fqdns" yaml:"fqdns"`
	IpAddresses               []string                 `json:"ipAddresses" yaml:"ipAddresses"`
	OrganizationalUnit        string                   `json:"organizationalUnit" yaml:"organizationalUnit"`
	AdminEmailAddress         string                   `json:"adminEmailAddress" yaml:"adminEmailAddress"`
	ExpiresInYears            int                      `json:"expiresInYears" yaml:"expiresInYears"`
	KeySize                   types.CertificateKeySize `json:"keySize" yaml:"keySize"`
	SignatureAlgorithm        types.SignatureAlgorithm `json:"signatureAlgorithm" yaml:"signatureAlgorithm"`
	Password                  string                   `json:"password" yaml:"password"`
	RootCertificateID         string                   `json:"rootCertificateId" yaml:"rootCertificateId"`
	IntermediateCertificateID string                   `json:"intermediateCertificateId" yaml:"intermediateCertificateId"`
}

func (c *CertificateConfig) Validate() *diagnostics.Diagnostics {
	diag := diagnostics.New("validate_certificate_config")
	moduleName := "certificate_config"
	if c.Country == "" {
		diag.AddError("country_required", "country is required", moduleName, nil)
	}
	if c.State == "" {
		diag.AddError("state_required", "state is required", moduleName, nil)
	}
	if c.Organization == "" {
		diag.AddError("organization_required", "organization is required", moduleName, nil)
	}
	if c.CommonName == "" {
		diag.AddError("common_name_required", "common name is required", moduleName, nil)
	}
	if c.KeySize == 0 {
		diag.AddError("key_size_required", "key size is required", moduleName, nil)
	}
	if c.SignatureAlgorithm == 0 {
		diag.AddError("signature_algorithm_required", "signature algorithm is required", moduleName, nil)
	}
	if c.ExpiresInYears == 0 {
		diag.AddError("expires_in_years_required", "expires in years is required", moduleName, nil)
	}
	if c.Password == "" {
		diag.AddError("password_required", "password is required", moduleName, nil)
	}
	if c.RootCertificateID == "" {
		diag.AddError("root_certificate_id_required", "root certificate id is required", moduleName, nil)
	}
	if c.IntermediateCertificateID == "" {
		diag.AddError("intermediate_certificate_id_required", "intermediate certificate id is required", moduleName, nil)
	}
	return diag
}
