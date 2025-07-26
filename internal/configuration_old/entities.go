package configuration

import (
	"crypto/x509"

	"github.com/cjlapao/locally-cli/internal/context"
	"github.com/cjlapao/locally-cli/internal/context/entities"
)

type GlobalConfiguration struct {
	format               string
	verbose              bool
	CurrentContext       string                      `json:"currentContext" yaml:"currentContext"`
	Contexts             []*context.Context          `json:"contexts,omitempty" yaml:"contexts,omitempty"`
	Tools                *entities.Tools             `json:"tools,omitempty" yaml:"tools,omitempty"`
	Network              *Network                    `json:"network,omitempty" yaml:"network,omitempty"`
	CertificateGenerator *CertificateGeneratorConfig `json:"certificateGenerator,omitempty" yaml:"certificateGenerator,omitempty"`
	Cors                 *Cors                       `json:"cors,omitempty" yaml:"cors,omitempty"`
}

type Cors struct {
	AllowedMethods string   `json:"allowedMethods" yaml:"allowedMethods"`
	AllowedHeaders string   `json:"allowedHeaders" yaml:"allowedHeaders"`
	AllowedOrigins []string `json:"allowedOrigins" yaml:"allowedOrigins"`
}

type Network struct {
	LocalIP        string `json:"localIp,omitempty" yaml:"localIp,omitempty"`
	DomainName     string `json:"domainName,omitempty" yaml:"domainName,omitempty"`
	CERTPath       string `json:"certPath,omitempty" yaml:"certPath,omitempty"`
	PrivateKeyPath string `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
}

type CertificateGeneratorConfig struct {
	Root         []*RootCertificate `json:"rootCertificates" yaml:"rootCertificates"`
	OutputToFile bool               `json:"outputToFile" yaml:"outputToFile"`
}

type RootCertificate struct {
	Name                     string                     `json:"name" yaml:"name"`
	Config                   *CertificateConfig         `json:"config" yaml:"config"`
	PemCertificate           string                     `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey            string                     `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	IntermediateCertificates []*IntermediateCertificate `json:"intermediateCertificates" yaml:"intermediateCertificates"`
}

type IntermediateCertificate struct {
	Name           string             `json:"name" yaml:"name"`
	Config         *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey  string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
	Certificates   []*Certificate     `json:"certificates" yaml:"certificates"`
}

type Certificate struct {
	Name           string             `json:"name" yaml:"name"`
	Config         *CertificateConfig `json:"config" yaml:"config"`
	PemCertificate string             `json:"PemCertificate" yaml:"PemCertificate"`
	PemPrivateKey  string             `json:"PemPrivateKey" yaml:"PemPrivateKey"`
}

type CertificateConfig struct {
	Country            string             `json:"country" yaml:"country"`
	State              string             `json:"state" yaml:"state"`
	Organization       string             `json:"organization" yaml:"organization"`
	CommonName         string             `json:"commonName" yaml:"commonName"`
	City               string             `json:"city" yaml:"city"`
	FQDNs              []string           `json:"fqdns" yaml:"fqdns"`
	IpAddresses        []string           `json:"ipAddresses" yaml:"ipAddresses"`
	OrganizationalUnit string             `json:"organizationalUnit" yaml:"organizationalUnit"`
	AdminEmailAddress  string             `json:"adminEmailAddress" yaml:"adminEmailAddress"`
	ExpiresInYears     int                `json:"expiresInYears" yaml:"expiresInYears"`
	KeySize            CertificateKeySize `json:"keySize" yaml:"keySize"`
	SignatureAlgorithm SignatureAlgorithm `json:"signatureAlgorithm" yaml:"signatureAlgorithm"`
	Password           string             `json:"password" yaml:"password"`
}

type CertificateKeySize int

const (
	Key2048 CertificateKeySize = 2048
	Key4096 CertificateKeySize = 4096
	Key8192 CertificateKeySize = 8192
)

type SignatureAlgorithm int

const (
	SHA256 SignatureAlgorithm = 4
	SHA384 SignatureAlgorithm = 5
	SHA512 SignatureAlgorithm = 6
)

func (s SignatureAlgorithm) ToX509SignatureAlgorithm() x509.SignatureAlgorithm {
	switch s {
	case 4:
		return x509.SHA256WithRSA
	case 5:
		return x509.SHA384WithRSA
	case 6:
		return x509.SHA512WithRSA
	default:
		return x509.SHA256WithRSA
	}
}
