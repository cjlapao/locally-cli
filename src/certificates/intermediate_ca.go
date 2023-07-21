package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"github.com/cjlapao/locally-cli/configuration"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjlapao/common-go/helper"
)

type X509IntermediateCertificate struct {
	Name                     string
	PrivateKey               *rsa.PrivateKey
	Certificate              *x509.Certificate
	Configuration            configuration.CertificateConfig
	IntermediateCertificates []configuration.RootCertificate
	Pem                      []byte
	Csr                      []byte
	PrivateKeyPem            []byte
}

func (intermediateCA *X509IntermediateCertificate) baseFileName() string {
	baseFileName := strings.ReplaceAll(intermediateCA.Name, " ", "_")
	if baseFileName == "" {
		baseFileName = strings.ReplaceAll(intermediateCA.Configuration.CommonName, " ", "_")
	}
	return baseFileName
}

func (intermediateCA *X509IntermediateCertificate) CertificateFileName() string {
	certificateFileName := "intermediateCA_" + intermediateCA.baseFileName() + ".crt"
	return certificateFileName
}

func (intermediateCA *X509IntermediateCertificate) PrivateKeyFileName() string {
	privateKeyFileName := "intermediateCA_" + intermediateCA.baseFileName() + ".key"
	return privateKeyFileName
}

func (intermediateCA *X509IntermediateCertificate) CertificateRequestFileName() string {
	certificateRequestFileName := "intermediateCA_" + intermediateCA.baseFileName() + ".csr"
	return certificateRequestFileName
}

func (intermediateCA *X509IntermediateCertificate) Generate(rootCA *X509RootCertificate, config *configuration.CertificateConfig) (*x509.Certificate, []byte, *rsa.PrivateKey) {
	logger.Debug("Starting to generate root certificate")

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

	intermediateCertificateTemplate := x509.Certificate{
		SerialNumber: generateSerialNumber(),
		Subject:      subject,
		NotBefore:    time.Now().Add(-10 * time.Second),
		NotAfter:     time.Now().AddDate(config.ExpiresInYears, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
		SignatureAlgorithm:    config.SignatureAlgorithm.ToX509SignatureAlgorithm(),
		DNSNames:              config.FQDNs,
		BasicConstraintsValid: true,
		PolicyIdentifiers: []asn1.ObjectIdentifier{
			policy4,
			policy5,
			policy6,
		},
	}

	if config.FQDNs != nil && len(config.FQDNs) > 0 {
		intermediateCertificateTemplate.DNSNames = config.FQDNs
	}

	if config.IpAddresses != nil && len(config.IpAddresses) > 0 {
		for _, ip := range config.IpAddresses {
			intermediateCertificateTemplate.IPAddresses = append(intermediateCertificateTemplate.IPAddresses, net.ParseIP(ip))
		}
	}

	logger.Debug("Starting to generate private key")
	priv, err := rsa.GenerateKey(rand.Reader, int(config.KeySize))
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

	intermediateCA.PrivateKey = priv
	intermediateCA.Certificate = intermediateCertificate
	intermediateCA.Configuration = *config
	intermediateCA.Pem = intermediatePemCertificate
	intermediateCA.Pem = append(intermediateCA.Pem, rootCA.Pem...)
	intermediateCA.Csr = csr
	intermediateCA.PrivateKeyPem = generatePemPrivateKey(priv)

	return intermediateCertificate, intermediatePemCertificate, priv
}

func (intermediateCA *X509IntermediateCertificate) LoadFromFile() error {
	return nil
}

func (intermediateCA *X509IntermediateCertificate) Parse(certificate string, privateKey string) error {
	if certificate != "" {
		certBlock, _ := pem.Decode([]byte(certificate))
		if certBlock == nil {
			err := errors.New("no valid certificate block found")
			logger.Error("found error while parsing  pem certificate block, err %v", err.Error())
			return err
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			logger.Error("found error while parsing certificate block, err %v", err.Error())
			return err
		}

		intermediateCA.Certificate = cert
		intermediateCA.Pem = []byte(certificate)
	}

	if privateKey != "" {
		privBlock, _ := pem.Decode([]byte(privateKey))
		if privBlock == nil {
			err := errors.New("no valid private key block found")
			logger.Error("found error while parsing  pem private key block, err %v", err.Error())
			return err
		}
		priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		if err != nil {
			logger.Error("found error while parsing private key block, err %v", err.Error())
			return err
		}

		intermediateCA.PrivateKey = priv
		intermediateCA.PrivateKeyPem = []byte(privateKey)
	}
	return nil
}

func (intermediateCA *X509IntermediateCertificate) SaveToFile() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)

	exPath = helper.JoinPath(exPath, "certificates")
	if !helper.FileExists(exPath) {
		helper.CreateDirectory(exPath, fs.ModePerm)
	}

	intermediateCertificateFileName := helper.JoinPath(exPath, intermediateCA.CertificateFileName())
	intermediatePrivateKeyFileName := helper.JoinPath(exPath, intermediateCA.PrivateKeyFileName())
	intermediateCsrFileName := helper.JoinPath(exPath, intermediateCA.CertificateRequestFileName())

	logger.Debug("Exporting Intermediate CA Certificate")
	if helper.FileExists(intermediateCertificateFileName) {
		helper.DeleteFile(intermediateCertificateFileName)
	}

	helper.WriteToFile(string(intermediateCA.Pem), intermediateCertificateFileName)

	logger.Debug("Exporting Intermediate CA Private Key")
	if helper.FileExists(intermediatePrivateKeyFileName) {
		helper.DeleteFile(intermediatePrivateKeyFileName)
	}

	helper.WriteToFile(string(intermediateCA.PrivateKeyPem), intermediatePrivateKeyFileName)

	logger.Debug("Exporting Intermediate CA CSR")
	if helper.FileExists(intermediateCsrFileName) {
		helper.DeleteFile(intermediateCsrFileName)
	}

	helper.WriteToFile(string(intermediateCA.Csr), intermediateCsrFileName)
	return nil
}

func (intermediateCA *X509IntermediateCertificate) Install() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)

	exPath = helper.JoinPath(exPath, "certificates")
	if !helper.FileExists(exPath) {
		helper.CreateDirectory(exPath, fs.ModePerm)
	}

	intermediateCertificateFileName := helper.JoinPath(exPath, intermediateCA.CertificateFileName())
	instalSvc := Installer{}
	instalSvc.InstallCertificate(intermediateCertificateFileName, CAStore)
	return nil
}
