package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/certificates/constants"
	"github.com/cjlapao/locally-cli/internal/certificates/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"

	"github.com/cjlapao/common-go/helper"
)

type X509ServerCertificate struct {
	prefix         string
	name           string
	slug           string
	privateKey     *rsa.PrivateKey
	certificate    *x509.Certificate
	configuration  models.CertificateConfig
	rootCA         interfaces.X509Certificate
	intermediateCA interfaces.X509Certificate
	pem            []byte
	csr            []byte
	privateKeyPem  []byte
}

func NewX509ServerCertificate(name string, rootCA interfaces.X509Certificate, intermediateCA interfaces.X509Certificate, certConfig models.CertificateConfig) interfaces.X509Certificate {
	cert := &X509ServerCertificate{
		prefix:         "certificate",
		name:           name,
		rootCA:         rootCA,
		intermediateCA: intermediateCA,
		configuration:  certConfig,
	}
	cert.slug = utils.Slugify(name)

	return cert
}

func NewEmptyX509ServerCertificate() interfaces.X509Certificate {
	cert := &X509ServerCertificate{
		prefix:         "certificate",
		name:           "unknown",
		rootCA:         nil,
		intermediateCA: nil,
		configuration: models.CertificateConfig{
			FQDNs:       []string{},
			IpAddresses: []string{},
		},
	}
	cert.slug = utils.Slugify("unknown")
	return cert
}

func (c *X509ServerCertificate) GetType() pkg_types.CertificateType {
	return pkg_types.CertificateTypeCertificate
}

func (c *X509ServerCertificate) GetName() string {
	return c.name
}

func (c *X509ServerCertificate) GetSlug() string {
	return c.slug
}

func (c *X509ServerCertificate) GetCertificateFileName() string {
	certificateFileName := c.prefix + "_" + c.baseFileName() + ".crt"
	return certificateFileName
}

func (c *X509ServerCertificate) GetPrivateKeyFileName() string {
	privateKeyFileName := c.prefix + "_" + c.baseFileName() + ".key"
	return privateKeyFileName
}

func (c *X509ServerCertificate) GetCertificateRequestFileName() string {
	certificateRequestFileName := c.prefix + "_" + c.baseFileName() + ".csr"
	return certificateRequestFileName
}

func (c *X509ServerCertificate) GetCertificatePfxFileName() string {
	certificatePfxFileName := c.prefix + "_" + c.baseFileName() + ".pfx"
	return certificatePfxFileName
}

func (c *X509ServerCertificate) Generate(ctx *appctx.AppContext) (interfaces.X509Certificate, *diagnostics.Diagnostics) {
	ctx.Log().Debug("Starting to generate server certificate certificate")

	subject := pkix.Name{
		Country:            []string{c.configuration.Country},
		Organization:       []string{c.configuration.Organization},
		OrganizationalUnit: []string{c.configuration.OrganizationalUnit},
		Province:           []string{c.configuration.State},
		Locality:           []string{c.configuration.City},
		CommonName:         c.configuration.CommonName,
	}

	if c.configuration.AdminEmailAddress != "" {
		subject.ExtraNames = []pkix.AttributeTypeAndValue{
			{
				Type: constants.OidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(c.configuration.AdminEmailAddress),
				},
			},
		}
	}

	serverCertificateTemplate := x509.Certificate{
		SerialNumber:       generateSerialNumber(),
		Subject:            subject,
		NotBefore:          time.Now().Add(-10 * time.Second),
		NotAfter:           time.Now().AddDate(c.configuration.ExpiresInYears, 0, 0),
		IsCA:               false,
		SignatureAlgorithm: c.configuration.SignatureAlgorithm.ToX509SignatureAlgorithm(),
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		PolicyIdentifiers: []asn1.ObjectIdentifier{
			constants.Policy5,
			constants.Policy4,
		},
		BasicConstraintsValid: true,
	}

	if len(c.configuration.FQDNs) > 0 {
		serverCertificateTemplate.DNSNames = c.configuration.FQDNs
	}

	if len(c.configuration.IpAddresses) > 0 {
		for _, ip := range c.configuration.IpAddresses {
			serverCertificateTemplate.IPAddresses = append(serverCertificateTemplate.IPAddresses, net.ParseIP(ip))
		}
	}

	ctx.Log().Debug("Starting to generate private key")
	priv, err := rsa.GenerateKey(rand.Reader, int(c.configuration.KeySize))
	if err != nil {
		panic(err)
	}

	subjectKeyId, err := generateSubjectKeyId(priv)
	if err == nil {
		serverCertificateTemplate.SubjectKeyId = subjectKeyId
		serverCertificateTemplate.AuthorityKeyId = subjectKeyId
	}

	serverCertificate, serverPemCertificate := generateCertificate(&serverCertificateTemplate, c.intermediateCA.GetCertificate(), &priv.PublicKey, c.intermediateCA.GetPrivateKey())
	csr, _ := generateCertificateRequest(serverCertificate, priv)

	c.privateKey = priv
	c.certificate = serverCertificate
	c.pem = serverPemCertificate
	// Add the root CA and intermediate CA to the certificate
	c.pem = append(c.pem, c.rootCA.GetPemCertificate()...)
	c.pem = append(c.pem, c.intermediateCA.GetPemCertificate()...)
	c.csr = csr
	c.privateKeyPem = generatePemPrivateKey(priv)
	c.configuration.CertificateType = pkg_types.CertificateTypeCertificate

	return c, nil
}

