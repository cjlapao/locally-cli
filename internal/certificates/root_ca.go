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

type X509RootCertificate struct {
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

func NewX509RootCertificate(ctx *appctx.AppContext, name string, config models.CertificateConfig) *X509RootCertificate {
	return &X509RootCertificate{
		ctx:           ctx,
		Name:          name,
		Configuration: config,
	}
}

func (rootCA *X509RootCertificate) baseFileName() string {
	baseFileName := strings.ReplaceAll(rootCA.Name, " ", "_")
	if baseFileName == "" {
		baseFileName = strings.ReplaceAll(rootCA.Configuration.CommonName, " ", "_")
	}
	return baseFileName
}

func (rootCA *X509RootCertificate) CertificateFileName() string {
	certificateFileName := "rootca_" + rootCA.baseFileName() + ".crt"
	return certificateFileName
}

func (rootCA *X509RootCertificate) PrivateKeyFileName() string {
	privateKeyFileName := "rootca_" + rootCA.baseFileName() + ".key"
	return privateKeyFileName
}

func (rootCA *X509RootCertificate) CertificateRequestFileName() string {
	certificateRequestFileName := "rootca_" + rootCA.baseFileName() + ".csr"
	return certificateRequestFileName
}

func (rootCA *X509RootCertificate) Generate(ctx *appctx.AppContext) (*models.RootCertificate, *diagnostics.Diagnostics) {
	ctx.Log().Debug("Starting to generate root certificate")
	diag := diagnostics.New("generate_root_certificate")

	subject := pkix.Name{
		Country:            []string{rootCA.Configuration.Country},
		Organization:       []string{rootCA.Configuration.Organization},
		OrganizationalUnit: []string{rootCA.Configuration.OrganizationalUnit},
		Province:           []string{rootCA.Configuration.State},
		Locality:           []string{rootCA.Configuration.City},
		CommonName:         rootCA.Configuration.CommonName,
	}

	if rootCA.Configuration.AdminEmailAddress != "" {
		subject.ExtraNames = []pkix.AttributeTypeAndValue{
			{
				Type: oidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(rootCA.Configuration.AdminEmailAddress),
				},
			},
		}
	}

	rootCertificateTemplate := x509.Certificate{
		SerialNumber: generateSerialNumber(),
		Subject:      subject,
		NotBefore:    time.Now().Add(-10 * time.Second),
		NotAfter:     time.Now().AddDate(rootCA.Configuration.ExpiresInYears, 0, 0),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageEmailProtection,
			x509.ExtKeyUsageIPSECEndSystem,
			x509.ExtKeyUsageIPSECTunnel,
			x509.ExtKeyUsageIPSECUser,
			x509.ExtKeyUsageOCSPSigning,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageTimeStamping,
		},
		IsCA:                  true,
		MaxPathLen:            2,
		SignatureAlgorithm:    rootCA.Configuration.SignatureAlgorithm.ToX509SignatureAlgorithm(),
		DNSNames:              rootCA.Configuration.FQDNs,
		BasicConstraintsValid: true,
		PolicyIdentifiers: []asn1.ObjectIdentifier{
			policy5,
			policy6,
			rootPolicy3,
			rootPolicy4,
		},
	}

	if len(rootCA.Configuration.FQDNs) > 0 {
		rootCertificateTemplate.DNSNames = rootCA.Configuration.FQDNs
	}

	if len(rootCA.Configuration.IpAddresses) > 0 {
		for _, ip := range rootCA.Configuration.IpAddresses {
			rootCertificateTemplate.IPAddresses = append(rootCertificateTemplate.IPAddresses, net.ParseIP(ip))
		}
	}

	ctx.Log().Debug("Starting to generate root certificate private key")
	priv, err := rsa.GenerateKey(rand.Reader, int(rootCA.Configuration.KeySize))
	if err != nil {
		diag.AddError("generate_root_certificate", fmt.Sprintf("found error while generating private key, err %v", err.Error()), CertificateComponent)
		return nil, diag
	}

	subjectKeyId, err := generateSubjectKeyId(priv)
	if err == nil {
		rootCertificateTemplate.SubjectKeyId = subjectKeyId
		rootCertificateTemplate.AuthorityKeyId = rootCertificateTemplate.SubjectKeyId
	}

	rootCertificate, rootPemCertificate := generateCertificate(&rootCertificateTemplate, &rootCertificateTemplate, &priv.PublicKey, priv)
	csr, _ := generateCertificateRequest(rootCertificate, priv)

	rootCA.PrivateKey = priv
	rootCA.Certificate = rootCertificate
	rootCA.Pem = rootPemCertificate
	rootCA.Csr = csr
	rootCA.PrivateKeyPem = generatePemPrivateKey(priv)

	response := &models.RootCertificate{
		Name:           rootCA.Name,
		Config:         &rootCA.Configuration,
		PemCertificate: string(rootCA.Pem),
		PemPrivateKey:  string(rootCA.PrivateKeyPem),
		Csr:            string(rootCA.Csr),
	}

	return response, diag
}

func (rootCA *X509RootCertificate) LoadFromFile() error {
	return nil
}

func (rootCA *X509RootCertificate) Parse(ctx *appctx.AppContext, certificate string, privateKey string) *diagnostics.Diagnostics {
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

		rootCA.Certificate = cert
		rootCA.Pem = []byte(certificate)
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

		rootCA.PrivateKey = priv
		rootCA.PrivateKeyPem = []byte(privateKey)
	}
	return diag
}

func (rootCA *X509RootCertificate) SaveToFile(ctx *appctx.AppContext) *diagnostics.Diagnostics {
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

	rootCertificateFileName := helper.JoinPath(exPath, rootCA.CertificateFileName())
	rootPrivateKeyFileName := helper.JoinPath(exPath, rootCA.PrivateKeyFileName())
	rootCsrFileName := helper.JoinPath(exPath, rootCA.CertificateRequestFileName())

	ctx.Log().Debug("Exporting Certificate")
	if helper.FileExists(rootCertificateFileName) {
		helper.DeleteFile(rootCertificateFileName)
	}

	helper.WriteToFile(string(rootCA.Pem), rootCertificateFileName)

	ctx.Log().Debug("Exporting Private Key")
	if helper.FileExists(rootPrivateKeyFileName) {
		helper.DeleteFile(rootPrivateKeyFileName)
	}

	helper.WriteToFile(string(rootCA.PrivateKeyPem), rootPrivateKeyFileName)

	ctx.Log().Debug("Exporting CSR")
	if helper.FileExists(rootCsrFileName) {
		helper.DeleteFile(rootCsrFileName)
	}

	helper.WriteToFile(string(rootCA.Csr), rootCsrFileName)
	return diag
}

func (rootCA *X509RootCertificate) Install(ctx *appctx.AppContext) *diagnostics.Diagnostics {
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

	rootCertificateFileName := helper.JoinPath(exPath, rootCA.CertificateFileName())

	instalSvc := Installer{}
	instalSvc.InstallCertificate(rootCA.ctx, rootCertificateFileName, RootStore)
	return diag
}
