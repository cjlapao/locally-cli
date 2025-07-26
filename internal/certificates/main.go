// Package certificates provides a service for generating and managing certificates
package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"sync"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"

	"software.sslmate.com/src/go-pkcs12"
)

var (
	certificateService     *CertificateService
	certificateServiceOnce sync.Once
	CertificateComponent   = "Certificates Service"
)

type CertificateType string

const (
	RootCertificate         CertificateType = "Root"
	IntermediateCertificate CertificateType = "Intermediate"
	ServerCertificate       CertificateType = "Server"
	ClientCertificate       CertificateType = "Client"
)

// var policy1 = asn1.ObjectIdentifier{1, 2, 4, 5}
// var policy2 = asn1.ObjectIdentifier{1, 1, 3, 4}
// var policy3 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 2}

// Certificate Keys
var (
	oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}

	ServerAuthentication      = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 1}
	ClientAuthentication      = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 2}
	CodeSigning               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 3}
	SecureEmail               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 4}
	TimeStamping              = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 8}
	OCSPSigning               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 9}
	MicrosoftTrustListSigning = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 1}
	EncryptingFileSystem      = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 4}

	policy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3}
	policy5 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 1}
	policy6 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 2}

	rootPolicy3 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 2}
	rootPolicy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 3}
)

type GeneratedCertificate interface{}

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

// func GenerateCertificates(ctx *appctx.AppContext) {
// 	config := configuration.Get()

// 	if config.GlobalConfiguration.CertificateGenerator != nil {
// 		ctx.Log().Info("|- Root")
// 		needsSaving := false
// 		if config.GlobalConfiguration.CertificateGenerator.Root != nil && len(config.GlobalConfiguration.CertificateGenerator.Root) > 0 {
// 			for _, rootCert := range config.GlobalConfiguration.CertificateGenerator.Root {
// 				x509RootCert := X509RootCertificate{}
// 				x509RootCert.Name = rootCert.Name
// 				if rootCert.PemCertificate == "" {
// 					x509RootCert.Generate(ctx, rootCert.Config)
// 					rootCert.PemCertificate = string(x509RootCert.Pem)
// 					rootCert.PemPrivateKey = string(x509RootCert.PrivateKeyPem)
// 					needsSaving = true
// 					ctx.Log().Info("|  |- %s", rootCert.Name)
// 				} else {
// 					x509RootCert.Configuration = *rootCert.Config
// 					x509RootCert.Parse(ctx, rootCert.PemCertificate, rootCert.PemPrivateKey)
// 					ctx.Log().Info("|  |- %s [Cached]", rootCert.Name)
// 				}

// 				if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
// 					x509RootCert.SaveToFile(ctx)
// 				}
// 				x509RootCert.Install(ctx)
// 				for _, intermediateCA := range rootCert.IntermediateCertificates {
// 					x509IntermediateCert := X509IntermediateCertificate{}
// 					x509IntermediateCert.Name = intermediateCA.Name
// 					if intermediateCA.PemCertificate == "" {
// 						x509IntermediateCert.Generate(ctx, &x509RootCert, intermediateCA.Config)
// 						intermediateCA.PemCertificate = string(x509IntermediateCert.Pem)
// 						intermediateCA.PemPrivateKey = string(x509IntermediateCert.PrivateKeyPem)
// 						needsSaving = true
// 						ctx.Log().Info("|  |  |- %s", intermediateCA.Name)
// 					} else {
// 						x509IntermediateCert.Configuration = *intermediateCA.Config
// 						x509IntermediateCert.Parse(ctx, intermediateCA.PemCertificate, intermediateCA.PemPrivateKey)
// 						ctx.Log().Info("|  |  |- %s [Cached]", intermediateCA.Name)
// 					}
// 					if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
// 						x509IntermediateCert.SaveToFile(ctx)
// 					}
// 					x509IntermediateCert.Install(ctx)

