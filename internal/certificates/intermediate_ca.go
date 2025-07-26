package certificates

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
	"path/filepath"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"

	"github.com/cjlapao/common-go/helper"
)

type X509IntermediateCertificate struct {
	ctx                      *appctx.AppContext
	Name                     string
	PrivateKey               *rsa.PrivateKey
	Certificate              *x509.Certificate
	Configuration            models.CertificateConfig
	IntermediateCertificates []models.RootCertificate
	Pem                      []byte
	Csr                      []byte
	PrivateKeyPem            []byte
}

func NewX509IntermediateCertificate(ctx *appctx.AppContext, name string, config models.CertificateConfig) *X509IntermediateCertificate {
	return &X509IntermediateCertificate{
		ctx:           ctx,
		Name:          name,
		Configuration: config,
	}
}

func (c *X509IntermediateCertificate) baseFileName() string {
	baseFileName := strings.ReplaceAll(c.Name, " ", "_")
	if baseFileName == "" {
		baseFileName = strings.ReplaceAll(c.Configuration.CommonName, " ", "_")
	}
	return baseFileName
}

func (c *X509IntermediateCertificate) CertificateFileName() string {
	certificateFileName := "intermediateCA_" + c.baseFileName() + ".crt"
	return certificateFileName
}

func (c *X509IntermediateCertificate) PrivateKeyFileName() string {
	privateKeyFileName := "intermediateCA_" + c.baseFileName() + ".key"
	return privateKeyFileName
}

func (c *X509IntermediateCertificate) CertificateRequestFileName() string {
	certificateRequestFileName := "intermediateCA_" + c.baseFileName() + ".csr"
	return certificateRequestFileName
}

func (c *X509IntermediateCertificate) Generate(ctx *appctx.AppContext, rootCA *X509RootCertificate) (*models.IntermediateCertificate, *diagnostics.Diagnostics) {
	ctx.Log().Debug("Starting to generate intermediate certificate")
	diag := diagnostics.New("generate_intermediate_certificate")

	subject := pkix.Name{
		Country:            []string{c.Configuration.Country},
		Organization:       []string{c.Configuration.Organization},
		OrganizationalUnit: []string{c.Configuration.OrganizationalUnit},
		Province:           []string{c.Configuration.State},
		Locality:           []string{c.Configuration.City},
		CommonName:         c.Configuration.CommonName,
	}

	if c.Configuration.AdminEmailAddress != "" {
		subject.ExtraNames = []pkix.AttributeTypeAndValue{
			{
				Type: oidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(c.Configuration.AdminEmailAddress),
				},
			},
		}
	}

	intermediateCertificateTemplate := x509.Certificate{
		SerialNumber: generateSerialNumber(),
		Subject:      subject,
		NotBefore:    time.Now().Add(-10 * time.Second),
		NotAfter:     time.Now().AddDate(c.Configuration.ExpiresInYears, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
		SignatureAlgorithm:    c.Configuration.SignatureAlgorithm.ToX509SignatureAlgorithm(),
		DNSNames:              c.Configuration.FQDNs,
		BasicConstraintsValid: true,
		PolicyIdentifiers: []asn1.ObjectIdentifier{
			policy4,
			policy5,
			policy6,
		},
	}

	if len(c.Configuration.FQDNs) > 0 {
		intermediateCertificateTemplate.DNSNames = c.Configuration.FQDNs
	}

	if len(c.Configuration.IpAddresses) > 0 {
		for _, ip := range c.Configuration.IpAddresses {
			intermediateCertificateTemplate.IPAddresses = append(intermediateCertificateTemplate.IPAddresses, net.ParseIP(ip))
		}
	}

	ctx.Log().Debug("Starting to generate intermediate certificate private key")
	priv, err := rsa.GenerateKey(rand.Reader, int(c.Configuration.KeySize))
	if err != nil {
		panic(err)
	}

	subjectKeyId, err := generateSubjectKeyId(priv)
	if err == nil {
		intermediateCertificateTemplate.SubjectKeyId = subjectKeyId
		intermediateCertificateTemplate.AuthorityKeyId = intermediateCertificateTemplate.SubjectKeyId
	}

	intermediateCertificate, intermediatePemCertificate := generateCertificate(&intermediateCertificateTemplate, rootCA.Certificate, &priv.PublicKey, rootCA.PrivateKey)
	csr, _ := generateCertificateRequest(intermediateCertificate, priv)

	c.PrivateKey = priv
	c.Certificate = intermediateCertificate
	c.Pem = intermediatePemCertificate
	c.Pem = append(c.Pem, rootCA.Pem...)
	c.Csr = csr
	c.PrivateKeyPem = generatePemPrivateKey(priv)

	response := &models.IntermediateCertificate{
		Name:           c.Name,
		Config:         &c.Configuration,
		PemCertificate: string(c.Pem),
		PemPrivateKey:  string(c.PrivateKeyPem),
		Csr:            string(c.Csr),
	}

	return response, diag
}

