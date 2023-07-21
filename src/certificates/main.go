package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"github.com/cjlapao/locally-cli/configuration"
	"math/big"

	"github.com/cjlapao/common-go/log"
	"software.sslmate.com/src/go-pkcs12"
)

var logger = log.Get()
var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}

// var policy1 = asn1.ObjectIdentifier{1, 2, 4, 5}
// var policy2 = asn1.ObjectIdentifier{1, 1, 3, 4}

// Keys
var ServerAuthentication = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 1}
var ClientAuthentication = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 2}
var CodeSigning = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 3}
var SecureEmail = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 4}
var TimeStamping = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 8}
var OCSPSigning = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 9}
var MicrosoftTrustListSigning = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 1}
var EncryptingFileSystem = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 4}

// var policy3 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 2}
var policy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3}
var policy5 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 1}
var policy6 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 2}

var rootPolicy3 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 2}
var rootPolicy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 3}

type GeneratedCertificate interface {
}

func generateCertificate(template, parent *x509.Certificate, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (*x509.Certificate, []byte) {
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		panic("Failed to create certificate:" + err.Error())
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		panic("Failed to parse certificate:" + err.Error())
	}

	b := pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
	certPEM := pem.EncodeToMemory(&b)

	return cert, certPEM
}

func generateCertificateRequest(certificate *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
	template := x509.CertificateRequest{
		Subject:            certificate.Subject,
		SignatureAlgorithm: certificate.SignatureAlgorithm,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)

	if err != nil {
		return nil, err
	}

	b := pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}
	csrPem := pem.EncodeToMemory(&b)

	return csrPem, nil
}

func generatePemPrivateKey(privateKey *rsa.PrivateKey) []byte {

	b := pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privateKeyPem := pem.EncodeToMemory(&b)

	return privateKeyPem
}

func generateSerialNumber() *big.Int {
	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return big.NewInt(1)
	}

	return serialNumber
}

func generateSubjectKeyId(privateKey *rsa.PrivateKey) ([]byte, error) {
	publicKeyBytes, err := asn1.Marshal(*privateKey.Public().(*rsa.PublicKey))
	if err != nil {
		return nil, err
	}

	subjectKeyId := sha1.Sum(publicKeyBytes)

	return subjectKeyId[:], nil
}

func generatePfx(certificate *x509.Certificate, privateKey *rsa.PrivateKey, password string) ([]byte, error) {
	pfxBytes, err := pkcs12.Encode(rand.Reader, privateKey, certificate, []*x509.Certificate{}, password)

	if err != nil {
		return nil, err
	}

	// see if pfxBytes valid
	_, _, _, err = pkcs12.DecodeChain(pfxBytes, password)
	if err != nil {
		return nil, err
	}

	return pfxBytes, nil
}

func GenerateCertificates() {
	var config = configuration.Get()

	if config.GlobalConfiguration.CertificateGenerator != nil {
		logger.Info("|- Root")
		needsSaving := false
		if config.GlobalConfiguration.CertificateGenerator.Root != nil && len(config.GlobalConfiguration.CertificateGenerator.Root) > 0 {
			for _, rootCert := range config.GlobalConfiguration.CertificateGenerator.Root {
				x509RootCert := X509RootCertificate{}
				x509RootCert.Name = rootCert.Name
				if rootCert.PemCertificate == "" {
					x509RootCert.Generate(rootCert.Config)
					rootCert.PemCertificate = string(x509RootCert.Pem)
					rootCert.PemPrivateKey = string(x509RootCert.PrivateKeyPem)
					needsSaving = true
					logger.Info("|  |- %s", rootCert.Name)
				} else {
					x509RootCert.Configuration = *rootCert.Config
					x509RootCert.Parse(rootCert.PemCertificate, rootCert.PemPrivateKey)
					logger.Info("|  |- %s [Cached]", rootCert.Name)
				}

				if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
					x509RootCert.SaveToFile()
				}
				x509RootCert.Install()
				for _, intermediateCA := range rootCert.IntermediateCertificates {
					x509IntermediateCert := X509IntermediateCertificate{}
					x509IntermediateCert.Name = intermediateCA.Name
					if intermediateCA.PemCertificate == "" {
						x509IntermediateCert.Generate(&x509RootCert, intermediateCA.Config)
						intermediateCA.PemCertificate = string(x509IntermediateCert.Pem)
						intermediateCA.PemPrivateKey = string(x509IntermediateCert.PrivateKeyPem)
						needsSaving = true
						logger.Info("|  |  |- %s", intermediateCA.Name)
					} else {
						x509IntermediateCert.Configuration = *intermediateCA.Config
						x509IntermediateCert.Parse(intermediateCA.PemCertificate, intermediateCA.PemPrivateKey)
						logger.Info("|  |  |- %s [Cached]", intermediateCA.Name)
					}
					if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
						x509IntermediateCert.SaveToFile()
					}
					x509IntermediateCert.Install()

					for _, serverCert := range intermediateCA.Certificates {
						x509ServerCert := X509ServerCertificate{}
						x509ServerCert.Name = serverCert.Name
						if serverCert.PemCertificate == "" {
							x509ServerCert.Generate(&x509IntermediateCert, serverCert.Config)
							serverCert.PemCertificate = string(x509ServerCert.Pem)
							serverCert.PemPrivateKey = string(x509ServerCert.PrivateKeyPem)
							needsSaving = true
							logger.Info("|  |  |  |- %s", serverCert.Name)
						} else {
							x509ServerCert.Configuration = *serverCert.Config
							x509ServerCert.Parse(serverCert.PemCertificate, serverCert.PemPrivateKey)
							logger.Info("|  |  |  |- %s [Cached]", serverCert.Name)
						}
						if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
							x509ServerCert.SaveToFile()
						}
						x509ServerCert.Install()
					}
				}
			}
		}

		if needsSaving {
			config.SaveConfigFile()
		}
	} else {
		logger.Warn("No Certificate Generator configuration found, exiting")
	}
}

func CleanConfig() {
	var config = configuration.Get()

	if config.GlobalConfiguration.CertificateGenerator != nil {
		if config.GlobalConfiguration.CertificateGenerator.Root != nil && len(config.GlobalConfiguration.CertificateGenerator.Root) > 0 {
			for _, rootCert := range config.GlobalConfiguration.CertificateGenerator.Root {
				rootCert.PemCertificate = ""
				rootCert.PemPrivateKey = ""
				for _, intermediateCert := range rootCert.IntermediateCertificates {
					intermediateCert.PemCertificate = ""
					intermediateCert.PemPrivateKey = ""
					for _, serverCert := range intermediateCert.Certificates {
						serverCert.PemCertificate = ""
						serverCert.PemPrivateKey = ""
					}
				}
			}
		}
	}

	config.SaveConfigFile()
}