// 					for _, serverCert := range intermediateCA.Certificates {
// 						x509ServerCert := X509ServerCertificate{}
// 						x509ServerCert.Name = serverCert.Name
// 						if serverCert.PemCertificate == "" {
// 							x509ServerCert.Generate(ctx, &x509IntermediateCert, serverCert.Config)
// 							serverCert.PemCertificate = string(x509ServerCert.Pem)
// 							serverCert.PemPrivateKey = string(x509ServerCert.PrivateKeyPem)
// 							needsSaving = true
// 							ctx.Log().Info("|  |  |  |- %s", serverCert.Name)
// 						} else {
// 							x509ServerCert.Configuration = *serverCert.Config
// 							x509ServerCert.Parse(ctx, serverCert.PemCertificate, serverCert.PemPrivateKey)
// 							ctx.Log().Info("|  |  |  |- %s [Cached]", serverCert.Name)
// 						}
// 						if config.GlobalConfiguration.CertificateGenerator.OutputToFile {
// 							x509ServerCert.SaveToFile(ctx)
// 						}
// 						x509ServerCert.Install(ctx)
// 					}
// 				}
// 			}
// 		}

// 		if needsSaving {
// 			config.SaveConfigFile()
// 		}
// 	} else {
// 		ctx.Log().Warn("No Certificate Generator configuration found, exiting")
// 	}
// }

type CertificateService struct {
	store *stores.CertificatesDataStore
}

func (s *CertificateService) GetName() string {
	return "CertificateService"
}

func Initialize(store *stores.CertificatesDataStore) *CertificateService {
	logging.Info("Initializing Certificate Service")
	certificateServiceOnce.Do(func() {
		if store == nil {
			logging.Error("Certificate Service not initialized, store is nil")
			return
		}
		certificateService = New(store)
		logging.Info("Certificate Service initialized")
	})

	return certificateService
}

func GetInstance() *CertificateService {
	if certificateService == nil {
		logging.Error("Certificate Service not initialized, returning nil")
		return nil
	}

	return certificateService
}

func New(store *stores.CertificatesDataStore) *CertificateService {
	return &CertificateService{
		store: store,
	}
}

func (s *CertificateService) GenerateRootCertificate(ctx *appctx.AppContext) (*models.RootCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")

	rootCA := NewX509RootCertificate(ctx, "root", models.CertificateConfig{
		CommonName:         "Locally Root CA",
		Country:            "UK",
		State:              "London",
		City:               "London",
		Organization:       "Locally",
		OrganizationalUnit: "Locally",
		ExpiresInYears:     10,
		KeySize:            types.Key2048,
		SignatureAlgorithm: types.SHA512,
	})

	certificate, certDiag := rootCA.Generate(ctx)
	if certDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error generating root certificate", certDiag.Errors)
		diag.Append(certDiag)
		return nil, certDiag
	}

	ctx.Log().Debug("Generated Root Certificate")
	ctx.Log().Debugf("\nPem: %s", certificate.PemCertificate)
	ctx.Log().Debugf("\nPrivate Key: %s", certificate.PemPrivateKey)
	ctx.Log().Debugf("\nCSR: %s", certificate.Csr)
	ctx.Log().Debugf("\nConfig: %v", certificate.Config)

	return certificate, diag
}

func (s *CertificateService) GenerateIntermediateCertificate(ctx *appctx.AppContext, rootCA *models.RootCertificate) (*models.IntermediateCertificate, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_certificate")

	intermediateCA := NewX509IntermediateCertificate(ctx, "intermediate", models.CertificateConfig{
		CommonName:         "Locally Intermediate CA",
		Country:            "UK",
		State:              "London",
		City:               "London",
		Organization:       "Locally",
		OrganizationalUnit: "Locally",
		ExpiresInYears:     10,
		KeySize:            types.Key2048,
		SignatureAlgorithm: types.SHA512,
	})

	rootCertificate := X509RootCertificate{}
	rootCertificate.Parse(ctx, rootCA.PemCertificate, rootCA.PemPrivateKey)

	certificate, certDiag := intermediateCA.Generate(ctx, &rootCertificate)

	if certDiag.HasErrors() {
		ctx.Log().WithField("component", CertificateComponent).Error("Error generating intermediate certificate", certDiag.Errors)
		diag.Append(certDiag)
		return nil, certDiag
	}

	return certificate, diag
}
