// Package service provides the service for the certificates.
package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"

	"software.sslmate.com/src/go-pkcs12"

	activity_types "github.com/cjlapao/locally-cli/internal/activity/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
)

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

func createInfoActivityRecord(ctx *appctx.AppContext, message string, metadata map[string]interface{}) *activity_types.ActivityRecord {
	return createActivityRecord(ctx, activity_types.ActivityTypeInfo, activity_types.ActivityLevelInfo, message, metadata)
}

func createWarningActivityRecord(ctx *appctx.AppContext, message string, metadata map[string]interface{}) *activity_types.ActivityRecord {
	return createActivityRecord(ctx, activity_types.ActivityTypeWarning, activity_types.ActivityLevelWarning, message, metadata)
}

func createErrorActivityRecord(ctx *appctx.AppContext, message string, metadata map[string]interface{}) *activity_types.ActivityRecord {
	return createActivityRecord(ctx, activity_types.ActivityTypeError, activity_types.ActivityLevelError, message, metadata)
}

func createActivityRecord(ctx *appctx.AppContext, activityType activity_types.ActivityType, activityLevel activity_types.ActivityLevel, message string, metadata map[string]interface{}) *activity_types.ActivityRecord {
	return &activity_types.ActivityRecord{
		TenantID:      ctx.GetTenantID(),
		ActorID:       ctx.GetUserID(),
		ActorName:     ctx.GetUsername(),
		Module:        CertificateModuleKey,
		Message:       message,
		Service:       "certificates",
		Success:       true,
		ActorType:     activity_types.ActorTypeUser,
		ActivityType:  activityType,
		ActivityLevel: activityLevel,
		Data: &activity_types.ActivityData{
			Metadata: metadata,
		},
	}
}