func (c *X509ServerCertificate) LoadFromFile(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("load_certificate")
	cfg := config.GetInstance().Get()
	rootFolder := cfg.StoragePath()
	certificateFolder := helper.JoinPath(rootFolder, constants.CertificateStorageFolder)
	// create the folder if it doesn't exist
	if !helper.FileExists(certificateFolder) {
		helper.CreateDirectory(certificateFolder, fs.ModePerm)
	}

	certificateFileName := helper.JoinPath(certificateFolder, c.GetCertificateFileName())
	privateKeyFileName := helper.JoinPath(certificateFolder, c.GetPrivateKeyFileName())
	certificateRequestFileName := helper.JoinPath(certificateFolder, c.GetCertificateRequestFileName())

	if helper.FileExists(certificateFileName) {
		certificate, err := os.ReadFile(certificateFileName)
		if err != nil {
			diag.AddError("load_certificate", fmt.Sprintf("found error while reading certificate file, err %v", err.Error()), c.GetName())
			return diag
		}
		cert, err := x509.ParseCertificate(certificate)
		if err != nil {
			diag.AddError("load_certificate", fmt.Sprintf("found error while parsing certificate file, err %v", err.Error()), c.GetName())
			return diag
		}
		c.certificate = cert
	}

	if helper.FileExists(privateKeyFileName) {
		privateKey, err := os.ReadFile(privateKeyFileName)
		if err != nil {
			diag.AddError("load_certificate", fmt.Sprintf("found error while reading private key file, err %v", err.Error()), c.GetName())
			return diag
		}
		priv, err := x509.ParsePKCS1PrivateKey(privateKey)
		if err != nil {
			diag.AddError("load_certificate", fmt.Sprintf("found error while parsing private key file, err %v", err.Error()), c.GetName())
			return diag
		}
		c.privateKey = priv
	}

	if helper.FileExists(certificateRequestFileName) {
		certificateRequest, err := os.ReadFile(certificateRequestFileName)
		if err != nil {
			diag.AddError("load_certificate", fmt.Sprintf("found error while reading certificate request file, err %v", err.Error()), c.GetName())
			return diag
		}
		c.csr = certificateRequest
	}

	return diag
}

