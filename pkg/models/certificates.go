package models

import (
	"github.com/cjlapao/locally-cli/pkg/types"
)

type RootCertificate struct {
	BaseModel
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
	BaseModel
	RootCertificateID string             `json:"root_certificate_id" yaml:"root_certificate_id"`
	Name              string             `json:"name" yaml:"name"`
	Slug              string             `json:"slug" yaml:"slug"`
	Config            *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate    string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey     string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Csr               string             `json:"Csr" yaml:"Csr"`
	Certificates      []*Certificate     `json:"certificates,omitempty" yaml:"certificates,omitempty"`
}

func (i *IntermediateCertificate) TableName() string {
	return "intermediate_certificates"
}

type Certificate struct {
	BaseModel
	RootCertificateID         string             `json:"root_certificate_id" yaml:"root_certificate_id"`
	IntermediateCertificateID string             `json:"intermediate_certificate_id" yaml:"intermediate_certificate_id"`
	Name                      string             `json:"name" yaml:"name"`
	Slug                      string             `json:"slug" yaml:"slug"`
	Config                    *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate            string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey             string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Csr                       string             `json:"Csr" yaml:"Csr"`
}

func (c *Certificate) TableName() string {
	return "certificates"
}

type CertificateConfig struct {
	Country            string                   `json:"country" yaml:"country"`
	State              string                   `json:"state" yaml:"state"`
	Organization       string                   `json:"organization" yaml:"organization"`
	CommonName         string                   `json:"commonName" yaml:"commonName"`
	City               string                   `json:"city" yaml:"city"`
	FQDNs              []string                 `json:"fqdns" yaml:"fqdns"`
	IpAddresses        []string                 `json:"ipAddresses" yaml:"ipAddresses"`
	OrganizationalUnit string                   `json:"organizationalUnit" yaml:"organizationalUnit"`
	AdminEmailAddress  string                   `json:"adminEmailAddress" yaml:"adminEmailAddress"`
	ExpiresInYears     int                      `json:"expiresInYears" yaml:"expiresInYears"`
	KeySize            types.CertificateKeySize `json:"keySize" yaml:"keySize"`
	SignatureAlgorithm types.SignatureAlgorithm `json:"signatureAlgorithm" yaml:"signatureAlgorithm"`
	Password           string                   `json:"password" yaml:"password"`
}