func (c *X509IntermediateCertificate) LoadFromFile() error {
	return nil
}

func (c *X509IntermediateCertificate) Parse(ctx *appctx.AppContext, certificate string, privateKey string) *diagnostics.Diagnostics {
	diag := diagnostics.New("parse_certificate")
	if certificate != "" {
		certBlock, _ := pem.Decode([]byte(certificate))
		if certBlock == nil {
			err := errors.New("no valid certificate block found")
			ctx.Log().Errorf("found error while parsing  pem certificate block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("no valid certificate block found, err %v", err.Error()), CertificateComponent)
			return diag
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			ctx.Log().Errorf("found error while parsing certificate block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing certificate block, err %v", err.Error()), CertificateComponent)
			return diag
		}

		c.Certificate = cert
		c.Pem = []byte(certificate)
	}

	if privateKey != "" {
		privBlock, _ := pem.Decode([]byte(privateKey))
		if privBlock == nil {
			err := errors.New("no valid private key block found")
			ctx.Log().Errorf("found error while parsing  pem private key block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing  pem private key block, err %v", err.Error()), CertificateComponent)
			return diag
		}
		priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		if err != nil {
			ctx.Log().Errorf("found error while parsing private key block, err %v", err.Error())
			diag.AddError("parse_certificate", fmt.Sprintf("found error while parsing private key block, err %v", err.Error()), CertificateComponent)
			return diag
		}

		c.PrivateKey = priv
		c.PrivateKeyPem = []byte(privateKey)
	}
	return nil
}

func (c *X509IntermediateCertificate) SaveToFile(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("save_certificate")
	ex, err := os.Executable()
	if err != nil {
		diag.AddError("save_certificate", fmt.Sprintf("found error while getting executable path, err %v", err.Error()), CertificateComponent)
		return diag
	}

	exPath := filepath.Dir(ex)

	exPath = helper.JoinPath(exPath, "certificates")
	if !helper.FileExists(exPath) {
		helper.CreateDirectory(exPath, fs.ModePerm)
	}

	intermediateCertificateFileName := helper.JoinPath(exPath, c.CertificateFileName())
	intermediatePrivateKeyFileName := helper.JoinPath(exPath, c.PrivateKeyFileName())
	intermediateCsrFileName := helper.JoinPath(exPath, c.CertificateRequestFileName())

	ctx.Log().Debug("Exporting Intermediate CA Certificate")
	if helper.FileExists(intermediateCertificateFileName) {
		helper.DeleteFile(intermediateCertificateFileName)
	}

	helper.WriteToFile(string(c.Pem), intermediateCertificateFileName)

	ctx.Log().Debug("Exporting Intermediate CA Private Key")
	if helper.FileExists(intermediatePrivateKeyFileName) {
		helper.DeleteFile(intermediatePrivateKeyFileName)
	}

	helper.WriteToFile(string(c.PrivateKeyPem), intermediatePrivateKeyFileName)

	ctx.Log().Debug("Exporting Intermediate CA CSR")
	if helper.FileExists(intermediateCsrFileName) {
		helper.DeleteFile(intermediateCsrFileName)
	}

	helper.WriteToFile(string(c.Csr), intermediateCsrFileName)
	return diag
}

func (c *X509IntermediateCertificate) Install(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("install_certificate")
	ex, err := os.Executable()
	if err != nil {
		diag.AddError("install_certificate", fmt.Sprintf("found error while getting executable path, err %v", err.Error()), CertificateComponent)
		return diag
	}

	exPath := filepath.Dir(ex)

	exPath = helper.JoinPath(exPath, "certificates")
	if !helper.FileExists(exPath) {
		helper.CreateDirectory(exPath, fs.ModePerm)
	}

	intermediateCertificateFileName := helper.JoinPath(exPath, c.CertificateFileName())
	instalSvc := Installer{}
	instalSvc.InstallCertificate(c.ctx, intermediateCertificateFileName, CAStore)
	return diag
}
