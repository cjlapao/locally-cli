package entities

import (
	"github.com/cjlapao/locally-cli/internal/database/types"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
)

type RootCertificate struct {
	BaseModel
	Name                      string                              `json:"name" yaml:"name" gorm:"column:name;type:varchar(255);not null"`
	Config                    types.JSONObject[CertificateConfig] `json:"config" yaml:"config" gorm:"column:config;type:json;not null"`
	PemCertificate            string                              `json:"pemCertificate" yaml:"pemCertificate" gorm:"column:pem_certificate;type:text;not null"`
	PemPrivateKey             string                              `json:"pemPrivateKey" yaml:"pemPrivateKey" gorm:"column:pem_private_key;type:text;not null"`
	CertificateSigningRequest string                              `json:"certificateSigningRequest" yaml:"certificateSigningRequest" gorm:"column:certificate_signing_request;type:text;not null"`
	IntermediateCertificates  []IntermediateCertificate           `json:"intermediateCertificates" yaml:"intermediateCertificates" gorm:"foreignKey:RootCertificateID;references:ID;constraint:OnDelete:CASCADE"`
}

func (r *RootCertificate) TableName() string {
	return "root_certificates"
}

type IntermediateCertificate struct {
	BaseModel
	RootCertificateID         string                              `json:"rootCertificateId" yaml:"rootCertificateId" gorm:"column:root_certificate_id;type:uuid;not null"`
	Name                      string                              `json:"name" yaml:"name" gorm:"column:name;type:varchar(255);not null"`
	Config                    types.JSONObject[CertificateConfig] `json:"config" yaml:"config" gorm:"column:config;type:json;not null"`
	PemCertificate            string                              `json:"pemCertificate" yaml:"pemCertificate" gorm:"column:pem_certificate;type:text;not null"`
	PemPrivateKey             string                              `json:"pemPrivateKey" yaml:"pemPrivateKey" gorm:"column:pem_private_key;type:text;not null"`
	CertificateSigningRequest string                              `json:"certificateSigningRequest" yaml:"certificateSigningRequest" gorm:"column:certificate_signing_request;type:text;not null"`
	Certificates              []Certificate                       `json:"certificates" yaml:"certificates" gorm:"foreignKey:IntermediateCertificateID;references:ID;constraint:OnDelete:CASCADE"`
}

func (i *IntermediateCertificate) TableName() string {
	return "intermediate_certificates"
}

type Certificate struct {
	BaseModelWithTenant
	Type                    pkg_types.CertificateType `json:"type" yaml:"type" gorm:"column:type;type:varchar(255);not null"`
	Name                    string                    `json:"name" yaml:"name" gorm:"column:name;type:varchar(255);not null"`
	Config                  CertificateConfig         `json:"config" yaml:"config" gorm:"foreignKey:CertificateID;references:ID;constraint:OnDelete:CASCADE"`
	PemCertificate          string                    `json:"pem_certificate" yaml:"pem_certificate" gorm:"column:pem_certificate;type:text;not null;unique"`
	PemPrivateKey           string                    `json:"pem_private_key" yaml:"pem_private_key" gorm:"column:pem_private_key;type:text;not null;unique"`
	PemCsr                  string                    `json:"pem_csr" yaml:"pem_csr" gorm:"column:pem_csr;type:text;not null;unique"`
	RootCertificate         *Certificate              `json:"root_certificate" yaml:"root_certificate" gorm:"-"`
	IntermediateCertificate *Certificate              `json:"intermediate_certificate" yaml:"intermediate_certificate" gorm:"-"`
}

func (c *Certificate) TableName() string {
	return "certificates"
}

type CertificateConfig struct {
	BaseModelWithTenant
	CertificateID             string                       `json:"certificate_id" yaml:"certificate_id" gorm:"column:certificate_id;type:uuid;not null;unique"`
	Country                   string                       `json:"country" yaml:"country" gorm:"column:country;type:varchar(255);not null"`
	State                     string                       `json:"state" yaml:"state" gorm:"column:state;type:varchar(255);not null"`
	Organization              string                       `json:"organization" yaml:"organization" gorm:"column:organization;type:varchar(255);not null"`
	CommonName                string                       `json:"commonName" yaml:"commonName" gorm:"column:common_name;type:varchar(255);not null"`
	City                      string                       `json:"city" yaml:"city" gorm:"column:city;type:varchar(255);not null"`
	FQDNs                     types.StringSlice            `json:"fqdns" yaml:"fqdns" gorm:"column:fqdns;type:json;not null;default:[]"`
	IpAddresses               types.StringSlice            `json:"ipAddresses" yaml:"ipAddresses" gorm:"column:ip_addresses;type:json;not null;default:[]"`
	OrganizationalUnit        string                       `json:"organizationalUnit" yaml:"organizationalUnit" gorm:"column:organizational_unit;type:varchar(255);not null"`
	AdminEmailAddress         string                       `json:"adminEmailAddress" yaml:"adminEmailAddress" gorm:"column:admin_email_address;type:varchar(255);not null"`
	ExpiresInYears            int                          `json:"expiresInYears" yaml:"expiresInYears" gorm:"column:expires_in_years;type:int;not null"`
	KeySize                   pkg_types.CertificateKeySize `json:"keySize" yaml:"keySize" gorm:"column:key_size;type:int;not null"`
	SignatureAlgorithm        pkg_types.SignatureAlgorithm `json:"signatureAlgorithm" yaml:"signatureAlgorithm" gorm:"column:signature_algorithm;type:int;not null"`
	Password                  string                       `json:"password" yaml:"password" gorm:"column:password;type:varchar(255);not null"`
	RootCertificateID         string                       `json:"root_certificate_id" yaml:"root_certificate_id" gorm:"column:root_certificate_id;type:uuid"`
	IntermediateCertificateID string                       `json:"intermediate_certificate_id" yaml:"intermediate_certificate_id" gorm:"column:intermediate_certificate_id;type:uuid"`
}

func (c *CertificateConfig) TableName() string {
	return "certificate_configs"
}
