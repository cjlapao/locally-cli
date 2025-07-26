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

type X509ServerCertificate struct {
	ctx           *appctx.AppContext
	Name          string
	PrivateKey    *rsa.PrivateKey
	Certificate   *x509.Certificate
	Configuration models.CertificateConfig
	Pem           []byte
	Csr           []byte
	PrivateKeyPem []byte
}

func NewX509ServerCertificate(ctx *appctx.AppContext, name string, config models.CertificateConfig) *X509ServerCertificate {
	return &X509ServerCertificate{
		ctx:           ctx,
		Name:          name,
		Configuration: config,
	}
}

func (serverCert *X509ServerCertificate) baseFileName() string {
	baseFileName := strings.ReplaceAll(serverCert.Name, " ", "_")
	if baseFileName == "" {
		baseFileName = strings.ReplaceAll(serverCert.Configuration.CommonName, " ", "_")
	}
	return baseFileName
}

func (serverCert *X509ServerCertificate) CertificateFileName() string {
	certificateFileName := serverCert.baseFileName() + ".crt"
	return certificateFileName
}

func (serverCert *X509ServerCertificate) PrivateKeyFileName() string {
	privateKeyFileName := serverCert.baseFileName() + ".key"
	return privateKeyFileName
}

func (serverCert *X509ServerCertificate) CertificateRequestFileName() string {
	certificateRequestFileName := serverCert.baseFileName() + ".csr"
	return certificateRequestFileName
}

func (serverCert *X509ServerCertificate) CertificatePfxFileName() string {
	certificatePfxFileName := serverCert.baseFileName() + ".pfx"
	return certificatePfxFileName
}

func (serverCert *X509ServerCertificate) Generate(ctx *appctx.AppContext, intermediateCA *X509IntermediateCertificate, config *models.CertificateConfig) (*x509.Certificate, []byte, *rsa.PrivateKey) {
	ctx.Log().Debug("Starting to generate server certificate certificate")

	subject := pkix.Name{
		Country:            []string{config.Country},
		Organization:       []string{config.Organization},
		OrganizationalUnit: []string{config.OrganizationalUnit},
		Province:           []string{config.State},
		Locality:           []string{config.City},
		CommonName:         config.CommonName,
	}

	if config.AdminEmailAddress != "" {
		subject.ExtraNames = []pkix.AttributeTypeAndValue{
			{
				Type: oidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(config.AdminEmailAddress),
				},
			},
		}
	}

	serverCertificateTemplate := x509.Certificate{
		SerialNumber:       generateSerialNumber(),
		Subject:            subject,
		NotBefore:          time.Now().Add(-10 * time.Second),
		NotAfter:           time.Now().AddDate(config.ExpiresInYears, 0, 0),
		IsCA:               false,
		SignatureAlgorithm: config.SignatureAlgorithm.ToX509SignatureAlgorithm(),
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		PolicyIdentifiers: []asn1.ObjectIdentifier{
			policy5,
			policy4,
		},
		BasicConstraintsValid: true,
	}

	if len(config.FQDNs) > 0 {
		serverCertificateTemplate.DNSNames = config.FQDNs
	}

	if len(config.IpAddresses) > 0 {
		for _, ip := range config.IpAddresses {
			serverCertificateTemplate.IPAddresses = append(serverCertificateTemplate.IPAddresses, net.ParseIP(ip))
		}
	}

	ctx.Log().Debug("Starting to generate private key")
	priv, err := rsa.GenerateKey(rand.Reader, int(config.KeySize))
	if err != nil {
		panic(err)
	}

	subjectKeyId, err := generateSubjectKeyId(priv)
	if err == nil {
		serverCertificateTemplate.SubjectKeyId = subjectKeyId
		serverCertificateTemplate.AuthorityKeyId = subjectKeyId
	}

	serverCertificate, serverPemCertificate := generateCertificate(&serverCertificateTemplate, intermediateCA.Certificate, &priv.PublicKey, intermediateCA.PrivateKey)
	csr, _ := generateCertificateRequest(serverCertificate, priv)

	serverCert.PrivateKey = priv
	serverCert.Certificate = serverCertificate
	serverCert.Configuration = *config
	serverCert.Pem = serverPemCertificate
	serverCert.Pem = append(serverCert.Pem, intermediateCA.Pem...)
	serverCert.Csr = csr
	serverCert.PrivateKeyPem = generatePemPrivateKey(priv)

	return serverCertificate, serverPemCertificate, priv
}

func (serverCert *X509ServerCertificate) LoadFromFile() error {
	return nil
}

func (serverCert *X509ServerCertificate) Parse(ctx *appctx.AppContext, certificate string, privateKey string) *diagnostics.Diagnostics {
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

		serverCert.Certificate = cert
		serverCert.Pem = []byte(certificate)
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

		serverCert.PrivateKey = priv
		serverCert.PrivateKeyPem = []byte(privateKey)
	}
	return diag
}

func (serverCert *X509ServerCertificate) SaveToFile(ctx *appctx.AppContext) *diagnostics.Diagnostics {
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

	serverCertificateFileName := helper.JoinPath(exPath, serverCert.CertificateFileName())
	serverPrivateKeyFileName := helper.JoinPath(exPath, serverCert.PrivateKeyFileName())
	serverCsrFileName := helper.JoinPath(exPath, serverCert.CertificateRequestFileName())
	serverPfxFileName := helper.JoinPath(exPath, serverCert.CertificatePfxFileName())
	ctx.Log().Debug("Exporting Server Certificate Certificate")
	if helper.FileExists(serverCertificateFileName) {
		helper.DeleteFile(serverCertificateFileName)
	}

	helper.WriteToFile(string(serverCert.Pem), serverCertificateFileName)

	ctx.Log().Debug("Exporting Server Certificate Private Key")
	if helper.FileExists(serverPrivateKeyFileName) {
		helper.DeleteFile(serverPrivateKeyFileName)
	}

	helper.WriteToFile(string(serverCert.PrivateKeyPem), serverPrivateKeyFileName)

	ctx.Log().Debug("Exporting Server Certificate CSR")
	if helper.FileExists(serverCsrFileName) {
		helper.DeleteFile(serverCsrFileName)
	}

	helper.WriteToFile(string(serverCert.Csr), serverCsrFileName)
	if serverCert.Configuration.Password != "" {
		ctx.Log().Debug("Exporting Server Certificate PFX")
		if helper.FileExists(serverPfxFileName) {
			helper.DeleteFile(serverPfxFileName)
		}

		pfx, err := generatePfx(serverCert.Certificate, serverCert.PrivateKey, serverCert.Configuration.Password)
		if err != nil {
			diag.AddError("save_certificate", fmt.Sprintf("found error while generating pfx, err %v", err.Error()), CertificateComponent)
			return diag
		}

		helper.WriteToFile(string(pfx), serverPfxFileName)
	}
	return diag
}

func (serverCert *X509ServerCertificate) Install(ctx *appctx.AppContext) *diagnostics.Diagnostics {
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

	serverCertificateFileName := helper.JoinPath(exPath, serverCert.CertificateFileName())

	instalSvc := Installer{}
	instalSvc.InstallCertificate(serverCert.ctx, serverCertificateFileName, WebHosting)
	return diag
}