func (c *X509ServerCertificate) SaveToFile(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("save_certificate")
	cfg := config.GetInstance().Get()
	rootFolder := cfg.StoragePath()
	certificateFolder := helper.JoinPath(rootFolder, constants.CertificateStorageFolder)
	// create the folder if it doesn't exist
	if !helper.FileExists(certificateFolder) {
		helper.CreateDirectory(certificateFolder, fs.ModePerm)
	}

	certificateFileName := helper.JoinPath(certificateFolder, c.GetCertificateFileName())
	privateKeyFileName := helper.JoinPath(certificateFolder, c.GetPrivateKeyFileName())
	certificateRequestFileName := helper.JoinPath(certificateFolder, c.GetCertificateRequestFileName())

	ctx.Log().Debug("Exporting Certificate")
	if helper.FileExists(certificateFileName) {
		helper.DeleteFile(certificateFileName)
	}

	helper.WriteToFile(string(c.pem), certificateFileName)

	ctx.Log().Debug("Exporting Private Key")
	if helper.FileExists(privateKeyFileName) {
		helper.DeleteFile(privateKeyFileName)
	}

	helper.WriteToFile(string(c.privateKeyPem), privateKeyFileName)

	ctx.Log().Debug("Exporting CSR")
	if helper.FileExists(certificateRequestFileName) {
		helper.DeleteFile(certificateRequestFileName)
	}

	helper.WriteToFile(string(c.csr), certificateRequestFileName)
	return diag
}

func (c *X509ServerCertificate) Parse(ctx *appctx.AppContext, certificate string, privateKey string) *diagnostics.Diagnostics {
	diag := diagnostics.New("parse_certificate")
	if certificate != "" {
		certBlock, _ := pem.Decode([]byte(certificate))
		if certBlock == nil {
			err := errors.New("no valid certificate block found")
			ctx.Log().Errorf("found error while parsing  pem certificate block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("no valid certificate block found, err %v", err.Error()), c.GetName())
			return diag
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			ctx.Log().Errorf("found error while parsing certificate block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing certificate block, err %v", err.Error()), c.GetName())
			return diag
		}

		c.certificate = cert
		c.pem = []byte(certificate)
	}

	if privateKey != "" {
		privBlock, _ := pem.Decode([]byte(privateKey))
		if privBlock == nil {
			err := errors.New("no valid private key block found")
			ctx.Log().Errorf("found error while parsing  pem private key block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing  pem private key block, err %v", err.Error()), c.GetName())
			return diag
		}
		priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		if err != nil {
			ctx.Log().Errorf("found error while parsing private key block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing private key block, err %v", err.Error()), c.GetName())
			return diag
		}

		c.privateKey = priv
		c.privateKeyPem = []byte(privateKey)
	}
	return diag
}

func (c *X509ServerCertificate) Install(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate")
	cfg := config.GetInstance().Get()
	rootFolder := cfg.StoragePath()
	certificateFolder := helper.JoinPath(rootFolder, constants.CertificateStorageFolder)
	// create the folder if it doesn't exist
	if !helper.FileExists(certificateFolder) {
		helper.CreateDirectory(certificateFolder, fs.ModePerm)
	}

	certificateFileName := helper.JoinPath(certificateFolder, c.GetCertificateFileName())
	instalSvc := NewCertificateInstaller()
	instalSvc.InstallCertificate(ctx, certificateFileName, RootStore)
	return diag
}

func (c *X509ServerCertificate) GetCertificate() *x509.Certificate {
	return c.certificate
}

func (c *X509ServerCertificate) GetPrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func (c *X509ServerCertificate) GetConfiguration() *models.CertificateConfig {
	return &c.configuration
}

func (c *X509ServerCertificate) GetPemCertificate() []byte {
	return c.pem
}

func (c *X509ServerCertificate) GetPemCsr() []byte {
	return c.csr
}

func (c *X509ServerCertificate) GetPemPrivateKey() []byte {
	return c.privateKeyPem
}

func (c *X509ServerCertificate) FromDatabase(ctx *appctx.AppContext, certificate *entities.Certificate) *diagnostics.Diagnostics {
	diag := diagnostics.New("from_database")
	if parseDiag := c.Parse(ctx, certificate.PemCertificate, certificate.PemPrivateKey); parseDiag.HasErrors() {
		diag.Append(parseDiag)
		return diag
	}
	config := mappers.MapCertificateConfigToDto(certificate.Config)
	c.name = certificate.Name
	c.slug = certificate.Slug
	c.configuration = config
	return diag
}

func (c *X509ServerCertificate) baseFileName() string {
	baseFileName := strings.ReplaceAll(c.name, " ", "_")
	if baseFileName == "" {
		baseFileName = strings.ReplaceAll(c.configuration.CommonName, " ", "_")
	}
	return baseFileName
}
